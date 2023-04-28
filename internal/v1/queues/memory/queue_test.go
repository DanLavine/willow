package memory

import (
	"context"
	"testing"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var (
	createParams             = &v1.Create{Name: datatypes.String("test"), QueueMaxSize: 5, ItemRetryAttempts: 2, DeadLetterQueueMaxSize: 0}
	enqueueItemUpdateable    = &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: datatypes.String("test"), BrokerType: v1.Queue, Tags: datatypes.Strings{"updateable"}}, Data: []byte(`squash me!`), Updateable: true}
	enqueueItemNotUpdateable = &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: datatypes.String("test"), BrokerType: v1.Queue, Tags: datatypes.Strings{"not updateable"}}, Data: []byte(`hello world`), Updateable: false}
)

func TestMemoryQueue_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns the default queue properties", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal(datatypes.String("test")))
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
		g.Expect(metrics.Name).To(Equal(datatypes.String("test")))
		g.Expect(metrics.Total).To(Equal(uint64(1)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.Strings{"updateable"}))
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
		g.Expect(metrics.Name).To(Equal(datatypes.String("test")))
		g.Expect(metrics.Total).To(Equal(uint64(4)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(4)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.Strings{"not updateable"}))
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
		g.Expect(metrics.Name).To(Equal(datatypes.String("test")))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Total).To(Equal(uint64(1)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.Strings{"updateable"}))
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

func TestMemoryQueue_Readers(t *testing.T) {
	g := NewGomegaWithT(t)

	enqueueItem1 := v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name:       "test",
			BrokerType: v1.Queue,
			Tags:       datatypes.Strings{"a", "b", "c"},
		},
		Data:       []byte(`first`),
		Updateable: false,
	}
	enqueueItem2 := v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name:       "test",
			BrokerType: v1.Queue,
			Tags:       datatypes.Strings{"b", "c", "d"},
		},
		Data:       []byte(`first`),
		Updateable: false,
	}
	enqueueItem3 := v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name:       "test",
			BrokerType: v1.Queue,
			Tags:       datatypes.Strings{"c", "d", "e"},
		},
		Data:       []byte(`first`),
		Updateable: false,
	}

	// setup the queue with a proper item
	setupQueue := func(g *GomegaWithT) *Queue {
		queue := NewQueue(createParams)
		g.Expect(queue).ToNot(BeNil())

		// run the queue in the background
		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		g.Expect(taskManager.AddExecuteTask("queue", queue)).ToNot(HaveOccurred())
		go func() {
			_ = taskManager.Run(context.Background())
		}()

		return queue
	}

	t.Run("it return nil if the match query is nil", func(t *testing.T) {
		queue := setupQueue(g)
		defer queue.Stop()

		readers := queue.Readers(nil)
		g.Expect(readers).To(BeNil())
	})

	t.Run("it return nil with an invalid MatchTagsRestrictions", func(t *testing.T) {
		queue := setupQueue(g)
		defer queue.Stop()

		readers := queue.Readers(&v1.MatchQuery{MatchTagsRestrictions: 300})
		g.Expect(readers).To(BeNil())
	})

	t.Run("it updates the processing count on metrics for the tag group", func(t *testing.T) {
		queue := setupQueue(g)
		defer queue.Stop()

		// add an item in into the queue
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())

		readers := queue.Readers(&v1.MatchQuery{MatchTagsRestrictions: v1.ALL})
		g.Expect(len(readers)).To(Equal(1))

		select {
		case <-time.After(time.Second):
			g.Fail("Failed to dequeue item")
		case dequeueFunc := <-readers[0]:
			dequeueFunc()
		}

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics).To(Equal(&v1.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1.TagMetricsResponse{{Processing: 1, Ready: 0, Tags: datatypes.Strings{"not updateable"}}}}))
	})

	t.Run("STRICT match restrictions", func(t *testing.T) {
		t.Run("only finds the reader for a message queue", func(t *testing.T) {
			queue := setupQueue(g)
			defer queue.Stop()

			// add an item in into the queue
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem2)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem3)).ToNot(HaveOccurred())

			readers := queue.Readers(&v1.MatchQuery{MatchTagsRestrictions: v1.STRICT, Tags: datatypes.Strings{"a", "b", "c"}})
			g.Expect(len(readers)).To(Equal(1))

			var dequeuedItemResponses []*v1.DequeueItemResponse
			select {
			case <-time.After(time.Second):
				g.Fail("Failed to dequeue item")
			case dequeueFunc := <-readers[0]:
				dequeuedItemResponses = append(dequeuedItemResponses, dequeueFunc())
			}

			g.Expect(dequeuedItemResponses).To(ContainElement(&v1.DequeueItemResponse{BrokerInfo: v1.BrokerInfo{Name: "test", BrokerType: v1.Queue, Tags: datatypes.Strings{"a", "b", "c"}}, ID: 1, Data: []byte(`first`)}))
		})
	})

	t.Run("SUBSET match restrictions", func(t *testing.T) {
		t.Run("it can find any message that matches the subset", func(t *testing.T) {
			queue := setupQueue(g)
			defer queue.Stop()

			// add an item in into the queue
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem2)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem3)).ToNot(HaveOccurred())

			readers := queue.Readers(&v1.MatchQuery{MatchTagsRestrictions: v1.SUBSET, Tags: datatypes.Strings{"a", "c"}})
			g.Expect(len(readers)).To(Equal(1))

			var dequeuedItemResponses []*v1.DequeueItemResponse
			select {
			case <-time.After(time.Second):
				g.Fail("Failed to dequeue item")
			case dequeueFunc := <-readers[0]:
				dequeuedItemResponses = append(dequeuedItemResponses, dequeueFunc())
			}

			g.Expect(dequeuedItemResponses).To(ContainElement(&v1.DequeueItemResponse{BrokerInfo: v1.BrokerInfo{Name: "test", BrokerType: v1.Queue, Tags: datatypes.Strings{"a", "b", "c"}}, ID: 1, Data: []byte(`first`)}))
		})
	})

	t.Run("ANY match restrictions", func(t *testing.T) {
		t.Run("it can find any message that matches one of the provided tags", func(t *testing.T) {
			queue := setupQueue(g)
			defer queue.Stop()

			// add an item in into the queue
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem2)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem3)).ToNot(HaveOccurred())

			readers := queue.Readers(&v1.MatchQuery{MatchTagsRestrictions: v1.ANY, Tags: datatypes.Strings{"b"}})
			g.Expect(len(readers)).To(Equal(1))

			var dequeuedItemResponses []*v1.DequeueItemResponse
			for i := 0; i < 2; i++ {
				select {
				case <-time.After(time.Second):
					g.Fail("Failed to dequeue item")
				case dequeueFunc := <-readers[0]:
					dequeuedItemResponses = append(dequeuedItemResponses, dequeueFunc())
				}
			}

			g.Expect(dequeuedItemResponses).To(ContainElement(&v1.DequeueItemResponse{BrokerInfo: v1.BrokerInfo{Name: "test", BrokerType: v1.Queue, Tags: datatypes.Strings{"a", "b", "c"}}, ID: 1, Data: []byte(`first`)}))
			g.Expect(dequeuedItemResponses).To(ContainElement(&v1.DequeueItemResponse{BrokerInfo: v1.BrokerInfo{Name: "test", BrokerType: v1.Queue, Tags: datatypes.Strings{"b", "c", "d"}}, ID: 1, Data: []byte(`first`)}))
		})
	})

	t.Run("ALL match restrictions", func(t *testing.T) {
		t.Run("it can find any message regardless of tag groups", func(t *testing.T) {
			queue := setupQueue(g)
			defer queue.Stop()

			// add an item in into the queue
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())

			// add another item
			item2 := *enqueueItemNotUpdateable
			item2.BrokerInfo.Tags = datatypes.Strings{"another", "tag", "set"}
			g.Expect(queue.Enqueue(zap.NewNop(), &item2)).ToNot(HaveOccurred())

			readers := queue.Readers(&v1.MatchQuery{MatchTagsRestrictions: v1.ALL})
			g.Expect(len(readers)).To(Equal(1))

			var dequeuedItemResponses []*v1.DequeueItemResponse

			for i := 0; i < 2; i++ {
				select {
				case <-time.After(time.Second):
					g.Fail("Failed to dequeue item")
				case dequeueFunc := <-readers[0]:
					dequeuedItemResponses = append(dequeuedItemResponses, dequeueFunc())
				}
			}

			g.Expect(dequeuedItemResponses).To(ContainElement(&v1.DequeueItemResponse{BrokerInfo: v1.BrokerInfo{Name: "test", BrokerType: v1.Queue, Tags: datatypes.Strings{"not updateable"}}, ID: 1, Data: []byte(`hello world`)}))
			g.Expect(dequeuedItemResponses).To(ContainElement(&v1.DequeueItemResponse{BrokerInfo: v1.BrokerInfo{Name: "test", BrokerType: v1.Queue, Tags: datatypes.Strings{"another", "tag", "set"}}, ID: 1, Data: []byte(`hello world`)}))
		})
	})
}
