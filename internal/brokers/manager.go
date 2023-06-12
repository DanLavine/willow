package brokers

import (
	"context"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

type brokerManager struct {
	// general constructor to create any type of queue
	queueConstructor queues.QueueConstructor

	// all queues. Each data type save in here is of interface type ManagedQueue
	queues btree.BTree

	// task manger ensures shutdown requests are processsed properly
	taskManager goasync.AsyncTaskManager
}

func NewBrokerManager(queueConstructor queues.QueueConstructor) *brokerManager {
	btree, err := btree.New(2)
	if err != nil {
		panic(err)
	}

	return &brokerManager{
		queueConstructor: queueConstructor,
		queues:           btree,
		taskManager:      goasync.NewTaskManager(goasync.RelaxedConfig()),
	}
}

// nothing to do here
func (bm *brokerManager) Initialize() error { return nil }

// nothing to do here
func (bm *brokerManager) Cleanup() error { return nil }

func (bm *brokerManager) Execute(ctx context.Context) error {
	_ = bm.taskManager.Run(ctx)
	return nil
}

func (bm *brokerManager) Create(logger *zap.Logger, createRequest *v1.Create) *v1.Error {
	logger = logger.Named("Create")
	_, err := bm.queues.CreateOrFind(createRequest.Name, nil, bm.create(logger, createRequest))
	if err != nil {
		return err.(*v1.Error)
	}

	return nil
}

func (bm *brokerManager) create(logger *zap.Logger, create *v1.Create) func() (any, error) {
	return func() (any, error) {
		queue, err := bm.queueConstructor.NewQueue(create)
		if err != nil {
			logger.Error("failed creating queue", zap.Error(err))
			return nil, err
		}

		// if there is an error, we are shutting down so thats fine
		_ = bm.taskManager.AddExecuteTask(create.Name.ToString(), queue)

		// return the new queue
		return queue, nil
	}
}

func (bm *brokerManager) Find(logger *zap.Logger, queueName datatypes.String) (queues.Queue, *v1.Error) {
	logger = logger.Named("Find")

	queue := bm.queues.Find(queueName, nil)
	if queue == nil {
		logger.Error("failed to find queue", zap.String("name", queueName.ToString()))
		return nil, errors.QueueNotFound
	}

	return queue.(queues.Queue), nil
}

// Metrics is currently a simple call that will report all metrics for all queues
func (m *brokerManager) Metrics() *v1.MetricsResponse {
	metrics := &v1.MetricsResponse{}

	iterator := func(key datatypes.CompareType, value any) {
		managedQueue := value.(queues.ManagedQueue)
		metrics.Queues = append(metrics.Queues, managedQueue.Metrics())
	}

	m.queues.Iterate(iterator)

	return metrics
}
