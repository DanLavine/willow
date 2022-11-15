package brokers

import (
	"fmt"
	"sync"

	"github.com/DanLavine/gomultiplex"
	"go.uber.org/zap"
)

type Queues interface {
	CreateChannel(name string, updateable bool)
	ConnectChannel(name string, pipe *gomultiplex.Pipe) error

	AddMessage(name string, message []byte) error
}

type queues struct {
	logger *zap.Logger

	channelLock *sync.Mutex
	channels    map[string]*queue
}

func NewQueues(logger *zap.Logger) *queues {
	return &queues{
		logger:      logger.Named("queues"),
		channelLock: new(sync.Mutex),
		channels:    map[string]*queue{},
	}
}

func (q *queues) CreateChannel(name string, updateable bool) {
	q.channelLock.Lock()
	defer q.channelLock.Unlock()

	if _, ok := q.channels[name]; !ok {
		q.channels[name] = newQueue(updateable)
	}
}

func (q *queues) ConnectChannel(name string, pipe *gomultiplex.Pipe) error {
	q.channelLock.Lock()
	defer q.channelLock.Unlock()

	if queue, ok := q.channels[name]; ok {
		if err := queue.addClient(pipe); err != nil {
			q.logger.Error("Failed connecting a pipe to channel", zap.Uint32("pipe_id", pipe.ID()), zap.Error(err))
			return err
		}
	} else {
		return fmt.Errorf("Pipe '%s' does not exist", name)
	}

	return nil
}

func (q *queues) AddMessage(name string, message []byte) error {
	q.channelLock.Lock()
	defer q.channelLock.Unlock()

	if queue, ok := q.channels[name]; ok {
		queue.addMessage(message)
	} else {
		return fmt.Errorf("Pipe '%s' does not exist", name)
	}

	return nil
}
