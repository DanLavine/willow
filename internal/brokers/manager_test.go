package brokers

import (
	"testing"

	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/internal/brokers/queues/queuesfakes"
	"github.com/DanLavine/willow/pkg/config"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
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
		fakeQueueConstructor := &queuesfakes.FakeQueueConstructor{}
		fakeQueueConstructor.NewQueueReturns(nil, &v1.Error{Message: "Failed to create queue"})

		// queue manager
		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		err := queueManager.Create(logger, &v1.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Failed to create queue"))

		g.Expect(testLogs.Len()).To(Equal(1))
		g.Expect(testLogs.All()[0].Message).To(ContainSubstring("failed creating queue"))
	})

	t.Run("it only creates a queue's tag group if it does not exist", func(t *testing.T) {
		// fake constructor to return error
		fakeQueueConstructor := &queuesfakes.FakeQueueConstructor{}
		fakeQueueConstructor.NewQueueReturns(&queuesfakes.FakeManagedQueue{}, nil)

		// queue manager
		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		err := queueManager.Create(zap.NewNop(), &v1.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})
		g.Expect(err).ToNot(HaveOccurred())
		err = queueManager.Create(zap.NewNop(), &v1.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(fakeQueueConstructor.NewQueueCallCount()).To(Equal(1))
	})
}

func TestQueueManager_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the queue has not been created", func(t *testing.T) {
		fakeQueueConstructor := &queuesfakes.FakeQueueConstructor{}
		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		queue, err := queueManager.Find(zap.NewNop(), "no queue")
		g.Expect(queue).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Queue not found"))
	})

	t.Run("it returns the proper queue", func(t *testing.T) {
		fakeQueue := &queuesfakes.FakeManagedQueue{}

		fakeQueueConstructor := &queuesfakes.FakeQueueConstructor{}
		fakeQueueConstructor.NewQueueReturnsOnCall(0, fakeQueue, nil) // check that this is returned on the proper find
		fakeQueueConstructor.NewQueueReturns(&queuesfakes.FakeManagedQueue{}, nil)

		queueManager := NewBrokerManager(fakeQueueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		g.Expect(queueManager.Create(zap.NewNop(), &v1.Create{Name: "test1", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1.Create{Name: "test2", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1.Create{Name: "test3", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1.Create{Name: "test4", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(zap.NewNop(), &v1.Create{Name: "test5", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())

		queue, err := queueManager.Find(zap.NewNop(), "test1")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(queue).To(BeIdenticalTo(fakeQueue))

		// check the another queue isn't the same
		queue, err = queueManager.Find(zap.NewNop(), "test2")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(queue).ToNot(BeIdenticalTo(fakeQueue))
	})
}

func TestQueueManager_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns metrics for all queues", func(t *testing.T) {
		queueConstructor := queues.NewQueueConstructor(&config.Config{StorageConfig: &config.StorageConfig{Type: config.MemoryStorage}})
		queueManager := NewBrokerManager(queueConstructor)
		g.Expect(queueManager).ToNot(BeNil())

		logger := zap.NewNop()

		// create 3 queues
		g.Expect(queueManager.Create(logger, &v1.Create{Name: "test1", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(logger, &v1.Create{Name: "test2", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())
		g.Expect(queueManager.Create(logger, &v1.Create{Name: "test3", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0})).ToNot(HaveOccurred())

		// check defaults
		metrics := queueManager.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(len(metrics.Queues)).To(Equal(3))
		g.Expect(metrics.Queues).To(ContainElement(&v1.QueueMetricsResponse{Name: "test1", Total: 0, Max: 5, DeadLetterQueueMetrics: nil}))
		g.Expect(metrics.Queues).To(ContainElement(&v1.QueueMetricsResponse{Name: "test2", Total: 0, Max: 5, DeadLetterQueueMetrics: nil}))
		g.Expect(metrics.Queues).To(ContainElement(&v1.QueueMetricsResponse{Name: "test3", Total: 0, Max: 5, DeadLetterQueueMetrics: nil}))

		// publish a few messages
		test1, err := queueManager.Find(logger, "test1")
		g.Expect(err).ToNot(HaveOccurred())
		test2, err := queueManager.Find(logger, "test2")
		g.Expect(err).ToNot(HaveOccurred())
		test3, err := queueManager.Find(logger, "test3")
		g.Expect(err).ToNot(HaveOccurred())

		test1.Enqueue(logger, &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test1", Tags: datatypes.StringMap{"": ""}}, Data: []byte(`hello`), Updateable: false})
		test1.Enqueue(logger, &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test1", Tags: datatypes.StringMap{"other": "other"}}, Data: []byte(`hello`), Updateable: false})

		test2.Enqueue(logger, &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test2", Tags: datatypes.StringMap{"": ""}}, Data: []byte(`hello`), Updateable: false})

		test3.Enqueue(logger, &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test3", Tags: datatypes.StringMap{"one": "one"}}, Data: []byte(`hello1`), Updateable: false})
		test3.Enqueue(logger, &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test3", Tags: datatypes.StringMap{"one": "one"}}, Data: []byte(`hello2`), Updateable: false})
		test3.Enqueue(logger, &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test3", Tags: datatypes.StringMap{"one": "one", "two": "two"}}, Data: []byte(`hello3`), Updateable: false})

		// check the new ready counts
		metrics = queueManager.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(len(metrics.Queues)).To(Equal(3))
		g.Expect(metrics.Queues).To(ContainElement(&v1.QueueMetricsResponse{Name: "test1", Total: 2, Max: 5, Tags: []*v1.TagMetricsResponse{{Tags: datatypes.StringMap{"": ""}, Ready: 1, Processing: 0}, {Tags: datatypes.StringMap{"other": "other"}, Ready: 1, Processing: 0}}, DeadLetterQueueMetrics: nil}))
		g.Expect(metrics.Queues).To(ContainElement(&v1.QueueMetricsResponse{Name: "test2", Total: 1, Max: 5, Tags: []*v1.TagMetricsResponse{{Tags: datatypes.StringMap{"": ""}, Ready: 1, Processing: 0}}, DeadLetterQueueMetrics: nil}))
		g.Expect(metrics.Queues).To(ContainElement(&v1.QueueMetricsResponse{Name: "test3", Total: 3, Max: 5, Tags: []*v1.TagMetricsResponse{{Tags: datatypes.StringMap{"one": "one"}, Ready: 2, Processing: 0}, {Tags: datatypes.StringMap{"one": "one", "two": "two"}, Ready: 1, Processing: 0}}, DeadLetterQueueMetrics: nil}))
	})
}
