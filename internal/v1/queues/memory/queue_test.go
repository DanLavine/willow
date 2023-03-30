package memory

import (
	"testing"

	"github.com/DanLavine/willow/internal/errors"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var (
	createParams             = &v1.Create{Name: v1.String("test"), QueueMaxSize: 5, ItemRetryAttempts: 2, DeadLetterQueueMaxSize: 0}
	enqueueItemUpdateable    = &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: v1.String("test"), BrokerType: v1.Queue, Tags: v1.Tags{"updateable"}}, Data: []byte(`squash me!`), Updateable: true}
	enqueueItemNotUpdateable = &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: v1.String("test"), BrokerType: v1.Queue, Tags: v1.Tags{"not updateable"}}, Data: []byte(`hello world`), Updateable: false}
)

func TestMemoryQueue_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns the default queue properties", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal(v1.String("test")))
		g.Expect(metrics.Total).To(Equal(uint64(0)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Tags).To(BeNil())
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
	})
}

func TestMemoryQueue_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it records that an item is ready", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemUpdateable)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal(v1.String("test")))
		g.Expect(metrics.Total).To(Equal(uint64(1)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(v1.Tags{"updateable"}))
	})

	t.Run("it can enqueu multiple messages", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal(v1.String("test")))
		g.Expect(metrics.Total).To(Equal(uint64(4)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(4)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(v1.Tags{"not updateable"}))
	})

	t.Run("it can squash multiple messages", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemUpdateable)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal(v1.String("test")))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Total).To(Equal(uint64(1)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(v1.Tags{"updateable"}))
	})

	t.Run("it returns an error if the item cannot be enqueued because there are to many messages waiting to process", func(t *testing.T) {
		queue := NewQueue(&v1.Create{QueueMaxSize: 1})
		defer queue.Stop()

		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())

		err := queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(Equal(errors.MaxEnqueuedItems))
	})
}
