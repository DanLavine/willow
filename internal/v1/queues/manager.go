package queues

import (
	"context"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

type manager struct {
	// general constructor to create any type of queue
	queueConstructor QueueConstructor

	// all queues. Each data type save in here is of interface type ManagedQueue
	queues datastructures.BTree

	// task manger ensures shutdown requests are processsed properly
	taskManager goasync.AsyncTaskManager
}

func NewManager(queueConstructor QueueConstructor) *manager {
	btree, err := datastructures.NewBTree(2)
	if err != nil {
		panic(err)
	}

	return &manager{
		queueConstructor: queueConstructor,
		queues:           btree,
		taskManager:      goasync.NewTaskManager(goasync.RelaxedConfig()),
	}
}

// nothing to do here
func (m *manager) Initialize() error { return nil }

// nothing to do here
func (m *manager) Cleanup() error { return nil }

func (m *manager) Execute(ctx context.Context) error {
	_ = m.taskManager.Run(ctx)
	return nil
}

func (m *manager) Create(logger *zap.Logger, create *v1.Create) *v1.Error {
	logger = logger.Named("Create")
	_, err := m.queues.FindOrCreate(create.Name, "", m.create(logger, create))
	if err != nil {
		return err.(*v1.Error)
	}

	return nil
}

func (m *manager) create(logger *zap.Logger, create *v1.Create) func() (any, error) {
	return func() (any, error) {
		queue, err := m.queueConstructor.NewQueue(create)
		if err != nil {
			logger.Error("failed creating queue", zap.Error(err))
			return nil, err
		}

		// if there is an error, we are shutting down so thats fine
		_ = m.taskManager.AddExecuteTask(create.Name.ToString(), queue)

		// return the new queue
		return queue, nil
	}
}

func (m *manager) Find(logger *zap.Logger, queueName v1.String) (Queue, *v1.Error) {
	logger = logger.Named("Find")
	queue := m.queues.Find(queueName, "")
	if queue == nil {
		logger.Error("failed to find queue", zap.String("name", queueName.ToString()))
		return nil, errors.QueueNotFound
	}

	return queue.(Queue), nil
}

// Metrics is currently a simple call that will report all metrics for all queues
func (m *manager) Metrics() *v1.Metrics {
	metrics := &v1.Metrics{}

	iterator := func(queue any) {
		managedQueue := queue.(ManagedQueue)
		metrics.Queues = append(metrics.Queues, managedQueue.Metrics())
	}

	m.queues.Iterate(iterator)

	return metrics
}
