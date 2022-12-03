package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/DanLavine/gomultiplex"
	"github.com/DanLavine/gomultiplex/multiplexerrors"
	"github.com/DanLavine/willow/pkg/brokers"
	"github.com/DanLavine/willow/pkg/logger"
	"go.uber.org/zap"
)

type tcp struct {
	lock   *sync.Mutex
	closed bool

	logger *zap.Logger
	port   string

	brokerManager brokers.BrokerManager
}

func NewTCP(logger *zap.Logger, port string, brokerManager brokers.BrokerManager) *tcp {
	return &tcp{
		lock:          &sync.Mutex{},
		closed:        false,
		logger:        logger.Named("tcp_server"),
		port:          port,
		brokerManager: brokerManager,
	}
}

func (t *tcp) Initialize() error { return nil }
func (t *tcp) Cleanup() error    { return nil }

// NOTE This will be nil on a graceful shutdown
func (t *tcp) Execute(ctx context.Context) error {
	// configure the server witth a logger
	serverConfig := gomultiplex.NewDevConfig()
	serverConfig.Logger = logger.NewLogger(t.logger)
	server, err := gomultiplex.NewServer(ctx, serverConfig, "tcp", fmt.Sprintf("localhost:%s", t.port))
	if err != nil {
		return err
	}

	for {
		// Accept all new connections
		pipe, err := server.Accept()
		if err != nil {
			if err == multiplexerrors.ServerShutdown {
				// server was told to shut down, so this is the clean case.
				t.logger.Info("clean shutdown")
				return nil
			}

			// something else happened for some reason. Return the error
			t.logger.Error("received unexpected error. Shutting down", zap.Error(err))
			return err
		}

		// this doesn't need to be in a goroutine as the broker manager handles those. But, maybe fine?
		go func(pipe *gomultiplex.Pipe) {
			t.brokerManager.HandleConnection(t.logger, pipe)
		}(pipe)
	}
}
