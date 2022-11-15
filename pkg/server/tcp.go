package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/DanLavine/gomultiplex"
	"github.com/DanLavine/gomultiplex/multiplexerrors"
	"github.com/DanLavine/willow-message/protocol"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/brokers"
	"github.com/DanLavine/willow/pkg/logger"
	"go.uber.org/zap"
)

type tcp struct {
	lock   *sync.Mutex
	closed bool

	logger *zap.Logger
	port   string

	queues brokers.Queues
}

func NewTCP(logger *zap.Logger, port string, queues brokers.Queues) *tcp {
	return &tcp{
		lock:   &sync.Mutex{},
		closed: false,
		logger: logger.Named("tcp_server"),
		port:   port,
		queues: queues,
	}
}

func (t *tcp) Initialize() error { return nil }
func (t *tcp) Cleanup() error    { return nil }

// TODO - wait for connections to drain
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

		go func(pipe *gomultiplex.Pipe) {
			t.handlePipe(pipe)
		}(pipe)
	}
}

func (t *tcp) handlePipe(pipe *gomultiplex.Pipe) {
	logger := t.logger.With(zap.Uint32("pipe_id", pipe.ID()))
	logger.Info("handling new pipe")

	brokerName := ""

	for {
		header, messageBody, err := readRequest(logger, pipe)
		if err != nil {
			pipe.Close()
			return
		}

		switch header.MessageFilter() {
		case protocol.Create:
			logger.Info("received create request")

			// parse create request
			createRequst := &v1.Create{}
			if err = json.Unmarshal(messageBody, createRequst); err != nil {
				logger.Error("Failed parsing create request", zap.Error(err))
				pipe.Close()
				return
			}

			// create the new queue
			brokerName = createRequst.Name
			t.queues.CreateChannel(brokerName, createRequst.Updatable)
		case protocol.Connect:
			logger.Info("received connect request")

			// parse connect request
			connectRequest := &v1.Connect{}
			if err = json.Unmarshal(messageBody, connectRequest); err != nil {
				logger.Error("Failed parsing create request", zap.Error(err))
				pipe.Close()
				return
			}

			// create the new queue
			brokerName = connectRequest.Name
			if err = t.queues.ConnectChannel(brokerName, pipe); err != nil {
				pipe.Close()
				return
			}
		case protocol.Message:
			logger.Info("received message")

			if brokerName == "" {
				logger.Error("Received a message, but client is not subscribed or producing")
				pipe.Close()
				return
			}

			// parse connect request
			messageRequest := &v1.Message{}
			if err = json.Unmarshal(messageBody, messageRequest); err != nil {
				logger.Error("Failed parsing create request", zap.Error(err))
				pipe.Close()
				return
			}

			// enqueue the message
			if err = t.queues.AddMessage(brokerName, messageRequest.Data); err != nil {
				pipe.Close()
				return
			}
		default:
			logger.Error("unkown header message filter received", zap.Any("filter", header.MessageFilter()))
			pipe.Close()
			return
		}

		// TODO. Actually use the body somehow
	}
}

// readBody will log and close the pipe for us if we fail to read the contents of the body
func readRequest(logger *zap.Logger, pipe *gomultiplex.Pipe) (protocol.Header, []byte, error) {
	// read the original header
	header := protocol.NewHeader()
	if _, err := pipe.Read(header); err != nil {
		logger.Error("failed reading header", zap.Error(err))
		return nil, nil, err
	}

	if err := header.Validate(); err != nil {
		logger.Error("failed validating header", zap.Error(err))
		return header, nil, err
	}

	// read the message body if there is one
	messageBody, err := readBody(logger, pipe, header.BodySize())
	if err != nil {
		return header, nil, err
	}

	return header, messageBody, nil
}

func readBody(logger *zap.Logger, pipe *gomultiplex.Pipe, size int64) ([]byte, error) {
	body := make([]byte, size)
	if _, err := pipe.Read(body); err != nil {
		logger.Error("failed reading message body", zap.Error(err))
		pipe.Close()
	}

	return body, nil
}

//func (t *tcp) handleConn(conn net.Conn) {
//	decoder := json.NewDecoder(conn)
//	connectionRequest := &models.ConnectRequest{}
//
//	if err := decoder.Decode(connectionRequest); err != nil {
//		t.logger.Error("failed to decode connection request", zap.Error(err))
//
//		conn.Write([]byte(`failed to decode request`))
//		conn.Close()
//
//		return
//	}
//
//}
