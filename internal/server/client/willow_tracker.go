package client

import (
	"net"
	"reflect"
	"sync"

	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"go.uber.org/zap"
)

type Tracker interface {
	Add(id string, brokerInfo v1willow.BrokerInfo)
	Remove(id string, brokerInfo v1willow.BrokerInfo)

	// don't keep track of this thing on each of our queues, just pass it in on the dissconnect
	Disconnect(logger *zap.Logger, conn net.Conn, queueManager queues.QueueManager)
}

type processingID struct {
	id         string
	brokerInfo v1willow.BrokerInfo
}

// tracker is used to keep track of messages being processed by a client.
// When a client dequeues an item, if it disconnects, we need to register the item as a failure
type tracker struct {
	lock          *sync.RWMutex
	processingIDs []processingID
}

func NewTracker() *tracker {
	return &tracker{
		lock:          new(sync.RWMutex),
		processingIDs: []processingID{},
	}
}

func (t *tracker) Add(id string, brokerInfo v1willow.BrokerInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.processingIDs = append(t.processingIDs, processingID{id: id, brokerInfo: brokerInfo})
}

func (t *tracker) Remove(id string, brokerInfo v1willow.BrokerInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for index, pID := range t.processingIDs {
		if pID.id == id && reflect.DeepEqual(pID.brokerInfo, brokerInfo) {
			t.processingIDs[index] = t.processingIDs[len(t.processingIDs)-1]
			t.processingIDs = t.processingIDs[:len(t.processingIDs)-1]
			break
		}
	}
}

// fail any items still processing
func (t *tracker) Disconnect(logger *zap.Logger, conn net.Conn, queueManager queues.QueueManager) {
	t.lock.Lock()
	defer t.lock.Unlock()

	logger = logger.Named("Disconnect")
	defer logger.Info("Client disconnected", zap.String("client_address", conn.RemoteAddr().String()))

	for _, pID := range t.processingIDs {
		queue, _ := queueManager.Find(logger, pID.brokerInfo.Name)
		if queue != nil {
			if err := queue.ACK(logger, &v1willow.ACK{BrokerInfo: pID.brokerInfo, ID: pID.id, Passed: false, RequeueLocation: v1willow.RequeueNone}); err != nil {
				logger.Error("failed attempting to ack a closed client", zap.Error(err))
			} else {
				logger.Debug("successfully failed pending item", zap.Any("broker_info", pID.brokerInfo), zap.String("id", pID.id))
			}
		} else {
			logger.Error("failed to find the queue", zap.String("queue_name", pID.brokerInfo.Name))
		}
	}
}
