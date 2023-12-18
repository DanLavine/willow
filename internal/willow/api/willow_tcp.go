package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type willowTCP struct {
	closed bool

	logger *zap.Logger
	config *config.WillowConfig
	server *http.Server

	mux *urlrouter.Router
}

func NewWillowTCP(logger *zap.Logger, config *config.WillowConfig, mux *urlrouter.Router) *willowTCP {
	return &willowTCP{
		closed: false,
		logger: logger.Named("willowTCP_server"),
		config: config,
		mux:    mux,
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

	willow.server.Handler = willow.mux
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
