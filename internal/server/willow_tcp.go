package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/internal/server/client"
	"github.com/DanLavine/willow/internal/server/v1server"
	"github.com/DanLavine/willow/pkg/config"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type willowTCP struct {
	closed bool

	logger *zap.Logger
	config *config.WillowConfig

	queueManager queues.QueueManager
	queueHandler v1server.QueueHandler
}

func NewWillowTCP(logger *zap.Logger, config *config.WillowConfig, queueManager queues.QueueManager, queueHandler v1server.QueueHandler) *willowTCP {
	return &willowTCP{
		closed:       false,
		logger:       logger.Named("willowTCP_server"),
		config:       config,
		queueManager: queueManager,
		queueHandler: queueHandler,
	}
}

func (t *willowTCP) Initialize() error { return nil }
func (t *willowTCP) Cleanup() error    { return nil }
func (t *willowTCP) Execute(ctx context.Context) error {
	logger := t.logger

	// capture any errors from the server
	errChan := make(chan error, 1)
	defer close(errChan)

	mux := http.NewServeMux()

	// broke function functions
	mux.HandleFunc("/v1/brokers/queues/create", t.queueHandler.Create)

	// message handlers
	mux.HandleFunc("/v1/brokers/item/enqueue", t.queueHandler.Enqueue)
	mux.HandleFunc("/v1/brokers/item/dequeue", t.queueHandler.Dequeue)
	mux.HandleFunc("/v1/brokers/item/ack", t.queueHandler.ACK)

	server := http2.Server{}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", *t.config.WillowPort))
	if err != nil {
		return err
	}

	go func() {
		logger.Info("TCP server running")

		for {
			conn, err := listener.Accept()
			if err != nil {
				errChan <- err
				return
			}

			go func(conn net.Conn) {
				clientTracker := client.NewTracker()
				defer clientTracker.Disconnect(logger, conn, t.queueManager)

				server.ServeConn(conn, &http2.ServeConnOpts{
					Context: context.WithValue(ctx, "clientTracker", clientTracker),
					Handler: mux,
				})

				logger.Debug("conn dissconnected")
			}(conn)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			listener.Close()
		case err := <-errChan:
			select {
			case <-ctx.Done():
				logger.Info("shutdown successfully")
				return nil
			default:
				logger.Error("received an error", zap.Error(err))
				return err
			}
		}
	}
}
