package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/DanLavine/willow/internal/server/v1server"
	"github.com/DanLavine/willow/pkg/config"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type tcp struct {
	lock   *sync.Mutex
	closed bool

	logger *zap.Logger
	config *config.Config

	queueHandler v1server.QueueHandler
}

func NewTCP(logger *zap.Logger, config *config.Config, queueHandler v1server.QueueHandler) *tcp {
	return &tcp{
		lock:         &sync.Mutex{},
		closed:       false,
		logger:       logger.Named("tcp_server"),
		config:       config,
		queueHandler: queueHandler,
	}
}

func (t *tcp) Initialize() error { return nil }
func (t *tcp) Cleanup() error    { return nil }
func (t *tcp) Execute(ctx context.Context) error {
	logger := t.logger.Named("tcp_server")

	// capture any errors from the server
	errChan := make(chan error, 1)
	defer close(errChan)

	mux := http.NewServeMux()

	// broke function functions
	mux.HandleFunc("/v1/brokers/queues/create", t.queueHandler.Create)

	// message handlers
	mux.HandleFunc("/v1/brokers/queues/message/enqueu", t.queueHandler.Enqueue)
	mux.HandleFunc("/v1/brokers/queues/message/dequeue", t.queueHandler.Dequeue)
	mux.HandleFunc("/v1/brokers/queues/message/ack", t.queueHandler.ACK)

	server := http2.Server{}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", t.config.WillowPort))
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
				server.ServeConn(conn, &http2.ServeConnOpts{
					Context: ctx,
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