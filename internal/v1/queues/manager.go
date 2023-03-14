package queues

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

//counterfeiter:generate . QueueManager
type QueueManager interface {
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
	Enqueue(logger *zap.Logger, item *v1.EnqueueItem) *v1.Error

	// GetItem retrieves the next item in the queue
	//GetItem(logger *zap.Logger, ctx context.Context, ready *v1.Ready) (*v1.DequeueItemResponse, *v1.Error)

	Metrics(matchQuery *v1.MatchQuery) *v1.Metrics
}

type manager struct {
	lock *sync.RWMutex

	// general constructor to create any type of queue
	queueConstructor QueueConstructor

	// all queues
	queues datastructures.BTree
}

func NewManager(queueConstructor QueueConstructor) (*manager, error) {
	btree, err := datastructures.NewBTree(2)
	if err != nil {
		return nil, err
	}

	return &manager{
		lock:             new(sync.RWMutex),
		queueConstructor: queueConstructor,
		queues:           btree,
	}, nil
}

func (m *manager) Create(logger *zap.Logger, create *v1.Create) *v1.Error {
	// create new queue
	queue, err := m.queueConstructor.NewQueue(create)
	if err != nil {
		return err
	}

	// on a creation the passed in item will be returned
	foundQueue := m.queues.FindOrCreate(datastructures.NewStringTreeKey(create.Name), queue)
	if foundQueue == queue {
		return queue.Init()
	}

	return nil
}

func (m *manager) Enqueue(logger *zap.Logger, item *v1.EnqueueItem) *v1.Error {
	queue := m.queues.Find(datastructures.NewStringTreeKey(item.Name))
	if queue == nil {
		return errors.QueueNotFound
	}

	return queue.(Queue).Enqueue(item)
}

func (m *manager) Metrics(matchQuery *v1.MatchQuery) *v1.Metrics {
	return nil
}
