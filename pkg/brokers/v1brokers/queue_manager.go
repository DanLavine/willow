package v1brokers

import (
	"context"
	"fmt"
	"sync"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/gomultiplex"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/brokers/v1brokers/v1queues"
	"github.com/DanLavine/willow/pkg/multiplex"
	"go.uber.org/zap"
)

type QueueManager interface {
	CreateChannel(logger *zap.Logger, createRequest *v1.Create, pipe *gomultiplex.Pipe)
	ConnectChannel(logger *zap.Logger, connectRequest *v1.Connect, pipe *gomultiplex.Pipe)
}

type queueManager struct {
	taskManager goasync.TaskManager

	channelLock *sync.Mutex
	channels    map[string]*v1queues.Queue
}

func NewQueueManager() *queueManager {
	return &queueManager{
		taskManager: goasync.NewTaskManager(goasync.RelaxedConfig()),

		channelLock: new(sync.Mutex),
		channels:    map[string]*v1queues.Queue{},
	}
}

func (qm *queueManager) Initialize() error { return nil }
func (qm *queueManager) Cleanup() error    { return nil }
func (qm *queueManager) Execute(ctx context.Context) error {
	// This task manager won't ever report an error
	// Any errors will be written directly to the clients
	// and the threads will be closed. This just ensures that everything
	// is setup properly and all added connections will drain
	_ = qm.taskManager.Run(ctx)

	return nil
}

// Create the Channel, or attach to a channel that already exists. This sets up a constant reader
// for the pipe which will add messages to the queue.
func (qm *queueManager) CreateChannel(logger *zap.Logger, createRequest *v1.Create, pipe *gomultiplex.Pipe) {
	logger = logger.Named("CreateChannel")

	qm.channelLock.Lock()
	defer qm.channelLock.Unlock()

	// setup a new channel
	if queue, ok := qm.channels[createRequest.Name]; ok {
		// queue already exists add a new producer
		if err := qm.taskManager.AddRunningTask(fmt.Sprintf("pipe producer %d", pipe.ID()), v1queues.NewProducerConn(logger, pipe, queue)); err != nil {
			// TODO send GO_AWAY since we are shutting down and they need to reconnect
			multiplex.WriteError(logger, pipe, 502, "server is shutting down")
		}
	} else {
		// create the queue
		newQueue := v1queues.NewQueue(createRequest.Updatable)
		qm.channels[createRequest.Name] = newQueue
		if err := qm.taskManager.AddRunningTask(fmt.Sprintf("pipe producer %d", pipe.ID()), v1queues.NewProducerConn(logger, pipe, queue)); err != nil {
			// TODO send GO_AWAY since we are shutting down and they need to reconnect
			multiplex.WriteError(logger, pipe, 502, "server is shutting down")
		}
	}
}

func (qm *queueManager) ConnectChannel(logger *zap.Logger, connectRequest *v1.Connect, pipe *gomultiplex.Pipe) {
	logger = logger.Named("ConnectChannel")

	qm.channelLock.Lock()
	defer qm.channelLock.Unlock()

	if queue, ok := qm.channels[connectRequest.Name]; ok {
		// channel exists so setup the reader
		if err := qm.taskManager.AddRunningTask(fmt.Sprintf("pipe consumer %d", pipe.ID()), v1queues.NewConsumerConn(logger, pipe, queue)); err != nil {
			// TODO send GO_AWAY since we are shutting down and they need to reconnect
			multiplex.WriteError(logger, pipe, 502, "server is shutting down")
		}
	} else {
		// channel does not exists, report error back to the client and close the pipe
		logger.Error("channel does not exist", zap.String("channel", connectRequest.Name))
		if err := qm.taskManager.AddRunningTask(fmt.Sprintf("pipe consumer %d", pipe.ID()), v1queues.NewConsumerConn(logger, pipe, queue)); err != nil {
			// TODO send GO_AWAY since we are shutting down and they need to reconnect
			multiplex.WriteError(logger, pipe, 502, "server is shutting down")
		}
	}
}
