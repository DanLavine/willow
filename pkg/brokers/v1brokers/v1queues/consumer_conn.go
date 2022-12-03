package v1queues

import (
	"context"

	"github.com/DanLavine/gomultiplex"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/multiplex"
	"go.uber.org/zap"
)

type ConsumerConn struct {
	logger *zap.Logger
	pipe   *gomultiplex.Pipe

	// should be exclusivly calling reads from this
	queue *Queue
}

func NewConsumerConn(logger *zap.Logger, pipe *gomultiplex.Pipe, queue *Queue) *ConsumerConn {
	return &ConsumerConn{
		logger: logger.Named("ConsumerConn"),
		pipe:   pipe,
		queue:  queue,
	}
}

func (cc *ConsumerConn) Execute(ctx context.Context) error {
	messageChan := cc.queue.messageChan()

	for {
		select {
		case <-ctx.Done():
			// just return. We shouldn't care about draining for the server side. This should be back by persistant storage
			// at some point. And on the restart, we will just start sending messages to new clients again
			//
			// TODO send a GO_AWAY to the client to let them know that the server is going down and they should try to reconnect
			return nil
		case message := <-messageChan:
			v1Message, err := v1.NewMessage(message)
			if err != nil {
				multiplex.WriteError(cc.logger, cc.pipe, 500, err.Error())
				return nil
			}

			if _, err = cc.pipe.Write(v1Message); err != nil {
				if err != nil {
					multiplex.WriteError(cc.logger, cc.pipe, 500, err.Error())
					return nil
				}
			}
		}
	}
}
