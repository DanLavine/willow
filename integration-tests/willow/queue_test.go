package willow_integration_tests

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	willowclient "github.com/DanLavine/willow/pkg/clients/willow_client"
	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	. "github.com/DanLavine/willow/integration-tests/integrationhelpers"
	. "github.com/onsi/gomega"
)

func setupLimitterClient(g *GomegaWithT, url string) limiterclient.LimiterClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:           url,
		CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
	}

	limiterClient, err := limiterclient.NewLimiterClient(cfg)
	g.Expect(err).ToNot(HaveOccurred())

	return limiterClient
}

func setupWillowClient(g *GomegaWithT, url string) willowclient.WillowServiceClient {
	_, currentDir, _, _ := runtime.Caller(0)

	cfg := &clients.Config{
		URL:           url,
		CAFile:        filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "ca.crt"),
		ClientKeyFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.key"),
		ClientCRTFile: filepath.Join(currentDir, "..", "..", "..", "testhelpers", "tls-keys", "client.crt"),
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

		createQueue := &v1willow.Queue{
			Spec: &v1willow.QueueSpec{
				DBDefinition: &v1willow.QueueDBDefinition{
					Name: helpers.PointerOf[string]("test queue"),
				},
				Properties: &v1willow.QueueProperties{
					MaxItems: helpers.PointerOf[int64](5),
				},
			},
		}

		err := willowClient.CreateQueue(context.Background(), createQueue)

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
			createQueue := &v1willow.Queue{
				Spec: &v1willow.QueueSpec{
					DBDefinition: &v1willow.QueueDBDefinition{
						Name: helpers.PointerOf[string](fmt.Sprintf("test queue %d", i)),
					},
					Properties: &v1willow.QueueProperties{
						MaxItems: helpers.PointerOf[int64](5),
					},
				},
			}

			err := willowClient.CreateQueue(context.Background(), createQueue)
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
			createQueue := &v1willow.Queue{
				Spec: &v1willow.QueueSpec{
					DBDefinition: &v1willow.QueueDBDefinition{
						Name: helpers.PointerOf[string](fmt.Sprintf("test queue %d", i)),
					},
					Properties: &v1willow.QueueProperties{
						MaxItems: helpers.PointerOf[int64](5),
					},
				},
			}

			err := willowClient.CreateQueue(context.Background(), createQueue)
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

		createQueue := &v1willow.Queue{
			Spec: &v1willow.QueueSpec{
				DBDefinition: &v1willow.QueueDBDefinition{
					Name: helpers.PointerOf[string]("test queue"),
				},
				Properties: &v1willow.QueueProperties{
					MaxItems: helpers.PointerOf[int64](5),
				},
			},
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		queue, err := willowClient.GetQueue(context.Background(), "test queue", &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(*queue.Spec.DBDefinition.Name).To(Equal("test queue"))
		g.Expect(*queue.Spec.Properties.MaxItems).To(Equal(int64(5)))
		g.Expect(queue.State.Deleting).To(BeFalse())
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
		createQueue := &v1willow.Queue{
			Spec: &v1willow.QueueSpec{
				DBDefinition: &v1willow.QueueDBDefinition{
					Name: helpers.PointerOf[string]("test queue"),
				},
				Properties: &v1willow.QueueProperties{
					MaxItems: helpers.PointerOf[int64](5),
				},
			},
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// update
		updateQueue := &v1willow.QueueProperties{
			MaxItems: helpers.PointerOf[int64](12),
		}
		err := willowClient.UpdateQueue(context.Background(), "test queue", updateQueue)
		g.Expect(err).ToNot(HaveOccurred())

		// get
		queue, err := willowClient.GetQueue(context.Background(), "test queue", &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(*queue.Spec.DBDefinition.Name).To(Equal("test queue"))
		g.Expect(*queue.Spec.Properties.MaxItems).To(Equal(int64(12)))
		g.Expect(queue.State.Deleting).To(BeFalse())
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
		createQueue := &v1willow.Queue{
			Spec: &v1willow.QueueSpec{
				DBDefinition: &v1willow.QueueDBDefinition{
					Name: helpers.PointerOf[string]("test queue"),
				},
				Properties: &v1willow.QueueProperties{
					MaxItems: helpers.PointerOf[int64](5),
				},
			},
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// ensure the override exists before deletion
		overrides, err := limiterClient.QueryOverrides(context.Background(), "_willow_queue_enqueued_limits", &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(1))

		// delete
		err = willowClient.DeleteQueue(context.Background(), "test queue")
		g.Expect(err).ToNot(HaveOccurred())

		fmt.Println(willowTestConstruct.ServerStdout.String())
		// get
		queue, err := willowClient.GetQueue(context.Background(), "test queue", &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(queue).To(BeNil())
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to find queue 'test queue' by name"))

		// ensure the override is deleted
		overrides, err = limiterClient.QueryOverrides(context.Background(), "_willow_queue_enqueued_limits", &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(overrides)).To(Equal(0))
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
		counters, err := limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(0))

		// setup queue
		createQueue := &v1willow.Queue{
			Spec: &v1willow.QueueSpec{
				DBDefinition: &v1willow.QueueDBDefinition{
					Name: helpers.PointerOf[string]("test queue"),
				},
				Properties: &v1willow.QueueProperties{
					MaxItems: helpers.PointerOf[int64](5),
				},
			},
		}
		g.Expect(willowClient.CreateQueue(context.Background(), createQueue)).ToNot(HaveOccurred())

		// enqueue multiple item
		enqueueQueueItem1 := &v1willow.Item{
			Spec: &v1willow.ItemSpec{
				DBDefinition: &v1willow.ItemDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"one": datatypes.Int(1),
					},
				},
				Properties: &v1willow.ItemProperties{
					Data:            []byte(`data for first item`),
					Updateable:      helpers.PointerOf(true),
					RetryAttempts:   helpers.PointerOf[uint64](1),
					RetryPosition:   helpers.PointerOf("front"),
					TimeoutDuration: helpers.PointerOf(5 * time.Second),
				},
			},
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem1)).ToNot(HaveOccurred())

		enqueueQueueItem2 := &v1willow.Item{ // updates the previous item
			Spec: &v1willow.ItemSpec{
				DBDefinition: &v1willow.ItemDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"one": datatypes.Int(1),
					},
				},
				Properties: &v1willow.ItemProperties{
					Data:            []byte(`data for second item`),
					Updateable:      helpers.PointerOf(true),
					RetryAttempts:   helpers.PointerOf[uint64](1),
					RetryPosition:   helpers.PointerOf("front"),
					TimeoutDuration: helpers.PointerOf(5 * time.Second),
				},
			},
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem2)).ToNot(HaveOccurred())

		enqueueQueueItem3 := &v1willow.Item{
			Spec: &v1willow.ItemSpec{
				DBDefinition: &v1willow.ItemDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"one": datatypes.Int(1),
						"two": datatypes.Int(2),
					},
				},
				Properties: &v1willow.ItemProperties{
					Data:            []byte(`data for third item`),
					Updateable:      helpers.PointerOf(true),
					RetryAttempts:   helpers.PointerOf[uint64](1),
					RetryPosition:   helpers.PointerOf("front"),
					TimeoutDuration: helpers.PointerOf(5 * time.Second),
				},
			},
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem3)).ToNot(HaveOccurred())

		enqueueQueueItem4 := &v1willow.Item{
			Spec: &v1willow.ItemSpec{
				DBDefinition: &v1willow.ItemDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"one":   datatypes.Int(1),
						"two":   datatypes.Int(2),
						"three": datatypes.String("3"),
					},
				},
				Properties: &v1willow.ItemProperties{
					Data:            []byte(`data for fourth item`),
					Updateable:      helpers.PointerOf(false),
					RetryAttempts:   helpers.PointerOf[uint64](1),
					RetryPosition:   helpers.PointerOf("front"),
					TimeoutDuration: helpers.PointerOf(5 * time.Second),
				},
			},
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem4)).ToNot(HaveOccurred())

		enqueueQueueItem5 := &v1willow.Item{
			Spec: &v1willow.ItemSpec{
				DBDefinition: &v1willow.ItemDBDefinition{
					KeyValues: dbdefinition.TypedKeyValues{
						"one":   datatypes.Int(1),
						"two":   datatypes.Int(2),
						"three": datatypes.String("3"),
					},
				},
				Properties: &v1willow.ItemProperties{
					Data:            []byte(`data for fifth item`),
					Updateable:      helpers.PointerOf(false),
					RetryAttempts:   helpers.PointerOf[uint64](1),
					RetryPosition:   helpers.PointerOf("front"),
					TimeoutDuration: helpers.PointerOf(5 * time.Second),
				},
			},
		}
		g.Expect(willowClient.EnqueueQueueItem(context.Background(), "test queue", enqueueQueueItem5)).ToNot(HaveOccurred())

		// delete a channel
		err = willowClient.DeleteQueue(context.Background(), "test queue")
		g.Expect(err).ToNot(HaveOccurred())

		// ensure the counters are destroyed properly
		counters, err = limiterClient.QueryCounters(context.Background(), &queryassociatedaction.AssociatedActionQuery{})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(counters)).To(Equal(0))
	})
}
