package queues

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

// QueueManager manges all queue operations such as Create/Update/Delete, etc
// It also is a shared resource that can be used by any other structs needing to find a Queue
// and thier associated readers
//
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

	// Find a particular queue
	//
	// ARGS:
	// * logger - standard zap logger
	// * queue - name of the queue to find
	//
	// RETURNS:
	// * Queue - queue if the queue is defined
	// * v1.Error - any errors encountered when finding the queue, or an error that it does not exist
	Find(logger *zap.Logger, queue string) (Queue, *v1.Error)

	// GetItem retrieves the next item in the queue
	//GetItem(logger *zap.Logger, ctx context.Context, ready *v1.Ready) (*v1.DequeueItemResponse, *v1.Error)

	Metrics(matchQuery *v1.MatchQuery) *v1.Metrics
}

type manager struct {
	// general constructor to create any type of queue
	queueConstructor QueueConstructor

	// all queues. Each data type save in here is of interface type ManagedQueue
	queues datastructures.BTree

	// task manger ensures shutdown requests are processsed properly
	taskManager goasync.TaskManager
}

func NewManager(queueConstructor QueueConstructor) (*manager, error) {
	btree, err := datastructures.NewBTree(2)
	if err != nil {
		return nil, err
	}

	return &manager{
		queueConstructor: queueConstructor,
		queues:           btree,
		taskManager:      goasync.NewTaskManager(goasync.RelaxedConfig()),
	}, nil
}

func (m *manager) Initialize() error { return nil }
func (m *manager) Cleanup() error    { return nil }

func (m *manager) Execute(ctx context.Context) error {
	errors := m.taskManager.Run(ctx)
	if errors != nil {
		err, _ := json.Marshal(errors)
		return fmt.Errorf("queue manager shutdown errors: %v", err)
	}

	return nil
}

func (m *manager) Create(logger *zap.Logger, create *v1.Create) *v1.Error {
	logger = logger.Named("Create")
	_, err := m.queues.FindOrCreate(datastructures.NewStringTreeKey(create.Name), "", m.create(logger, create))
	return err
}

func (m *manager) create(logger *zap.Logger, create *v1.Create) func() (any, error) {
	return func() (any, error) {
		queue, err := m.queueConstructor.NewQueue(create)
		if err != nil {
			logger.Error("failed creating queue", zap.Error(err))
			return nil, err
		}

		// if there is an error, we are shutting down so thats fine
		_ = m.taskManager.AddRunningTask(create.Name, queue)

		// return the new queue
		return queue, nil
	}
}

func (m *manager) Find(logger *zap.Logger, queue string) (Queue, *v1.Error) {
	queue := m.queues.Find(datastructures.NewStringTreeKey(queue), "")
	if queue == nil {
		return nil, errors.QueueNotFound
	}

	return queue.(Queue)
}

func (m *manager) Metrics(matchQuery *v1.MatchQuery) *v1.Metrics {
	return nil
}
