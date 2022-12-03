package v1queues

import (
	"context"
	"fmt"
	"io"

	"github.com/DanLavine/gomultiplex"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/multiplex"
	"go.uber.org/zap"
)

type ProducerConn struct {
	logger *zap.Logger
	pipe   *gomultiplex.Pipe

	// should be exclusivly calling reads from this. other than shutdown
	queue *Queue
}

func NewProducerConn(logger *zap.Logger, pipe *gomultiplex.Pipe, queue *Queue) *ProducerConn {
	return &ProducerConn{
		logger: logger.Named("ProducerConn"),
		pipe:   pipe,
		queue:  queue,
	}
}

func (pc *ProducerConn) Execute(ctx context.Context) error {
	for {
		// read the v1 header
		header := v1.NewHeader()
		_, err := pc.pipe.Read(header)
		if err != nil {
			select {
			case <-ctx.Done():
				// nothing to do here, we are shutting down, so meh
				return nil
			default:
				if err == io.EOF {
					// client closed the pipe
					pc.logger.Error("pipe was closed", zap.Error(err))
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

		// TODO on a clean "close" from the client this should receive a GO_AWAY header or something
		// Ensure we received a "message"
		switch messageType := header.MessageType(); messageType {
		case v1.TypeMessage:
			// read the message
			message := make([]byte, header.BodySize())
			if _, err := pc.pipe.Read(message); err != nil {
				pc.logger.Error("failed to read message", zap.Error(err))
				multiplex.WriteError(pc.logger, pc.pipe, 500, err.Error())
				return err
			}

			// place the message on the queue
			pc.queue.addMessage(message)
		default:
			err = fmt.Errorf("invalid header type received, expected message")
			pc.logger.Error(err.Error(), zap.Uint32("type", uint32(messageType)))
			multiplex.WriteError(pc.logger, pc.pipe, 400, "invalid header type received, expected message")
			return err
		}
	}
}
