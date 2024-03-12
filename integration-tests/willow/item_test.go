package willow_integration_tests

import (
	"context"
	"testing"
	"time"

	willowclient "github.com/DanLavine/willow/pkg/clients/willow_client"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func Test_Queue_ItemACK(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context success", func(t *testing.T) {
		t.Run("It removes the item and updates the Limiter", func(t *testing.T) {
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
			g.Expect(willowClient.CreateQueue(createQueue, nil)).ToNot(HaveOccurred())

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
			err := willowClient.EnqueueQueueItem("test queue", enqueueQueueItem, nil)
			g.Expect(err).ToNot(HaveOccurred())

			// dequeue the item
			item, err := willowClient.DequeueQueueItem(context.Background(), "test queue", &v1common.AssociatedQuery{}, nil)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item.Data()).To(Equal([]byte(`data for first item`)))

			// ensure the counters are setup properly
			counters, err := limiterClient.QueryCounters(&v1common.AssociatedQuery{}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(counters)).To(Equal(2))
			g.Expect(counters).To(ContainElements(
				&v1limiter.Counter{ // counter for enqueued item
					Counters: 1,
					KeyValues: datatypes.KeyValues{
						"_willow_queue_name": datatypes.String("test queue"),
						"_willow_enqueued":   datatypes.String("true"),
						"_willow_one":        datatypes.Int(1),
					},
				},
				&v1limiter.Counter{ // counter for running item
					Counters: 1,
					KeyValues: datatypes.KeyValues{
						"_willow_queue_name": datatypes.String("test queue"),
						"_willow_running":    datatypes.String("true"),
						"one":                datatypes.Int(1),
					},
				},
			))

			// ack the item
			err = item.ACK(true, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item.Done()).To(BeClosed())

			// ensure the counters are updated properly
			counters, err = limiterClient.QueryCounters(&v1common.AssociatedQuery{}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(counters)).To(Equal(0))
		})
	})

	t.Run("Context failure", func(t *testing.T) {
		t.Run("It re-queues the item up to the max retry count and updates the Limiter", func(t *testing.T) {
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
			g.Expect(willowClient.CreateQueue(createQueue, nil)).ToNot(HaveOccurred())

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
			err := willowClient.EnqueueQueueItem("test queue", enqueueQueueItem, nil)
			g.Expect(err).ToNot(HaveOccurred())

			// dequeue the item
			var item *willowclient.Item
			var dequeueErr error
			done := make(chan struct{})
			go func() {
				defer close(done)
				item, dequeueErr = willowClient.DequeueQueueItem(context.Background(), "test queue", &v1common.AssociatedQuery{}, nil)
			}()
			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueErr).ToNot(HaveOccurred())
			g.Expect(item.Data()).To(Equal([]byte(`data for first item`)))

			// ensure the counters are setup properly
			counters, err := limiterClient.QueryCounters(&v1common.AssociatedQuery{}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(counters)).To(Equal(2))
			g.Expect(counters).To(ContainElements(
				&v1limiter.Counter{ // counter for enqueued item
					Counters: 1,
					KeyValues: datatypes.KeyValues{
						"_willow_queue_name": datatypes.String("test queue"),
						"_willow_enqueued":   datatypes.String("true"),
						"_willow_one":        datatypes.Int(1),
					},
				},
				&v1limiter.Counter{ // counter for running item
					Counters: 1,
					KeyValues: datatypes.KeyValues{
						"_willow_queue_name": datatypes.String("test queue"),
						"_willow_running":    datatypes.String("true"),
						"one":                datatypes.Int(1),
					},
				},
			))

			// ack the item
			err = item.ACK(false, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item.Done()).To(BeClosed())

			// ensure the counters are updated properly
			counters, err = limiterClient.QueryCounters(&v1common.AssociatedQuery{}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(counters)).To(Equal(1))
			g.Expect(counters).To(ContainElements(
				&v1limiter.Counter{ // counter for enqueued item
					Counters: 1,
					KeyValues: datatypes.KeyValues{
						"_willow_queue_name": datatypes.String("test queue"),
						"_willow_enqueued":   datatypes.String("true"),
						"_willow_one":        datatypes.Int(1),
					},
				},
			))

			// dequeue the item again
			done = make(chan struct{})
			go func() {
				defer close(done)
				item, dequeueErr = willowClient.DequeueQueueItem(context.Background(), "test queue", &v1common.AssociatedQuery{}, nil)
			}()
			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueErr).ToNot(HaveOccurred())
			g.Expect(item.Data()).To(Equal([]byte(`data for first item`)))

			// ack the item again
			err = item.ACK(false, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(item.Done()).To(BeClosed())

			// ensure the item has not been requeued as it hit the limiter count
			counters, err = limiterClient.QueryCounters(&v1common.AssociatedQuery{}, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(len(counters)).To(Equal(0))
		})
	})
}
