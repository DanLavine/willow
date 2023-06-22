package memory

import (
	"context"
	"reflect"
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
	createParams             = &v1.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0}
	enqueueItemUpdateable    = &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test", Tags: datatypes.StringMap{"updateable": datatypes.String("true")}}, Data: []byte(`squash me!`), Updateable: true}
	enqueueItemNotUpdateable = &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test", Tags: datatypes.StringMap{"updateable": datatypes.String("false")}}, Data: []byte(`hello world`), Updateable: false}
)

func TestMemoryQueue_Metrics(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns the default queue properties", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Total).To(Equal(uint64(0)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Tags).To(BeNil())
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())
	})
}

func TestMemoryQueue_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the tags are empty", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		err := queue.Enqueue(zap.NewNop(), &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test"}, Data: []byte(``)})
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("Internal Server Error. Actual: keyValuePairs requires a length of at least 1."))
	})

	t.Run("it records that an item is ready", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemUpdateable)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Total).To(Equal(uint64(1)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.StringMap{"updateable": datatypes.String("true")}))
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
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Total).To(Equal(uint64(4)))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(4)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.StringMap{"updateable": datatypes.String("false")}))
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
		g.Expect(metrics.Name).To(Equal("test"))
		g.Expect(metrics.Max).To(Equal(uint64(5)))
		g.Expect(metrics.Total).To(Equal(uint64(1)))
		g.Expect(metrics.DeadLetterQueueMetrics).To(BeNil())

		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.StringMap{"updateable": datatypes.String("true")}))
	})

	t.Run("it can create a queue if child tags already exist", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		enqueue1 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test", Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b")}}, Data: []byte(`squash me!`), Updateable: true}
		enqueue2 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test", Tags: datatypes.StringMap{"a": datatypes.String("a")}}, Data: []byte(`squash me!`), Updateable: true}

		g.Expect(queue.Enqueue(zap.NewNop(), enqueue1)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b")}))

		g.Expect(queue.Enqueue(zap.NewNop(), enqueue2)).ToNot(HaveOccurred())
		metrics = queue.Metrics()
		g.Expect(len(metrics.Tags)).To(Equal(2))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.StringMap{"a": datatypes.String("a")}))
		g.Expect(metrics.Tags[1].Tags).To(Equal(datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b")}))
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
			Name: "test",
			Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
		},
		Data:       []byte(`first`),
		Updateable: false,
	}
	enqueueItem2 := v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name: "test",
			Tags: datatypes.StringMap{"b": datatypes.String("b"), "c": datatypes.String("c"), "d": datatypes.String("d")},
		},
		Data:       []byte(`second`),
		Updateable: false,
	}
	enqueueItem3 := v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name: "test",
			Tags: datatypes.StringMap{"c": datatypes.String("c"), "d": datatypes.String("d"), "e": datatypes.String("e")},
		},
		Data:       []byte(`third`),
		Updateable: false,
	}
	enqueueItem4 := v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name: "test",
			Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c"), "d": datatypes.String("d")},
		},
		Data:       []byte(`fourth`),
		Updateable: false,
	}

	setupQueue := func(g *GomegaWithT) (*Queue, context.CancelFunc) {
		queue := NewQueue(createParams)
		g.Expect(queue).ToNot(BeNil())

		// run the queue in the background
		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		g.Expect(taskManager.AddExecuteTask("queue", queue)).ToNot(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			_ = taskManager.Run(ctx)
		}()

		return queue, cancel
	}

	t.Run("it will create the requested readers even if they do not yet exist", func(t *testing.T) {
		queue, cancel := setupQueue(g)
		defer queue.Stop()
		defer cancel()

		readerSelect := &v1.ReaderSelect{
			BrokerName: "test",
			Queries: []v1.ReaderQuery{
				{Type: v1.ReaderExactly, Tags: datatypes.StringMap{"shouldn't be there yet": datatypes.String("ok")}},
				{Type: v1.ReaderMatches, Tags: datatypes.StringMap{"shouldn't be there yet": datatypes.String("ok")}},
			},
		}
		g.Expect(readerSelect.Validate()).ToNot(HaveOccurred())

		readers, err := queue.Readers(zap.NewNop(), readerSelect)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(readers)).To(Equal(2))
	})

	t.Run("it updates the processing count on metrics for the tag group", func(t *testing.T) {
		queue, cancel := setupQueue(g)
		defer queue.Stop()
		defer cancel()

		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())

		metrics := queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics).To(Equal(&v1.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1.TagMetricsResponse{{Processing: 0, Ready: 1, Tags: datatypes.StringMap{"updateable": datatypes.String("false")}}}}))

		readerSelect := &v1.ReaderSelect{
			BrokerName: "test",
			Queries: []v1.ReaderQuery{
				{Type: v1.ReaderExactly, Tags: datatypes.StringMap{"updateable": datatypes.String("false")}},
			},
		}
		g.Expect(readerSelect.Validate()).ToNot(HaveOccurred())

		readers, err := queue.Readers(zap.NewNop(), readerSelect)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(readers)).To(Equal(1))

		select {
		case <-time.After(time.Second):
			g.Fail("Failed to dequeue item")
		case dequeueFunc := <-readers[0]:
			dequeueFunc()
		}

		metrics = queue.Metrics()
		g.Expect(metrics).ToNot(BeNil())
		g.Expect(metrics).To(Equal(&v1.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1.TagMetricsResponse{{Processing: 1, Ready: 0, Tags: datatypes.StringMap{"updateable": datatypes.String("false")}}}}))
	})

	t.Run("'nil' query selection", func(t *testing.T) {
		t.Run("it return the global reader if the readerSelect is nil", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer queue.Stop()
			defer cancel()

			readers, err := queue.Readers(zap.NewNop(), nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(readers)).To(Equal(1))

			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())

			select {
			case <-time.After(time.Second):
				g.Fail("Failed to dequeue item")
			case dequeueFunc := <-readers[0]:
				g.Expect(dequeueFunc().Data).To(Equal([]byte(`first`)))
			}
		})
	})

	t.Run("'Match' query selections", func(t *testing.T) {
		t.Run("finds the reader for any message queues that use the provided tag pairs", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer queue.Stop()
			defer cancel()

			// add an item in into the queue
			// should only pull from enqueueItem1 and enqueueItem4
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem2)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem3)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem4)).ToNot(HaveOccurred())

			readerSelect := &v1.ReaderSelect{
				BrokerName: "test",
				Queries: []v1.ReaderQuery{
					{Type: v1.ReaderMatches, Tags: datatypes.StringMap{"a": datatypes.String("a")}},
				},
			}
			g.Expect(readerSelect.Validate()).ToNot(HaveOccurred())

			readers, err := queue.Readers(zap.NewNop(), readerSelect)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(readers)).To(Equal(1))

			itemOneFound, itemFourFound := false, false
			for i := 0; i < 2; i++ {
				select {
				case <-time.After(time.Second):
					g.Fail("Failed to dequeue item")
				case dequeueFunc := <-readers[0]:
					dequeuedItemResponses := dequeueFunc()
					if reflect.DeepEqual(dequeuedItemResponses.BrokerInfo.Tags, enqueueItem1.Tags) {
						itemOneFound = true
						g.Expect(dequeuedItemResponses.Data).To(Equal([]byte(`first`)))
					} else if reflect.DeepEqual(dequeuedItemResponses.BrokerInfo.Tags, enqueueItem4.Tags) {
						itemFourFound = true
						g.Expect(dequeuedItemResponses.Data).To(Equal([]byte(`fourth`)))
					} else {
						g.Fail("received an unexpected response")
					}
				}
			}

			g.Expect(itemOneFound).To(BeTrue())
			g.Expect(itemFourFound).To(BeTrue())
		})
	})

	t.Run("'Exact' query selections", func(t *testing.T) {
		t.Run("finds only the reader that matches the tags exactly", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer queue.Stop()
			defer cancel()

			// add an item in into the queue
			// should only pull from enqueueItem1 and enqueueItem4
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem2)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem3)).ToNot(HaveOccurred())
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem4)).ToNot(HaveOccurred())

			readerSelect := &v1.ReaderSelect{
				BrokerName: "test",
				Queries: []v1.ReaderQuery{
					{Type: v1.ReaderExactly, Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")}},
				},
			}
			g.Expect(readerSelect.Validate()).ToNot(HaveOccurred())

			readers, err := queue.Readers(zap.NewNop(), readerSelect)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(readers)).To(Equal(1))

			select {
			case <-time.After(time.Second):
				g.Fail("Failed to dequeue item")
			case dequeueFunc := <-readers[0]:
				dequeuedItemResponses := dequeueFunc()
				g.Expect(dequeuedItemResponses.Data).To(Equal([]byte(`first`)))
			}
		})
	})
}

func TestMemoryQueue_ACK(t *testing.T) {
	g := NewGomegaWithT(t)

	enqueueItem1 := v1.EnqueueItemRequest{
		BrokerInfo: v1.BrokerInfo{
			Name: "test",
			Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
		},
		Data:       []byte(`first`),
		Updateable: false,
	}

	setupQueue := func(g *GomegaWithT) (*Queue, context.CancelFunc) {
		queue := NewQueue(createParams)
		g.Expect(queue).ToNot(BeNil())

		// run the queue in the background
		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		g.Expect(taskManager.AddExecuteTask("queue", queue)).ToNot(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			_ = taskManager.Run(ctx)
		}()

		return queue, cancel
	}

	t.Run("it returns an error if the requested tags cannt be found", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		ack := &v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: "test",
				Tags: datatypes.StringMap{"not": datatypes.String("found")},
			},
			ID:     1,
			Passed: true,
		}
		g.Expect(ack.Validate()).ToNot(HaveOccurred())

		err := queue.ACK(zap.NewNop(), ack)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("tag group not found"))
	})

	t.Run("it returns an error if the exact tags cannt be found", func(t *testing.T) {
		queue, cancel := setupQueue(g)
		defer cancel()

		ack := &v1.ACK{
			BrokerInfo: v1.BrokerInfo{
				Name: "test",
				Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "not": datatypes.String("found")},
			},
			ID:     1,
			Passed: true,
		}
		g.Expect(ack.Validate()).ToNot(HaveOccurred())

		err := queue.ACK(zap.NewNop(), ack)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("tag group not found"))
	})

	t.Run("when ack is set to passed", func(t *testing.T) {
		setupWithEnqueuedItem := func(g *WithT) (*Queue, *v1.DequeueItemResponse, context.CancelFunc) {
			queue, cancel := setupQueue(g)

			// setup reader
			readerSelect := &v1.ReaderSelect{
				BrokerName: "test",
				Queries: []v1.ReaderQuery{
					{Type: v1.ReaderMatches, Tags: datatypes.StringMap{"a": datatypes.String("a")}},
				},
			}
			g.Expect(readerSelect.Validate()).ToNot(HaveOccurred())

			// enqueue an item for processing
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())

			// dequeue the item like a client so its processing
			readers, err := queue.Readers(zap.NewNop(), readerSelect)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(readers)).To(Equal(1))

			var dequeueItemResponse *v1.DequeueItemResponse
			select {
			case <-time.After(time.Second):
				g.Fail("Failed to dequeue item")
			case dequeueFunc := <-readers[0]:
				dequeueItemResponse = dequeueFunc()
			}

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1.TagMetricsResponse{{Processing: 1, Ready: 0, Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")}}}}))

			return queue, dequeueItemResponse, cancel
		}

		t.Run("it returns an error if the id is not processing", func(t *testing.T) {
			queue, _, cancel := setupWithEnqueuedItem(g)
			defer cancel()

			// enqueue an item for processing
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())

			ack := &v1.ACK{
				BrokerInfo: v1.BrokerInfo{
					Name: "test",
					Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
				},
				ID:     2,
				Passed: true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queue.ACK(zap.NewNop(), ack)
			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(ContainSubstring("ID 2 is not processing"))

		})

		t.Run("it returns an error if the ID cannt be found ", func(t *testing.T) {
			queue, _, cancel := setupWithEnqueuedItem(g)
			defer cancel()

			ack := &v1.ACK{
				BrokerInfo: v1.BrokerInfo{
					Name: "test",
					Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
				},
				ID:     9_123_098,
				Passed: true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queue.ACK(zap.NewNop(), ack)
			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(ContainSubstring("ID 9123098 does not exist for tag group"))

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1.TagMetricsResponse{{Processing: 1, Ready: 0, Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")}}}}))
		})

		t.Run("it removes the item from processing with a proper ID and removes the tag group if there are no more enqueued items", func(t *testing.T) {
			queue, _, cancel := setupWithEnqueuedItem(g)
			defer cancel()

			ack := &v1.ACK{
				BrokerInfo: v1.BrokerInfo{
					Name: "test",
					Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
				},
				ID:     1,
				Passed: true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queue.ACK(zap.NewNop(), ack)
			g.Expect(err).To(BeNil())

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1.QueueMetricsResponse{Name: "test", Total: 0, Max: 5, Tags: nil}))
		})

		t.Run("it removes the item from processing with a proper ID and leaves the tag group if there are more enqueued items", func(t *testing.T) {
			queue, _, cancel := setupWithEnqueuedItem(g)
			defer cancel()

			// enqueue a second item
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())

			ack := &v1.ACK{
				BrokerInfo: v1.BrokerInfo{
					Name: "test",
					Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
				},
				ID:     1,
				Passed: true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queue.ACK(zap.NewNop(), ack)
			g.Expect(err).To(BeNil())

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1.TagMetricsResponse{{Processing: 0, Ready: 1, Tags: datatypes.StringMap{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")}}}}))
		})
	})

	t.Run("when ack is set to failed", func(t *testing.T) {
		// TODO
		// 1. Should set to dead letter queue
		// 2. Should be able to requeue if that is configured
	})
}
