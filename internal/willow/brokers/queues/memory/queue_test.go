package memory

import (
	"context"
	"testing"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var (
	createParams             = &v1willow.Create{Name: "test", QueueMaxSize: 5, DeadLetterQueueMaxSize: 0}
	enqueueItemUpdateable    = &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test", Tags: datatypes.KeyValues{"updateable": datatypes.String("true")}}, Data: []byte(`squash me!`), Updateable: true}
	enqueueItemNotUpdateable = &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test", Tags: datatypes.KeyValues{"updateable": datatypes.String("false")}}, Data: []byte(`hello world`), Updateable: false}
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

		err := queue.Enqueue(zap.NewNop(), &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test"}, Data: []byte(``)})
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("Internal Server Error"))
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
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.KeyValues{"updateable": datatypes.String("true")}))
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
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.KeyValues{"updateable": datatypes.String("false")}))
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
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.KeyValues{"updateable": datatypes.String("true")}))
	})

	t.Run("it can create a queue if child tags already exist", func(t *testing.T) {
		queue := NewQueue(createParams)
		defer queue.Stop()

		enqueue1 := &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test", Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b")}}, Data: []byte(`squash me!`), Updateable: true}
		enqueue2 := &v1willow.EnqueueItemRequest{BrokerInfo: v1willow.BrokerInfo{Name: "test", Tags: datatypes.KeyValues{"a": datatypes.String("a")}}, Data: []byte(`squash me!`), Updateable: true}

		g.Expect(queue.Enqueue(zap.NewNop(), enqueue1)).ToNot(HaveOccurred())
		metrics := queue.Metrics()
		g.Expect(len(metrics.Tags)).To(Equal(1))
		g.Expect(metrics.Tags[0].Tags).To(Equal(datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b")}))

		g.Expect(queue.Enqueue(zap.NewNop(), enqueue2)).ToNot(HaveOccurred())
		metrics = queue.Metrics()
		g.Expect(len(metrics.Tags)).To(Equal(2))
		g.Expect(metrics.Tags).To(ContainElements([]*v1willow.TagMetricsResponse{
			{
				Ready:      1,
				Processing: 0,
				Tags:       datatypes.KeyValues{"a": datatypes.String("a")},
			},
			{
				Ready:      1,
				Processing: 0,
				Tags:       datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b")},
			},
		}))
	})

	t.Run("it returns an error if the item cannot be enqueued because there are to many messages waiting to process", func(t *testing.T) {
		queue := NewQueue(&v1willow.Create{QueueMaxSize: 1})
		defer queue.Stop()

		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)).ToNot(HaveOccurred())

		err := queue.Enqueue(zap.NewNop(), enqueueItemNotUpdateable)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Queue has reached max allowed enqueued items."))
	})
}

func TestMemoryQueue_Dequeue(t *testing.T) {
	g := NewGomegaWithT(t)

	globalSelect := datatypes.AssociatedKeyValuesQuery{}
	bValue := datatypes.String("b")
	bSelection := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{
				"b": {Value: &bValue, ValueComparison: datatypes.EqualsPtr()},
			},
		},
	}
	g.Expect(globalSelect.Validate()).ToNot(HaveOccurred())

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

	t.Run("It returns nil if the taskManager is canceled", func(t *testing.T) {
		queue := NewQueue(createParams)
		g.Expect(queue).ToNot(BeNil())

		// run the queue in the background
		taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
		g.Expect(taskManager.AddExecuteTask("queue", queue)).ToNot(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			_ = taskManager.Run(ctx)
		}()
		cancel()

		var dequeue *v1willow.DequeueItemResponse
		var onSuccess func()
		var onFail func()
		var err *errors.ServerError
		g.Eventually(func() bool {
			dequeue, onSuccess, onFail, err = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
			return true
		}).Should(BeTrue())
		g.Expect(dequeue).To(BeNil())
		g.Expect(onSuccess).To(BeNil())
		g.Expect(onFail).To(BeNil())
		g.Expect(err).ToNot(BeNil())
	})

	t.Run("It returns nil if the cancelContext is canceled and there are no queues", func(t *testing.T) {
		queue, cancel := setupQueue(g)
		defer cancel()

		cancelContext, stopCancelContext := context.WithCancel(context.Background())
		stopCancelContext()

		var dequeue *v1willow.DequeueItemResponse
		var onSuccess func()
		var onFail func()
		var err *errors.ServerError
		g.Eventually(func() bool {
			dequeue, onSuccess, onFail, err = queue.Dequeue(zap.NewNop(), cancelContext, globalSelect)
			return true
		}).Should(BeTrue())
		g.Expect(dequeue).To(BeNil())
		g.Expect(onSuccess).To(BeNil())
		g.Expect(onFail).To(BeNil())
		g.Expect(err).ToNot(BeNil())
	})

	t.Run("It can enqueue a message onto a client that is already waiting", func(t *testing.T) {
		queue, cancel := setupQueue(g)
		defer cancel()

		var dequeue *v1willow.DequeueItemResponse
		var onSuccess func()
		var onFail func()
		var err *errors.ServerError
		received := make(chan struct{})
		go func() {
			dequeue, onSuccess, onFail, err = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
			close(received)
		}()

		// ensure that we eventually have a client waiting
		g.Eventually(func() int {
			queue.clientsLock.Lock()
			defer queue.clientsLock.Unlock()
			return len(queue.clientsWaiting)
		}).Should(Equal(1))

		// Enqueu an item into the queue
		enqueueItemRequest := &v1willow.EnqueueItemRequest{
			BrokerInfo: v1willow.BrokerInfo{
				Name: "test",
				Tags: datatypes.KeyValues{
					"a": datatypes.String("a"),
				},
			},
			Data:       []byte(`squash me!`),
			Updateable: true,
		}
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

		// Eventually we shoud receive the enqueued ite,
		g.Eventually(received).Should(BeClosed())
		g.Expect(dequeue).ToNot(BeNil())
		g.Expect(onSuccess).ToNot(BeNil())
		g.Expect(onFail).ToNot(BeNil())
		g.Expect(err).To(BeNil())

		g.Expect(dequeue.ID).ToNot(Equal(""))
		g.Expect(dequeue.BrokerInfo.Name).To(Equal("test"))
		g.Expect(dequeue.BrokerInfo.Tags).To(Equal(datatypes.KeyValues{"a": datatypes.String("a")}))
		g.Expect(dequeue.Data).To(Equal([]byte(`squash me!`)))
		onSuccess()
	})

	t.Run("It can enqueue a message if a tag group already has one waitiing", func(t *testing.T) {
		queue, cancel := setupQueue(g)
		defer cancel()

		// Enqueu an item into the queue
		enqueueItemRequest := &v1willow.EnqueueItemRequest{
			BrokerInfo: v1willow.BrokerInfo{
				Name: "test",
				Tags: datatypes.KeyValues{
					"a": datatypes.String("a"),
				},
			},
			Data:       []byte(`squash me!`),
			Updateable: true,
		}
		g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

		var dequeue *v1willow.DequeueItemResponse
		var onSuccess func()
		var onFail func()
		var err *errors.ServerError
		received := make(chan struct{})
		go func() {
			dequeue, onSuccess, onFail, err = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
			close(received)
		}()

		// Eventually we shoud receive the enqueued ite,
		g.Eventually(received).Should(BeClosed())
		g.Expect(dequeue).ToNot(BeNil())
		g.Expect(onSuccess).ToNot(BeNil())
		g.Expect(onFail).ToNot(BeNil())
		g.Expect(err).To(BeNil())

		g.Expect(dequeue.ID).ToNot(Equal(""))
		g.Expect(dequeue.BrokerInfo.Name).To(Equal("test"))
		g.Expect(dequeue.BrokerInfo.Tags).To(Equal(datatypes.KeyValues{"a": datatypes.String("a")}))
		g.Expect(dequeue.Data).To(Equal([]byte(`squash me!`)))
		onSuccess()
	})

	t.Run("Describe query selection", func(t *testing.T) {
		t.Run("It won't dequeue a message from any tag groups that don't match the query", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer cancel()

			// Enqueu an item into the queue
			enqueueItemRequest := &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squash me!`),
				Updateable: true,
			}
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

			// ensure that the dequeue operation never finishes
			received := make(chan struct{})
			go func() {
				queue.Dequeue(zap.NewNop(), context.Background(), bSelection)
				close(received)
			}()
			g.Consistently(received).ShouldNot(BeClosed())
		})

		t.Run("It won't enqueue a new message on a waiting tag group if the original query does not match", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer cancel()

			received := make(chan struct{})
			go func() {
				queue.Dequeue(zap.NewNop(), context.Background(), bSelection)
				close(received)
			}()

			g.Consistently(received).ShouldNot(BeClosed())

			// Enqueu an item into the queue
			enqueueItemRequest := &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squash me!`),
				Updateable: true,
			}
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

			// check again that the dequeue operation is waiting
			g.Consistently(received).ShouldNot(BeClosed())
		})
	})

	t.Run("Context when calling OnSuccess callback", func(t *testing.T) {
		t.Run("It removes the value from the queue and doesn't reprocess", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer cancel()

			// Enqueu an item into the queue
			enqueueItemRequest := &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squash me!`),
				Updateable: true,
			}
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

			var onSuccess func()
			received := make(chan struct{})
			go func() {
				_, onSuccess, _, _ = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			// Eventually we shoud receive the enqueued ite,
			g.Eventually(received).Should(BeClosed())
			onSuccess()

			received = make(chan struct{})
			go func() {
				queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			g.Consistently(received).ShouldNot(BeClosed())
		})

		t.Run("It updates the metrics", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer cancel()

			// Enqueu an item into the queue
			enqueueItemRequest := &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squash me!`),
				Updateable: true,
			}
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

			var onSuccess func()
			received := make(chan struct{})
			go func() {
				_, onSuccess, _, _ = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			// Eventually we shoud receive the enqueued ite,
			g.Eventually(received).Should(BeClosed())
			onSuccess()

			metrics := queue.Metrics()
			g.Expect(metrics.Total).To(Equal(uint64(1)))
			g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(0)))
			g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(1)))
		})
	})

	t.Run("Context when calling OnFail callback", func(t *testing.T) {
		t.Run("It requeues the value to reprocess", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer cancel()

			// Enqueu an item into the queue
			enqueueItemRequest := &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squash me!`),
				Updateable: true,
			}
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

			var dequeue *v1willow.DequeueItemResponse
			var onFail func()
			received := make(chan struct{})
			go func() {
				dequeue, _, onFail, _ = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			// Eventually we shoud receive the enqueued ite,
			g.Eventually(received).Should(BeClosed())
			onFail()

			var dequeueAgain *v1willow.DequeueItemResponse
			received = make(chan struct{})
			go func() {
				dequeueAgain, _, _, _ = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			g.Eventually(received).Should(BeClosed())
			g.Expect(dequeue).To(Equal(dequeueAgain))
		})

		t.Run("It allows for another enqueue message to squash the requeued message before processing again", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer cancel()

			// Enqueu an item into the queue
			enqueueItemRequest := &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squash me!`),
				Updateable: true,
			}
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

			var onFail func()
			received := make(chan struct{})
			go func() {
				_, _, onFail, _ = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			// Eventually we shoud receive the enqueued ite,
			g.Eventually(received).Should(BeClosed())

			// enqueue another item before calling onFail
			enqueueItemRequest = &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squashed`),
				Updateable: true,
			}
			enqueuErr := make(chan error)
			go func() {
				enqueuErr <- queue.Enqueue(zap.NewNop(), enqueueItemRequest)
			}()

			onFail()
			g.Eventually(enqueuErr).Should(Receive(BeNil()))

			var dequeueAgain *v1willow.DequeueItemResponse
			received = make(chan struct{})
			go func() {
				dequeueAgain, _, _, _ = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			g.Eventually(received).Should(BeClosed())
			g.Expect(dequeueAgain.Data).To(Equal([]byte(`squashed`)))
		})

		t.Run("It keeps proper metrics", func(t *testing.T) {
			queue, cancel := setupQueue(g)
			defer cancel()

			// Enqueu an item into the queue
			enqueueItemRequest := &v1willow.EnqueueItemRequest{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{
						"a": datatypes.String("a"),
					},
				},
				Data:       []byte(`squash me!`),
				Updateable: true,
			}
			g.Expect(queue.Enqueue(zap.NewNop(), enqueueItemRequest)).ToNot(HaveOccurred())

			var onFail func()
			received := make(chan struct{})
			go func() {
				_, _, onFail, _ = queue.Dequeue(zap.NewNop(), context.Background(), globalSelect)
				close(received)
			}()

			// Eventually we shoud receive the enqueued ite,
			g.Eventually(received).Should(BeClosed())
			onFail()

			metrics := queue.Metrics()
			g.Expect(metrics.Total).To(Equal(uint64(1)))
			g.Expect(metrics.Tags[0].Ready).To(Equal(uint64(1)))
			g.Expect(metrics.Tags[0].Processing).To(Equal(uint64(0)))
		})
	})
}

func TestMemoryQueue_ACK(t *testing.T) {
	g := NewGomegaWithT(t)

	enqueueItem1 := v1willow.EnqueueItemRequest{
		BrokerInfo: v1willow.BrokerInfo{
			Name: "test",
			Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
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
		queue, cancel := setupQueue(g)
		defer cancel()

		ack := &v1willow.ACK{
			BrokerInfo: v1willow.BrokerInfo{
				Name: "test",
				Tags: datatypes.KeyValues{"not": datatypes.String("found")},
			},
			ID:     "not found",
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

		ack := &v1willow.ACK{
			BrokerInfo: v1willow.BrokerInfo{
				Name: "test",
				Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "not": datatypes.String("found")},
			},
			ID:     "not found",
			Passed: true,
		}
		g.Expect(ack.Validate()).ToNot(HaveOccurred())

		err := queue.ACK(zap.NewNop(), ack)
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("tag group not found"))
	})

	t.Run("Context when ack is set to passed", func(t *testing.T) {
		setupWithEnqueuedItem := func(g *WithT) (*Queue, *v1willow.DequeueItemResponse, context.CancelFunc) {
			queue, cancel := setupQueue(g)

			// setup reader
			dequeuRequest := &v1willow.DequeueItemRequest{
				Name:  "test",
				Query: datatypes.AssociatedKeyValuesQuery{},
			}
			g.Expect(dequeuRequest.Validate()).ToNot(HaveOccurred())

			// enqueue an item for processing
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())

			// dequeue the item like a client so its processing
			dequeueItem, success, _, err := queue.Dequeue(zap.NewNop(), context.Background(), dequeuRequest.Query)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dequeueItem).ToNot(BeNil())
			success()

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1willow.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1willow.TagMetricsResponse{{Processing: 1, Ready: 0, Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")}}}}))

			return queue, dequeueItem, cancel
		}

		t.Run("it returns an error if the ID cannt be found ", func(t *testing.T) {
			queue, _, cancel := setupWithEnqueuedItem(g)
			defer cancel()

			ack := &v1willow.ACK{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
				},
				ID:     "not_found",
				Passed: true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queue.ACK(zap.NewNop(), ack)
			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(ContainSubstring("ID not_found does not exist for tag group"))

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1willow.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1willow.TagMetricsResponse{{Processing: 1, Ready: 0, Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")}}}}))
		})

		t.Run("it removes the item from processing with a proper ID and removes the tag group if there are no more enqueued items", func(t *testing.T) {
			queue, dequeueItem, cancel := setupWithEnqueuedItem(g)
			defer cancel()

			ack := &v1willow.ACK{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
				},
				ID:     dequeueItem.ID,
				Passed: true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queue.ACK(zap.NewNop(), ack)
			g.Expect(err).To(BeNil())

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1willow.QueueMetricsResponse{Name: "test", Total: 0, Max: 5, Tags: nil}))
		})

		t.Run("it removes the item from processing with a proper ID and leaves the tag group if there are more enqueued items", func(t *testing.T) {
			queue, dequeueItem, cancel := setupWithEnqueuedItem(g)
			defer cancel()

			// enqueue a second item
			g.Expect(queue.Enqueue(zap.NewNop(), &enqueueItem1)).ToNot(HaveOccurred())

			ack := &v1willow.ACK{
				BrokerInfo: v1willow.BrokerInfo{
					Name: "test",
					Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")},
				},
				ID:     dequeueItem.ID,
				Passed: true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queue.ACK(zap.NewNop(), ack)
			g.Expect(err).To(BeNil())

			metrics := queue.Metrics()
			g.Expect(metrics).ToNot(BeNil())
			g.Expect(metrics).To(Equal(&v1willow.QueueMetricsResponse{Name: "test", Total: 1, Max: 5, Tags: []*v1willow.TagMetricsResponse{{Processing: 0, Ready: 1, Tags: datatypes.KeyValues{"a": datatypes.String("a"), "b": datatypes.String("b"), "c": datatypes.String("c")}}}}))
		})
	})

	t.Run("when ack is set to failed", func(t *testing.T) {
		// TODO
		// 1. Should set to dead letter queue
		// 2. Should be able to requeue if that is configured
	})
}
