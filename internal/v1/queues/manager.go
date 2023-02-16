package queues

import (
	"context"
	"sync"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

//counterfeiter:generate . Manager
type Manager interface {
	// Create a new queue. performs a no-op if the queue already exists
	//
	// ARGS:
	// * logger - standard zap logger
	// * create - parameters to create a queue
	//
	// RETURNs:
	// * v1.Error - any errors encountered when creating the queue
	Create(logger *zap.Logger, create *v1.Create) *v1.Error

	// Enqueue an item to the desired queue
	//
	// ARGS:
	// * logger - standard zap logger
	// * item - data requested from the producer clients to enque
	//
	// RETURNS:
	// * v1.Error - any errors encountered when enquing an item
	Enqueue(logger *zap.Logger, item *v1.EnqueItem) *v1.Error

	// GetItem retrieves the next item in the queue
	GetItem(logger *zap.Logger, ctx context.Context, ready *v1.Ready) (*v1.DequeueItem, *v1.Error)
}

type manager struct {
	lock *sync.RWMutex

	// general constructor to create any type of queue
	queueConstructor QueueConstructor

	// TODO. this type needs to change for generic queues. Maps don't delete fully on a all to delete(...)
	// but for now this is fine.
	queues map[string]Queue
}

func NewManager(queueConstructor QueueConstructor) *manager {
	return &manager{
		lock:             new(sync.RWMutex),
		queueConstructor: queueConstructor,
		queues:           map[string]Queue{},
	}
}

func (m *manager) Create(logger *zap.Logger, create *v1.Create) *v1.Error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// if queue already exists, bail early
	if _, ok := m.queues[create.Name]; ok {
		return nil
	}

	// create new queue
	queue, err := m.queueConstructor.NewQueue(create)
	if err != nil {
		return err
	}

	m.queues[create.Name] = queue
	return nil
}
