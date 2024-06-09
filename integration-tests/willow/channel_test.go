package willow_integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/helpers"
	willowclient "github.com/DanLavine/willow/pkg/clients/willow_client"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Queue_Enqueue(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can enqueue an Item to a queue", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// enqueue the item
		enqueueQueueItem := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for first item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		err := willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counters are setup properly
		counters, err := limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(1))
	})

	t.Run("It can enqueu multiple items to different channels on one queue", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// enqueue multiple item
		enqueueQueueItem1 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for first item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem1)).ToNot(HaveOccurred())

		enqueueQueueItem2 := &v1willow.EnqueueQueueItem{ // updates the previous item
			Item: []byte(`data for second item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem2)).ToNot(HaveOccurred())

		enqueueQueueItem3 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for third item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
				"two": datatypes.Int(2),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem3)).ToNot(HaveOccurred())

		enqueueQueueItem4 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for fourth item`),
			KeyValues: datatypes.KeyValues{
				"one":   datatypes.Int(1),
				"two":   datatypes.Int(2),
				"three": datatypes.String("3"),
			},
			Updateable:      false,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem4)).ToNot(HaveOccurred())

		enqueueQueueItem5 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for fifth item`),
			KeyValues: datatypes.KeyValues{
				"one":   datatypes.Int(1),
				"two":   datatypes.Int(2),
				"three": datatypes.String("3"),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem5)).ToNot(HaveOccurred())

		// ensure the counters are setup properly
		counters, err := limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(3))
		g.Expect(counters).To(ContainElements(
			//	these are all the enqueued values for the provided items
			&v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
			&v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
							"_willow_two":        datatypes.Int(2),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
			&v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
							"_willow_two":        datatypes.Int(2),
							"_willow_three":      datatypes.String("3"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](2),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
		))
	})

	t.Run("It respects the limits setup on a queue", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 2,
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// enqueue multiple item
		// ensure the first itme can collapse
		for i := 0; i < 3; i++ {
			updateable := false
			if i == 0 {
				updateable = true
			}

			enqueueQueueItem := &v1willow.EnqueueQueueItem{
				Item: []byte(fmt.Sprintf(`data for item %d`, i)),
				KeyValues: datatypes.KeyValues{
					"one": datatypes.Int(1),
				},
				Updateable:      updateable,
				RetryAttempts:   1,
				RetryPosition:   "front",
				TimeoutDuration: 5 * time.Second,
			}
			g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem)).ToNot(HaveOccurred())
		}

		// next item enqueued should error
		enqueueQueueItem := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for item 3`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      false,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		err := willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Queue has reached the total number of allowed queue items"))

		// ensure the counters are setup properly
		counters, err := limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(1))
		g.Expect(counters).To(ContainElements(
			&v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](2),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
		))
	})
}

func Test_Queue_Dequeue(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can stop the dequeue operation when the context is canceled", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// dequeue the item
		var item *willowclient.Item
		var err error

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			item, err = willowClient.DequeueQueueItem(ctx, "test queue", &queryassociatedaction.AssociatedActionQuery{})
		}()
		g.Consistently(done).ShouldNot(BeClosed())

		// cancel the dequeue
		cancel()
		g.Eventually(done).Should(BeClosed())
		g.Expect(item).To(BeNil())
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("context canceled"))
	})

	t.Run("It can dequeue an item that is already enqueued", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// enqueue the item
		enqueueQueueItem := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for first item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		err := willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem)
		g.Expect(err).ToNot(HaveOccurred())

		// dequeue the item
		item, err := willowClient.DequeueQueueItem(context.Background(), "test queue", &queryassociatedaction.AssociatedActionQuery{})

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(item.Data()).To(Equal([]byte(`data for first item`)))

		// ensure the counters are setup properly
		counters, err := limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(2))
		g.Expect(counters).To(ContainElements(
			&v1limiter.Counter{ // counter for enqueued item
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
			&v1limiter.Counter{ // counter for running item
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_running":    datatypes.String("true"),
							"one":                datatypes.Int(1),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
		))
	})

	t.Run("It can dequeue an item this is sent after the client is already waiting", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// dequeue the item
		var item *willowclient.Item
		var dequeueErr error

		done := make(chan struct{})
		go func() {
			defer close(done)
			item, dequeueErr = willowClient.DequeueQueueItem(context.Background(), "test queue", &queryassociatedaction.AssociatedActionQuery{})
		}()

		// ensure that the service is attempting to dequeue the item before sending the enqueue request
		g.Eventually(func() string {
			return willowTestConstruct.ServerStdout.String()
		}).Should(ContainSubstring("waiting for available item"))

		// enqueue the item
		enqueueQueueItem := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for first item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		err := willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem)
		g.Expect(err).ToNot(HaveOccurred())

		// dequeue should recieve
		g.Eventually(done).Should(BeClosed())
		g.Expect(dequeueErr).ToNot(HaveOccurred())
		g.Expect(item.Data()).To(Equal([]byte(`data for first item`)))

		// ensure the counters are setup properly
		counters, err := limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(2))
		g.Expect(counters).To(ContainElements(
			&v1limiter.Counter{ // counter for enqueued item
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
			&v1limiter.Counter{ // counter for running item
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_running":    datatypes.String("true"),
							"one":                datatypes.Int(1),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
		))
	})
}

func Test_Queue_DeleteChannel(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can delete a channel with all enqueued items", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// enqueue multiple item
		enqueueQueueItem1 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for first item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem1)).ToNot(HaveOccurred())

		enqueueQueueItem2 := &v1willow.EnqueueQueueItem{ // updates the previous item
			Item: []byte(`data for second item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem2)).ToNot(HaveOccurred())

		enqueueQueueItem3 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for third item`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
				"two": datatypes.Int(2),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem3)).ToNot(HaveOccurred())

		enqueueQueueItem4 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for fourth item`),
			KeyValues: datatypes.KeyValues{
				"one":   datatypes.Int(1),
				"two":   datatypes.Int(2),
				"three": datatypes.String("3"),
			},
			Updateable:      false,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem4)).ToNot(HaveOccurred())

		enqueueQueueItem5 := &v1willow.EnqueueQueueItem{
			Item: []byte(`data for fifth item`),
			KeyValues: datatypes.KeyValues{
				"one":   datatypes.Int(1),
				"two":   datatypes.Int(2),
				"three": datatypes.String("3"),
			},
			Updateable:      true,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: 5 * time.Second,
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem5)).ToNot(HaveOccurred())

		// delete a channel
		err := willowClient.DeleteQueueChannel(context.Background(), "test queue", datatypes.KeyValues{"one": datatypes.Int(1)})
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counters are setup properly
		counters, err := limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(2))
		g.Expect(counters).To(ContainElements(
			// these are all the enqueued values for the provided items
			&v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
							"_willow_two":        datatypes.Int(2),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](1),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
			&v1limiter.Counter{
				Spec: &v1limiter.CounterSpec{
					DBDefinition: &v1limiter.CounterDBDefinition{
						KeyValues: dbdefinition.TypedKeyValues{
							"_willow_queue_name": datatypes.String("test queue"),
							"_willow_enqueued":   datatypes.String("true"),
							"_willow_one":        datatypes.Int(1),
							"_willow_two":        datatypes.Int(2),
							"_willow_three":      datatypes.String("3"),
						},
					},
					Properties: &v1limiter.CounteProperties{
						Counters: helpers.PointerOf[int64](2),
					},
				},
				State: &v1limiter.CounterState{
					Deleting: false,
				},
			},
		))
	})
}
