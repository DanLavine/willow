package v1brokers

import (
	"sync"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow-message/protocol"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/brokers/v1brokers/v1queues"
	"github.com/DanLavine/willow/pkg/errors"
	"go.uber.org/zap"
)

type QueueMetrics struct {
	ChannelMetrics map[string]ChannelMetrics
}

type ChannelMetrics struct {
	MessagesReady      int32
	MessagesProcessing int32
}

type QueueManager interface {
	// report metrics about the state of the queue
	Metrics() QueueMetrics

	// Create a new queue. If one already exists, returns nil
	CreateQueue(logger *zap.Logger, createRequest *v1.Create) *errors.Error

	// Enqueue a message on a queue that already exists
	EnqueMessage(logger *zap.Logger, messageRequest *v1.Message) *errors.Error

	// Retrieve a message off a queue that already exists. Blocks unitl a message is ready or we are shutting down
	RetrieveMessage(logger *zap.Logger, readyRequest *v1.Ready) (*v1.Message, *errors.Error)

	// Report success or failure of a message
	ACKMessage(logger *zap.Logger, ackRequest *v1.ACK) *errors.Error
}

type queueManager struct {
	taskManager goasync.TaskManager

	channelLock *sync.RWMutex
	channels    map[string]*v1queues.Queue
}

func NewQueueManager() *queueManager {
	return &queueManager{
		taskManager: goasync.NewTaskManager(goasync.RelaxedConfig()),

		channelLock: new(sync.RWMutex),
		channels:    map[string]*v1queues.Queue{},
	}
}

// Create the Channel, or attach to a channel that already exists. This sets up a constant reader
// for the pipe which will add messages to the queue.
func (qm *queueManager) CreateQueue(logger *zap.Logger, createRequest *v1.Create) *errors.Error {
	logger = logger.Named("CreateQueue").With(zap.String("queue_name", createRequest.BrokerName))
	logger.Debug("creating queue")

	if createRequest.BrokerType != protocol.Queue {
		brokerType := createRequest.BrokerType.ToString()
		logger.Error("failed validating create request, incorrect broker type", zap.String("broker_type", brokerType))
		return errors.ValidationError.Expected("broker").Actual(brokerType)
	}

	qm.channelLock.Lock()
	defer qm.channelLock.Unlock()

	if _, ok := qm.channels[createRequest.BrokerName]; !ok {
		qm.channels[createRequest.BrokerName] = v1queues.NewQueue()
	}

	logger.Debug("created queue")
	return nil
}

func (qm *queueManager) EnqueMessage(logger *zap.Logger, messageRequest *v1.Message) *errors.Error {
	logger = logger.Named("EnqueMessage").With(zap.String("queue_name", messageRequest.BrokerName))
	logger.Debug("enquing message")

	qm.channelLock.RLock()
	defer qm.channelLock.RUnlock()

	if channel, ok := qm.channels[messageRequest.BrokerName]; ok {
		channel.AddMessage(messageRequest)
	} else {
		logger.Error("failed enquing message", zap.String("error", "queue does not exist"))
		return errors.QueueNotFound.Expected(messageRequest.BrokerName)
	}

	logger.Debug("enqued message")
	return nil
}

func (qm *queueManager) RetrieveMessage(logger *zap.Logger, readyRequest *v1.Ready) (*v1.Message, *errors.Error) {
	logger = logger.Named("RetrieveMessage").With(zap.String("queue_name", readyRequest.BrokerName))
	logger.Debug("enquing message")

	qm.channelLock.RLock()
	defer qm.channelLock.RUnlock()

	if channel, ok := qm.channels[readyRequest.BrokerName]; ok {
		message := channel.RetrieveMessage()
		if message == nil {
			logger.Debug("failed to retrieve message")
			return nil, errors.RetrieveError
		}

		logger.Debug("retrieved message")
		return message, nil
	} else {
		logger.Error("failed retrieve message", zap.String("error", "queue does not exist"))
		return nil, errors.QueueNotFound.Expected(readyRequest.BrokerName)
	}
}

func (qm *queueManager) ACKMessage(logger *zap.Logger, ACKRequest *v1.ACK) *errors.Error {
	logger = logger.Named("ACKMessage").With(zap.String("queue_name", ACKRequest.BrokerName))
	logger.Debug("acking message")

	qm.channelLock.RLock()
	defer qm.channelLock.RUnlock()

	if channel, ok := qm.channels[ACKRequest.BrokerName]; ok {
		_ = channel.ACK(ACKRequest.ID, ACKRequest.Passed)
		return nil
	} else {
		logger.Error("failed acking message", zap.String("error", "queue does not exist"))
		return errors.QueueNotFound.Expected(ACKRequest.BrokerName)
	}
}

func (qm *queueManager) Metrics() QueueMetrics {
	qm.channelLock.Lock()
	defer qm.channelLock.Unlock()

	channels := map[string]ChannelMetrics{}
	for channelName, queue := range qm.channels {
		channels[channelName] = ChannelMetrics{
			MessagesReady:      queue.GetMessageReadyCount(),
			MessagesProcessing: queue.GetMessageProcessingCount(),
		}
	}

	return QueueMetrics{
		ChannelMetrics: channels,
	}
}
