package v1queues

import (
	"context"
	"sync"

	"github.com/DanLavine/gomultiplex"
	"github.com/DanLavine/gomultiplex/multiplexerrors"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/multiplex"
	"go.uber.org/zap"
)

type ConsumerConn struct {
	logger *zap.Logger
	pipe   *gomultiplex.Pipe

	closePipe   *sync.Once
	clientClose chan struct{}
	ackChan     chan struct{}
	failChan    chan struct{}

	// should be exclusivly calling writes
	queue *Queue

	// cleanup func for manager and metrics
	cleanup func()
}

func NewConsumerConn(logger *zap.Logger, pipe *gomultiplex.Pipe, queue *Queue, cleanup func()) *ConsumerConn {
	return &ConsumerConn{
		logger:      logger.Named("ConsumerConn"),
		pipe:        pipe,
		closePipe:   new(sync.Once),
		clientClose: make(chan struct{}),
		ackChan:     make(chan struct{}),
		failChan:    make(chan struct{}),
		queue:       queue,
		cleanup:     cleanup,
	}
}

func (cc *ConsumerConn) Execute(ctx context.Context) error {
	defer cc.cleanup()

	messageChan := cc.queue.messageChan()

	// Constantly read from the client to receive GO_AWAY, ACK, Fail headers
	// This exits when:
	//  1. we fail to read a header or validate the data.
	//      - write an error back to the client and close the pipe
	//  2. An unexpected message type
	//      - write an error back to the client and close the pipe
	//  3. If the main thread recieves a ctx.Done():
	//      - close the pipe immediately if there is no work going on
	//      - close the pipe after processing a in flight queue message
	//      In both cases, this should result in a multiplexerrors.PipeClosed which is a clean shutdown
	go func() {
		for {
			// read the v1 header. On any failures, or shutdowns, the pipe will be canceld, causing an error to be
			// returned eventually. if it was multiplexerrors.PipeClosed then we know that we closed the pipe here
			header := v1.NewHeader()
			_, err := cc.pipe.Read(header)
			if err != nil {
				if err != multiplexerrors.PipeClosed {
					cc.logger.Error("faild to read header", zap.Error(err))
					multiplex.WriteError(cc.logger, cc.pipe, 500, "failed to read header")
				}

				return
			}

			// validate the header
			if err := header.Validate(); err != nil {
				cc.logger.Error("faild to validate header", zap.Error(err))
				multiplex.WriteError(cc.logger, cc.pipe, 500, "failed to validate header")
				return
			}

			switch messageType := header.MessageType(); messageType {
			case v1.TypeGoAway:
				// client wants to shut down
				cc.closePipe.Do(func() {
					close(cc.clientClose)
				})
			case v1.TypeACK:
				// client should have acked the lat message they processed
				select {
				case cc.ackChan <- struct{}{}:
				default:
					cc.logger.Error("received unexpected ack")
					multiplex.WriteError(cc.logger, cc.pipe, 400, "received unexpected ack")
					return
				}
			case v1.TypeFail:
				// client should be failing the last message they are processing
				select {
				case cc.failChan <- struct{}{}:
				default:
					cc.logger.Error("received unexpected fail")
					multiplex.WriteError(cc.logger, cc.pipe, 400, "received unexpected fail")
					return
				}
			default:
				cc.logger.Error("received unexpected message type", zap.Uint32("message_type", uint32(messageType)))
				multiplex.WriteError(cc.logger, cc.pipe, 400, "received unexpected message type")
				return
			}
		}
	}()

	for {
		// 1. client shutdown check
		select {
		case <-cc.clientClose:
			// client has sent a GO_AWAY. close the pipe and return
			cc.pipe.Close()
			return nil
		default:
			// still running
		}

		// 2. server shutdown check
		select {
		case <-ctx.Done():
			// we are shutting down the server. don't care about error here since we are closing
			_, _ = cc.pipe.Write(v1.NewGoAway())
			cc.pipe.Close() // safe to call multiple times, so don't worry about wrapping in once
			return nil
		default:
			// still running
		}

		// 3. read loop received an error and already closed things down
		select {
		case <-cc.pipe.Closed():
			// error should already be handled
			return nil
		default:
			// still running
		}

		// check for shutdowns still, or a new message
		select {
		case <-cc.clientClose:
			// client has sent a GO_AWAY. close the pipe and return
			cc.pipe.Close()
			return nil
		case <-ctx.Done():
			// we are shutting down the server. don't care about error here since we are closing
			_, _ = cc.pipe.Write(v1.NewGoAway())
			cc.pipe.Close() // safe to call multiple times, so don't worry about wrapping in once
			return nil
		case <-cc.pipe.Closed():
			return nil
		case message := <-messageChan:
			// pull a message off the queue

			// write the message to the client
			v1Message, err := v1.NewMessage(message)
			if err != nil {
				// TODO mark message as failed
				multiplex.WriteError(cc.logger, cc.pipe, 500, err.Error())
				return nil
			}

			if _, err = cc.pipe.Write(v1Message); err != nil {
				if err != nil {
					// TODO mark message as failed
					multiplex.WriteError(cc.logger, cc.pipe, 500, err.Error())
					return nil
				}
			}

			// read ACK or FAIL header from client. This blocks untill client is done
			select {
			case <-cc.ackChan:
				// TODO mark message as success
				// message has successfully processed
			case <-cc.failChan:
				// TODO mark message as failed
				// message has failed to process. put it back on the queue to try again
			}
		}
	}
}
