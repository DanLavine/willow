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

type QueueMetrics struct {
	ChannelMetrics map[string]ChannelMetrics
}

type ChannelMetrics struct {
	Messages  int32
	Producers int32
	Consumers int32
}

type QueueManager interface {
	Metrics() QueueMetrics
	Create(logger *zap.Logger, createRequest *v1.Create, pipe *gomultiplex.Pipe)
	Connect(logger *zap.Logger, connectRequest *v1.Connect, pipe *gomultiplex.Pipe)
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
func (qm *queueManager) Create(logger *zap.Logger, createRequest *v1.Create, pipe *gomultiplex.Pipe) {
	logger = logger.Named("CreateChannel")
	var queue *v1queues.Queue

	qm.channelLock.Lock()
	defer qm.channelLock.Unlock()

	// setup a new channel
	if foundQueue, ok := qm.channels[createRequest.Name]; ok {
		queue = foundQueue
	} else {
		// create the queue
		queue = v1queues.NewQueue(createRequest.Updatable)
		qm.channels[createRequest.Name] = queue
	}

	if err := qm.taskManager.AddRunningTask(fmt.Sprintf("producer %d", pipe.ID()), v1queues.NewProducerConn(logger, pipe, queue, queue.DecrementProducer)); err != nil {
		// TODO send GO_AWAY since we are shutting down and they need to reconnect
		multiplex.WriteError(logger, pipe, 502, "server is shutting down")
	} else {
		queue.AddProducer()
	}
}

func (qm *queueManager) Connect(logger *zap.Logger, connectRequest *v1.Connect, pipe *gomultiplex.Pipe) {
	logger = logger.Named("ConnectChannel")

	qm.channelLock.Lock()
	defer qm.channelLock.Unlock()

	if queue, ok := qm.channels[connectRequest.Name]; ok {
		// channel exists so setup the reader
		if err := qm.taskManager.AddRunningTask(fmt.Sprintf("consumer %d", pipe.ID()), v1queues.NewConsumerConn(logger, pipe, queue, queue.DecrementConsumer)); err != nil {
			// TODO send GO_AWAY since we are shutting down and they need to reconnect
			multiplex.WriteError(logger, pipe, 502, "server is shutting down")
		} else {
			queue.AddConsumer()
		}
	} else {
		// channel does not exists, report error back to the client and close the pipe
		logger.Error("channel does not exist", zap.String("channel", connectRequest.Name))
		multiplex.WriteError(logger, pipe, 404, "no producers for channel")
	}
}

func (qm *queueManager) decrementProducers() {

}
func (qm *queueManager) decrementConsumers() {

}

func (qm *queueManager) Metrics() QueueMetrics {
	qm.channelLock.Lock()
	defer qm.channelLock.Unlock()

	channels := map[string]ChannelMetrics{}
	for channelName, queue := range qm.channels {
		channels[channelName] = ChannelMetrics{
			Messages:  queue.GetMessageCount(),
			Producers: queue.GetProducerCount(),
			Consumers: queue.GetConsumerCount(),
		}
	}

	return QueueMetrics{
		ChannelMetrics: channels,
	}
}
