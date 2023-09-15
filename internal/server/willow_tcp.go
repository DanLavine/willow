package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/server/client"
	"github.com/DanLavine/willow/internal/server/versions/v1willow"
	"github.com/DanLavine/willow/pkg/config"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type willowTCP struct {
	closed bool

	logger *zap.Logger
	config *config.WillowConfig
	server *http.Server

	connTracker  btree.BTree
	queueManager queues.QueueManager
	queueHandler v1willow.QueueHandler
}

func NewWillowTCP(logger *zap.Logger, config *config.WillowConfig, queueManager queues.QueueManager, queueHandler v1willow.QueueHandler) *willowTCP {
	connTrackerTree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &willowTCP{
		closed:       false,
		logger:       logger.Named("willowTCP_server"),
		config:       config,
		connTracker:  connTrackerTree,
		queueManager: queueManager,
		queueHandler: queueHandler,
	}
}

func (willow *willowTCP) Cleanup() error { return nil }
func (willow *willowTCP) Initialize() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", *willow.config.WillowPort),
	}

	// load the server CRT and Key
	cert, err := tls.LoadX509KeyPair(*willow.config.WillowServerCRT, *willow.config.WillowServerKey)
	if err != nil {
		return err
	}

	// add them to the server
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// load the ROOT CA cert if it exists (for self signed certs)
	if *willow.config.WillowCA != "" {
		CaPEM, err := ioutil.ReadFile(*willow.config.WillowCA)
		if err != nil {
			return err
		}

		CAs := x509.NewCertPool()
		if !CAs.AppendCertsFromPEM(CaPEM) {
			return fmt.Errorf("failed to parse WillowCA")
		}

		server.TLSConfig.RootCAs = CAs
	}

	// enforce using http2
	willow.server = server

	return nil
}

func (willow *willowTCP) Execute(willowCTX context.Context) error {
	logger := willow.logger

	// capture any errors from the server
	errChan := make(chan error, 1)
	defer close(errChan)

	mux := http.NewServeMux()

	// broke function functions
	mux.HandleFunc("/v1/brokers/queues/create", willow.queueHandler.Create)

	// message handlers
	mux.HandleFunc("/v1/brokers/item/enqueue", willow.queueHandler.Enqueue)
	mux.HandleFunc("/v1/brokers/item/dequeue", willow.queueHandler.Dequeue)
	mux.HandleFunc("/v1/brokers/item/ack", willow.queueHandler.ACK)

	// first call to server that sets up the the tracker for any requests for a conn
	willow.server.ConnContext = func(ctx context.Context, conn net.Conn) context.Context {
		clientTracker := client.NewTracker()
		if err := willow.connTracker.CreateOrFind(datatypes.String(conn.RemoteAddr().String()), func() any { return clientTracker }, func(item any) {}); err != nil {
			panic(err)
		}

		return context.WithValue(ctx, "clientTracker", clientTracker)
	}

	// setup for when a client disconnects, need to cleanup the client tracker
	willow.server.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateClosed:
			willow.connTracker.Delete(datatypes.String(conn.RemoteAddr().String()), func(item any) bool {
				clientTracker := item.(client.Tracker)
				clientTracker.Disconnect(logger, conn, willow.queueManager)

				return true
			})
		}
	}

	willow.server.Handler = mux
	http2.ConfigureServer(willow.server, &http2.Server{})

	// handle shutdown
	shutdownErr := make(chan error)
	go func() {
		<-willowCTX.Done()
		shutdownErr <- willow.server.Shutdown(context.Background())
	}()

	logger.Info("Willow TCP server running")
	if err := willow.server.ListenAndServeTLS("", ""); err != nil {
		select {
		case <-willowCTX.Done():
			if err != http.ErrServerClosed {
				// must be an unexpected error during shutdown
				return err
			}

			// context was closed and server closed error. clean shutdown case
		default:
			// always return the error if the context was not closed
			logger.Error("server shutdown unexpectedly", zap.Error(err))
			return err
		}
	}

	// wait for any pending connections to drain
	if err := <-shutdownErr; err != nil {
		logger.Error("server shutdown with error", zap.Error(err))
		return err
	}

	logger.Info("server shutdown successfully")
	return nil
}
