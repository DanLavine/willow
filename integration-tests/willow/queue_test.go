package willow_integration_tests

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	willowclient "github.com/DanLavine/willow/pkg/clients/willow_client"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func setupLimitterClient(g *GomegaWithT, url string) limiterclient.LimiterClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:             url,
		ContentEncoding: api.ContentTypeJSON,
		CAFile:          filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}

	limiterClient, err := limiterclient.NewLimiterClient(cfg)
	g.Expect(err).ToNot(HaveOccurred())

	return limiterClient
}

func setupWillowClient(g *GomegaWithT, url string) willowclient.WillowServiceClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:             url,
		ContentEncoding: api.ContentTypeJSON,
		CAFile:          filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile:   filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}

	willowClient, err := willowclient.NewWillowClient(cfg)
	g.Expect(err).ToNot(HaveOccurred())

	return willowClient
}

func Test_Queue_Create(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can create a queue with proper name", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)

		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}

		err := willowClient.CreateQueue(createQueue, nil)

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It can have multiple queues with different names", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)
		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)

		for i := 0; i < 5; i++ {
			createQueue := &v1willow.QueueCreate{
				Name:         fmt.Sprintf("test queue %d", i),
				QueueMaxSize: 5,
			}

			err := willowClient.CreateQueue(createQueue, nil)
			g.Expect(err).ToNot(HaveOccurred())
		}
	})
}

func Test_Queue_List(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It list all queues without their channels", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)

		for i := 0; i < 5; i++ {
			createQueue := &v1willow.QueueCreate{
				Name:         fmt.Sprintf("test queue %d", i),
				QueueMaxSize: 5,
			}

			err := willowClient.CreateQueue(createQueue, nil)
			g.Expect(err).ToNot(HaveOccurred())
		}

		queues, err := willowClient.ListQueues(nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(queues)).To(Equal(5))
	})
}

func Test_Queue_Get(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can retrieve a specific queue", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)
		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)

		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(createQueue, nil)).ToNot(HaveOccurred())

		queue, err := willowClient.GetQueue("test queue", &v1common.AssociatedQuery{}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(queue.Name).To(Equal("test queue"))
		g.Expect(queue.QueueMaxSize).To(Equal(int64(5)))
		g.Expect(len(queue.Channels)).To(Equal(0))
	})

	t.Run("It can retrieve specific channels that are queried for", func(t *testing.T) {
		// TODO
	})
}

func Test_Queue_Update(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can update the max queue size", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)

		// create
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(createQueue, nil)).ToNot(HaveOccurred())

		// update
		updateQueue := &v1willow.QueueUpdate{
			QueueMaxSize: 12,
		}
		err := willowClient.UpdateQueue("test queue", updateQueue, nil)
		g.Expect(err).ToNot(HaveOccurred())

		// get
		queue, err := willowClient.GetQueue("test queue", &v1common.AssociatedQuery{}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(queue.Name).To(Equal("test queue"))
		g.Expect(queue.QueueMaxSize).To(Equal(int64(12)))
		g.Expect(len(queue.Channels)).To(Equal(0))
	})
}

func Test_Queue_Delete(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	t.Run("It can delete an empty queue", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// create
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(createQueue, nil)).ToNot(HaveOccurred())

		// ensure the override exists before deletion
		rule, err := limiterClient.GetRule("_willow_queue_enqueued_limits", &v1limiter.RuleGet{OverridesToMatch: &v1common.MatchQuery{}}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(rule.Overrides)).To(Equal(1))

		// delete
		err = willowClient.DeleteQueue("test queue", nil)
		g.Expect(err).ToNot(HaveOccurred())

		fmt.Println(willowTestConstruct.ServerStdout.String())
		// get
		queue, err := willowClient.GetQueue("test queue", &v1common.AssociatedQuery{}, nil)
		g.Expect(queue).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to find queue 'test queue' by name"))

		// ensure the override is deleted
		rule, err = limiterClient.GetRule("_willow_queue_enqueued_limits", &v1limiter.RuleGet{OverridesToMatch: &v1common.MatchQuery{}}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(rule.Overrides)).To(Equal(0))
	})

	t.Run("It can delete a queue and all Limiter counters for enqueued items", func(t *testing.T) {
		t.Parallel()

		lockerTestConstruct := StartLocker(g)
		defer lockerTestConstruct.Shutdown(g)

		limiterTestConstruct := StartLimiter(g, lockerTestConstruct.ServerURL)
		defer limiterTestConstruct.Shutdown(g)

		willowTestConstruct := StartWillow(g, limiterTestConstruct.ServerURL)
		defer willowTestConstruct.Shutdown(g)

		willowClient := setupWillowClient(g, willowTestConstruct.ServerURL)
		limiterClient := setupLimitterClient(g, limiterTestConstruct.ServerURL)

		// ensure the counters are in a clean state to start
		counters, err := limiterClient.QueryCounters(&v1common.AssociatedQuery{}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(0))

		// setup queue
		createQueue := &v1willow.QueueCreate{
			Name:         "test queue",
			QueueMaxSize: 5,
		}
		g.Expect(willowClient.CreateQueue(createQueue, nil)).ToNot(HaveOccurred())

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
		g.Expect(willowClient.EnqueueQueueItem("test queue", enqueueQueueItem1, nil)).ToNot(HaveOccurred())

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
		g.Expect(willowClient.EnqueueQueueItem("test queue", enqueueQueueItem2, nil)).ToNot(HaveOccurred())

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
		g.Expect(willowClient.EnqueueQueueItem("test queue", enqueueQueueItem3, nil)).ToNot(HaveOccurred())

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
		g.Expect(willowClient.EnqueueQueueItem("test queue", enqueueQueueItem4, nil)).ToNot(HaveOccurred())

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
		g.Expect(willowClient.EnqueueQueueItem("test queue", enqueueQueueItem5, nil)).ToNot(HaveOccurred())

		// delete a channel
		err = willowClient.DeleteQueue("test queue", nil)
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counters are destroyed properly
		counters, err = limiterClient.QueryCounters(&v1common.AssociatedQuery{}, nil)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(0))
	})
}
