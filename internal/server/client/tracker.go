package client

import (
	"net"
	"reflect"
	"sync"

	"github.com/DanLavine/willow/internal/brokers/queues"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

type Tracker interface {
	Add(id uint64, brokerInfo v1.BrokerInfo)
	Remove(id uint64, brokerInfo v1.BrokerInfo)

	// don't keep track of this thing on each of our queues, just pass it in on the dissconnect
	Disconnect(logger *zap.Logger, conn net.Conn, queueManager queues.QueueManager)
}

type processingID struct {
	id         uint64
	brokerInfo v1.BrokerInfo
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

func (t *tracker) Add(id uint64, brokerInfo v1.BrokerInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.processingIDs = append(t.processingIDs, processingID{id: id, brokerInfo: brokerInfo})
}

func (t *tracker) Remove(id uint64, brokerInfo v1.BrokerInfo) {
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
			if err := queue.ACK(logger, &v1.ACK{BrokerInfo: pID.brokerInfo, ID: pID.id, Passed: false}); err != nil {
				logger.Error("failed attempting to ack a closed client", zap.Error(err))
			} else {
				logger.Debug("failed item", zap.Any("broker_info", pID.brokerInfo), zap.Uint64("id", pID.id))
			}
		}
	}
}
