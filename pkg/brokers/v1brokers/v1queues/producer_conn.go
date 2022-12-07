package v1queues

import (
	"context"
	"fmt"

	"github.com/DanLavine/gomultiplex"
	"github.com/DanLavine/gomultiplex/multiplexerrors"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/multiplex"
	"go.uber.org/zap"
)

type ProducerConn struct {
	logger *zap.Logger
	pipe   *gomultiplex.Pipe

	// should be exclusivly calling reads from this. other than shutdown
	queue *Queue

	// cleanup func for manager and metrics
	cleanup func()
}

func NewProducerConn(logger *zap.Logger, pipe *gomultiplex.Pipe, queue *Queue, cleanup func()) *ProducerConn {
	return &ProducerConn{
		logger:  logger.Named("ProducerConn"),
		pipe:    pipe,
		queue:   queue,
		cleanup: cleanup,
	}
}

func (pc *ProducerConn) Execute(ctx context.Context) error {
	defer pc.cleanup()

	for {
		// read the v1 header
		header := v1.NewHeader()
		pc.logger.Info("reading header")
		_, err := pc.pipe.Read(header)
		if err != nil {
			select {
			case <-ctx.Done():
				// nothing to do here, we are shutting down, so meh
				return nil
			default:
				if err == multiplexerrors.PipeClosed {
					// client closed the pipe
					pc.logger.Info("pipe was closed", zap.Error(err))
					return nil
				} else {
					pc.logger.Error("Failed to read header", zap.Error(err))
					multiplex.WriteError(pc.logger, pc.pipe, 500, err.Error())
				}

				return err
			}
		}

		// validate the header
		if err := header.Validate(); err != nil {
			pc.logger.Error("invalid header received", zap.Error(err))
			multiplex.WriteError(pc.logger, pc.pipe, 400, err.Error())
			return err
		}

		switch messageType := header.MessageType(); messageType {
		// Client closed the sender. So so cleanup and return here
		//case v1.TypeGoAway:
		//	pc.pipe.Close()
		//	return nil
		// Client sent a message, we need to process
		case v1.TypeMessage:
			// read the message
			message := make([]byte, header.BodySize())
			if _, err := pc.pipe.Read(message); err != nil {
				pc.logger.Error("failed to read message", zap.Error(err))
				multiplex.WriteError(pc.logger, pc.pipe, 500, err.Error())
				return err
			}

			pc.queue.AddMessage(message)
		default:
			// received an unexpected message type. Attempt to write an error back and close the pipe
			err = fmt.Errorf("invalid header type received, expected message")
			pc.logger.Error(err.Error(), zap.Uint32("type", uint32(messageType)))
			multiplex.WriteError(pc.logger, pc.pipe, 400, "invalid header type received, expected message")
			return err
		}
	}
}
