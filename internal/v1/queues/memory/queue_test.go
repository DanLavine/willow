package memory_test

import (
	"testing"

	"github.com/DanLavine/willow/internal/v1/queues/memory"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

var (
	createParams = &v1.Create{Name: "test", QueueMaxSize: 5, ItemRetryAttempts: 2, DeadLetterQueueMaxSize: nil}
	enqueueItem  = &v1.EnqueItem{Tag: "", Data: []byte(`hello world`), Updateable: false}
)

func TestMemoryQueue_Enque(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it records an item for processing", func(t *testing.T) {
		queue := memory.NewQueue(createParams)

		err := queue.Enqueue(enqueueItem)
		g.Expect(err).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
	})

	t.Run("when an enqueued item is not updateable", func(t *testing.T) {
		t.Run("it adds the items to the list for processing", func(t *testing.T) {
			queue := memory.NewQueue(createParams)

			g.Expect(queue.Enqueue(enqueueItem)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(enqueueItem)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(enqueueItem)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(enqueueItem)).ToNot(HaveOccurred())

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics.Ready).To(Equal(uint64(4)))
		})

		t.Run("with multiple tags", func(t *testing.T) {
			t.Run("it enqueues all items for processing", func(t *testing.T) {
				queue := memory.NewQueue(createParams)

				enqueueItem1 := &v1.EnqueItem{Tag: "one", Data: []byte(`hello world`), Updateable: false}
				enqueueItem2 := &v1.EnqueItem{Tag: "two", Data: []byte(`hello world`), Updateable: false}
				enqueueItem3 := &v1.EnqueItem{Tag: "three", Data: []byte(`hello world`), Updateable: false}
				enqueueItem4 := &v1.EnqueItem{Tag: "four", Data: []byte(`hello world`), Updateable: false}

				g.Expect(queue.Enqueue(enqueueItem1)).ToNot(HaveOccurred())
				g.Expect(queue.Enqueue(enqueueItem2)).ToNot(HaveOccurred())
				g.Expect(queue.Enqueue(enqueueItem3)).ToNot(HaveOccurred())
				g.Expect(queue.Enqueue(enqueueItem4)).ToNot(HaveOccurred())

				metrics := queue.Metrics()
				g.Expect(metrics).ToNot(BeNil())
				g.Expect(metrics.Ready).To(Equal(uint64(4)))
			})
		})
	})

	t.Run("when an enqueued item is updateable", func(t *testing.T) {
		t.Run("it overwrites the item if it is not yet processing", func(t *testing.T) {
			queue := memory.NewQueue(createParams)

			enqueueItem1 := &v1.EnqueItem{Tag: "a", Data: []byte(`one`), Updateable: false}
			enqueueItem2 := &v1.EnqueItem{Tag: "a", Data: []byte(`two`), Updateable: false}
			enqueueItem3 := &v1.EnqueItem{Tag: "a", Data: []byte(`three`), Updateable: false}
			enqueueItem4 := &v1.EnqueItem{Tag: "a", Data: []byte(`four`), Updateable: false}

			g.Expect(queue.Enqueue(enqueueItem1)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(enqueueItem2)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(enqueueItem3)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(enqueueItem4)).ToNot(HaveOccurred())

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics.Ready).To(Equal(uint64(1)))
		})
	})
}
