package brokers

import (
	"fmt"
	"testing"

	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/willow/brokers/queues"
	"github.com/DanLavine/willow/internal/willow/brokers/queues/queuesfakes"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	servererrors "github.com/DanLavine/willow/internal/server_errors"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestQueueManager_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if creating the queue errors", func(t *testing.T) {
		// logger
		zapCore, testLogs := observer.New(zap.DebugLevel)
		logger := zap.New(zapCore)

		// fake constructor to return error
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		fakeQueueConstructor := queuesfakes.NewMockQueueConstructor(mockController)
		fakeQueueConstructor.EXPECT().NewQueue(gomock.Any()).DoAndReturn(func(createParams *v1willow.Create) (queues.ManagedQueue, *servererrors.ApiError) {
			fmt.Println("calling new queue")
			return nil, &servererrors.ApiError{Message: "Failed to create queue"}
		}).Times(1)

		// queue manager
		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		err := queueManager.Create(logger, &v1willow.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Failed to create queue"))

		g.Expect(testLogs.Len()).To(Equal(1))
		g.Expect(testLogs.All()[0].Message).To(ContainSubstring("failed creating queue"))
	})

	t.Run("it only creates a queue's tag group if it does not exist", func(t *testing.T) {
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		fakeQueueConstructor := queuesfakes.NewMockQueueConstructor(mockController)
		fakeQueueConstructor.EXPECT().NewQueue(gomock.Any()).DoAndReturn(func(createParams *v1willow.Create) (queues.ManagedQueue, *servererrors.ApiError) {
			return queuesfakes.NewMockManagedQueue(mockController), nil
		}).Times(1)

		// queue manager
		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		err := queueManager.Create(zap.NewNop(), &v1willow.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})
		g.Expect(err).ToNot(HaveOccurred())
		err = queueManager.Create(zap.NewNop(), &v1willow.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})
		g.Expect(err).ToNot(HaveOccurred())
	})
}

func TestQueueManager_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the queue has not been created", func(t *testing.T) {
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		fakeQueueConstructor := queuesfakes.NewMockQueueConstructor(mockController)
		fakeQueueConstructor.EXPECT().NewQueue(gomock.Any()).DoAndReturn(func(createParams *v1willow.Create) (queues.ManagedQueue, *servererrors.ApiError) {
			return nil, nil
		}).Times(0)

		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		queue, err := queueManager.Find(zap.NewNop(), "no queue")
		g.Expect(queue).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Queue not found"))
	})

	t.Run("it returns the proper queue", func(t *testing.T) {
		mockController := gomock.NewController(t)
		defer mockController.Finish()
		fakeQueueConstructor := queuesfakes.NewMockQueueConstructor(mockController)
		fakeQueueConstructor.EXPECT().NewQueue(gomock.Any()).DoAndReturn(func(createParams *v1willow.Create) (queues.ManagedQueue, *servererrors.ApiError) {
			return queuesfakes.NewMockManagedQueue(mockController), nil
		}).Times(5)

		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		g.Expect(queueManager.Create(zap.NewNop(), &v1willow.Create{Name: "test1", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1willow.Create{Name: "test2", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1willow.Create{Name: "test3", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1willow.Create{Name: "test4", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1willow.Create{Name: "test5", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())

		queue1, err := queueManager.Find(zap.NewNop(), "test1")
		g.Expect(err).ToNot(HaveOccurred())

		// check the another queue isn't the same
		queue2, err := queueManager.Find(zap.NewNop(), "test2")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(queue1).ToNot(BeIdenticalTo(queue2))
	})
}

func TestQueueManager_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns metrics for all queues", func(t *testing.T) {
		queueConstructor := queues.NewQueueConstructor(&config.WillowConfig{StorageConfig: &config.StorageConfig{Type: &config.MemoryStorage}})
		queueManager := NewBrokerManager(queueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		logger := zap.NewNop()

		// create 3 queues
		g.Expect(queueManager.Create(logger, &v1willow.Create{Name: "test1", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(logger, &v1willow.Create{Name: "test2", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(logger, &v1willow.Create{Name: "test3", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())

		// check defaults
		metrics := queueManager.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(len(metrics.Queues)).To(Equal(3))
		g.Expect(metrics.Queues).To(ContainElement(&v1willow.QueueMetricsResponse{Name: "test1", Total: 0, Max: 5, DeadLetterQueueMetrics: nil}))
		g.Expect(metrics.Queues).To(ContainElement(&v1willow.QueueMetricsResponse{Name: "test2", Total: 0, Max: 5, DeadLetterQueueMetrics: nil}))
		g.Expect(metrics.Queues).To(ContainElement(&v1willow.QueueMetricsResponse{Name: "test3", Total: 0, Max: 5, DeadLetterQueueMetrics: nil}))

		// publish a few messages
		test1, err := queueManager.Find(logger, "test1")
		g.Expect(err).ToNot(HaveOccurred())
		test2, err := queueManager.Find(logger, "test2")
		g.Expect(err).ToNot(HaveOccurred())
		test3, err := queueManager.Find(logger, "test3")
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(test1.Enqueue(logger, &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test1", Tags: datatypes.KeyValues{"": datatypes.String("")}}, Data: []byte(`hello`), Updateable: false})).ToNot(HaveOccurred())
		g.Expect(test1.Enqueue(logger, &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test1", Tags: datatypes.KeyValues{"other": datatypes.String("other")}}, Data: []byte(`hello`), Updateable: false})).ToNot(HaveOccurred())

		g.Expect(test2.Enqueue(logger, &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test2", Tags: datatypes.KeyValues{"": datatypes.String("")}}, Data: []byte(`hello`), Updateable: false})).ToNot(HaveOccurred())

		g.Expect(test3.Enqueue(logger, &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test3", Tags: datatypes.KeyValues{"one": datatypes.String("one")}}, Data: []byte(`hello1`), Updateable: false})).ToNot(HaveOccurred())
		g.Expect(test3.Enqueue(logger, &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test3", Tags: datatypes.KeyValues{"one": datatypes.String("one")}}, Data: []byte(`hello2`), Updateable: false})).ToNot(HaveOccurred())
		g.Expect(test3.Enqueue(logger, &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test3", Tags: datatypes.KeyValues{"one": datatypes.String("one"), "two": datatypes.String("two")}}, Data: []byte(`hello3`), Updateable: false})).ToNot(HaveOccurred())

		// check the new ready counts
		metrics = queueManager.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(len(metrics.Queues)).To(Equal(3))

		g.Expect(metrics.Queues[0].Name).To(Equal("test2"))
		g.Expect(metrics.Queues[0].Total).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[0].Tags).To(ContainElements([]*v1willow.TagMetricsResponse{{Tags: datatypes.KeyValues{"": datatypes.String("")}, Ready: 1, Processing: 0}}))

		g.Expect(metrics.Queues[1].Name).To(Equal("test1"))
		g.Expect(metrics.Queues[1].Total).To(Equal(uint64(2)))
		g.Expect(metrics.Queues[1].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[1].Tags).To(ContainElements([]*v1willow.TagMetricsResponse{{Tags: datatypes.KeyValues{"": datatypes.String("")}, Ready: 1, Processing: 0}, {Tags: datatypes.KeyValues{"other": datatypes.String("other")}, Ready: 1, Processing: 0}}))

		g.Expect(metrics.Queues[2].Name).To(Equal("test3"))
		g.Expect(metrics.Queues[2].Total).To(Equal(uint64(3)))
		g.Expect(metrics.Queues[2].Max).To(Equal(uint64(5)))
		g.Expect(metrics.Queues[2].Tags).To(ContainElements([]*v1willow.TagMetricsResponse{{Tags: datatypes.KeyValues{"one": datatypes.String("one")}, Ready: 2, Processing: 0}, {Tags: datatypes.KeyValues{"one": datatypes.String("one"), "two": datatypes.String("two")}, Ready: 1, Processing: 0}}))
	})
}
