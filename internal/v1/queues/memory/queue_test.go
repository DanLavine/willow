package memory

import (
	"testing"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

var (
	createParams             = &v1.Create{Name: "test", QueueMaxSize: 5, ItemRetryAttempts: 2, DeadLetterQueueMaxSize: nil}
	enqueueItemUpdateable    = &v1.EnqueueItem{Tags: []string{""}, Data: []byte(`squash me!`), Updateable: true}
	enqueueItemNotUpdateable = &v1.EnqueueItem{Tags: []string{""}, Data: []byte(`hello world`), Updateable: false}
)

func TestMemoryQueue_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns the default queue properties", func(t *testing.T) {
		queue := NewQueue(&v1.Create{Name: "test", QueueMaxSize: 5, ItemRetryAttempts: 2, DeadLetterQueueMaxSize: nil})
		defer queue.Stop()

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
	})
}

func TestMemoryQueue_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it records that an item is ready", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		err := queue.Enqueue(enqueueItemNotUpdateable)
		g.Expect(err).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
	})

	t.Run("it can enqueu multiple messages", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Ready).To(Equal(uint64(4)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
	})

	t.Run("it can squash multiple messages", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		g.Expect(queue.Enqueue(enqueueItemUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(enqueueItemUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(enqueueItemUpdateable)).ToNot(HaveOccurred())
		g.Expect(queue.Enqueue(enqueueItemUpdateable)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
	})
}

//func TestMemoryQueue_Enque(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	t.Run("it records an item for processing", func(t *testing.T) {
//		queue := memory.NewQueue(createParams)
//
//		err := queue.Enqueue(enqueueItemNotUpdateable)
//		g.Expect(err).ToNot(HaveOccurred())
//
//		metrics := queue.Metrics()
//		g.Expect(metrics).ToNot(BeNil())
//		g.Expect(metrics.Name).To(Equal("test"))
//		g.Expect(metrics.Ready).To(Equal(uint64(1)))
//		g.Expect(metrics.Processing).To(Equal(uint64(0)))
//		g.Expect(metrics.Max).To(Equal(uint64(5)))
//		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
//	})
//
//	t.Run("when an enqueued item is not updateable", func(t *testing.T) {
//		t.Run("it adds the items to the list for processing", func(t *testing.T) {
//			queue := memory.NewQueue(createParams)
//
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable)).ToNot(HaveOccurred())
//
//			metrics := queue.Metrics()
//			g.Expect(metrics).ToNot(BeNil())
//			g.Expect(metrics.Ready).To(Equal(uint64(4)))
//		})
//
//		t.Run("with multiple tags", func(t *testing.T) {
//			t.Run("it enqueues all items for processing", func(t *testing.T) {
//				queue := memory.NewQueue(createParams)
//
//				enqueueItemNotUpdateable1 := &v1.EnqueItem{Tag: "one", Data: []byte(`hello world`), Updateable: false}
//				enqueueItemNotUpdateable2 := &v1.EnqueItem{Tag: "two", Data: []byte(`hello world`), Updateable: false}
//				enqueueItemNotUpdateable3 := &v1.EnqueItem{Tag: "three", Data: []byte(`hello world`), Updateable: false}
//				enqueueItemNotUpdateable4 := &v1.EnqueItem{Tag: "four", Data: []byte(`hello world`), Updateable: false}
//
//				g.Expect(queue.Enqueue(enqueueItemNotUpdateable1)).ToNot(HaveOccurred())
//				g.Expect(queue.Enqueue(enqueueItemNotUpdateable2)).ToNot(HaveOccurred())
//				g.Expect(queue.Enqueue(enqueueItemNotUpdateable3)).ToNot(HaveOccurred())
//				g.Expect(queue.Enqueue(enqueueItemNotUpdateable4)).ToNot(HaveOccurred())
//
//				metrics := queue.Metrics()
//				g.Expect(metrics).ToNot(BeNil())
//				g.Expect(metrics.Ready).To(Equal(uint64(4)))
//			})
//		})
//	})
//
//	t.Run("when an enqueued item is updateable", func(t *testing.T) {
//		t.Run("it overwrites the item if it is not yet processing", func(t *testing.T) {
//			queue := memory.NewQueue(createParams)
//
//			enqueueItemNotUpdateable1 := &v1.EnqueItem{Tag: "a", Data: []byte(`one`), Updateable: false}
//			enqueueItemNotUpdateable2 := &v1.EnqueItem{Tag: "a", Data: []byte(`two`), Updateable: false}
//			enqueueItemNotUpdateable3 := &v1.EnqueItem{Tag: "a", Data: []byte(`three`), Updateable: false}
//			enqueueItemNotUpdateable4 := &v1.EnqueItem{Tag: "a", Data: []byte(`four`), Updateable: false}
//
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable1)).ToNot(HaveOccurred())
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable2)).ToNot(HaveOccurred())
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable3)).ToNot(HaveOccurred())
//			g.Expect(queue.Enqueue(enqueueItemNotUpdateable4)).ToNot(HaveOccurred())
//
//			metrics := queue.Metrics()
//			g.Expect(metrics).ToNot(BeNil())
//			g.Expect(metrics.Ready).To(Equal(uint64(1)))
//		})
//	})
//}
