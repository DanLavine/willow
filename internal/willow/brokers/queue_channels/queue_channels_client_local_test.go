package queuechannels

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	btreeonetomany "github.com/DanLavine/willow/internal/datastructures/btree_one_to_many"
	"github.com/DanLavine/willow/internal/willow/brokers/queue_channels/constructor"
	"github.com/DanLavine/willow/internal/willow/brokers/queue_channels/constructor/constructorfakes"
	"github.com/DanLavine/willow/pkg/clients/limiter_client/limiterclientfakes"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	. "github.com/onsi/gomega"
)

var checkTrue = true

func defaultEnqueueItem(g *GomegaWithT) *v1willow.EnqueueQueueItem {
	enqueuItem := &v1willow.EnqueueQueueItem{
		Item: []byte(`item to queue`),
		KeyValues: datatypes.KeyValues{
			"one": datatypes.Int(1),
		},
		Updateable:      true,
		RetryAttempts:   0,
		RetryPosition:   "front",
		TimeoutDuration: time.Second,
	}
	g.Expect(enqueuItem.Validate()).ToNot(HaveOccurred())

	return enqueuItem
}

func setupConstuctor(t *testing.T, g *GomegaWithT) (*gomock.Controller, constructor.QueueChannelsConstrutor) {
	// setup fake constructor
	mockController := gomock.NewController(t)

	// setup the limiter client to always pass
	fakeLimiterClient := limiterclientfakes.NewMockLimiterClient(mockController)
	fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter) error { return nil }).AnyTimes()
	fakeLimiterClient.EXPECT().SetCounters(gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter) error { return nil }).AnyTimes()

	constructor, err := constructor.NewQueueChannelConstructor("memory", fakeLimiterClient)
	g.Expect(err).ToNot(HaveOccurred())

	return mockController, constructor
}

func Test_queueChannelsClientLocal_EnqueueQueueItem(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context creating a channel", func(t *testing.T) {
		t.Run("It creates a new channel if the Name and KeyValues do not exists", func(t *testing.T) {
			// setup fake constructor
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			// setup fake queue channel
			mockQueueChannel := constructorfakes.NewMockQueueChannel(mockController)
			mockQueueChannel.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(*zap.Logger, *v1willow.EnqueueQueueItem) *errors.ServerError { return nil }).Times(1)
			mockQueueChannel.EXPECT().Dequeue().DoAndReturn(func() (channel <-chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())) {
				return nil
			}).Times(1)

			// setup fake constructor
			mockConstructor := constructorfakes.NewMockQueueChannelsConstrutor(mockController)
			mockConstructor.EXPECT().New(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ func(), _ string, _ datatypes.KeyValues) constructor.QueueChannel { return mockQueueChannel }).Times(1)

			// setup queue channel client local
			queueChannelClentLocal := NewLocalQueueChannelsClient(mockConstructor)

			err := queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", defaultEnqueueItem(g))
			g.Expect(err).ToNot(HaveOccurred())
		})

		t.Run("It creates a new channel if the different Name + same KeyValues", func(t *testing.T) {
			// setup fake constructor
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			// setup fake queue channel
			mockQueueChannel := constructorfakes.NewMockQueueChannel(mockController)
			mockQueueChannel.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(*zap.Logger, *v1willow.EnqueueQueueItem) *errors.ServerError { return nil }).Times(5)
			mockQueueChannel.EXPECT().Dequeue().DoAndReturn(func() (channel <-chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())) {
				return nil
			}).Times(5)

			// setup fake constructor
			mockConstructor := constructorfakes.NewMockQueueChannelsConstrutor(mockController)
			mockConstructor.EXPECT().New(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ func(), _ string, _ datatypes.KeyValues) constructor.QueueChannel { return mockQueueChannel }).Times(5)

			// setup queue channel client local
			queueChannelClentLocal := NewLocalQueueChannelsClient(mockConstructor)

			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name 1", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name 2", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name 3", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name 4", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name 5", defaultEnqueueItem(g)))
		})

		t.Run("It creates a new channel with the same Name + different KeyValues", func(t *testing.T) {
			// setup fake constructor
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			// setup fake queue channel
			mockQueueChannel := constructorfakes.NewMockQueueChannel(mockController)
			mockQueueChannel.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(*zap.Logger, *v1willow.EnqueueQueueItem) *errors.ServerError { return nil }).Times(5)
			mockQueueChannel.EXPECT().Dequeue().DoAndReturn(func() (channel <-chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())) {
				return nil
			}).Times(5)

			// setup fake constructor
			mockConstructor := constructorfakes.NewMockQueueChannelsConstrutor(mockController)
			mockConstructor.EXPECT().New(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ func(), _ string, _ datatypes.KeyValues) constructor.QueueChannel { return mockQueueChannel }).Times(5)

			// setup queue channel client local
			queueChannelClentLocal := NewLocalQueueChannelsClient(mockConstructor)

			for i := 0; i < 5; i++ {
				enqueuItem := &v1willow.EnqueueQueueItem{
					Item: []byte(`item to queue`),
					KeyValues: datatypes.KeyValues{
						"one": datatypes.Int(i),
					},
					Updateable:      true,
					RetryAttempts:   0,
					RetryPosition:   "front",
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueuItem.Validate()).ToNot(HaveOccurred())
				g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", enqueuItem))
			}
		})

		t.Run("It returns an error if the item cannot be enqueue", func(t *testing.T) {
			// setup fake constructor
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			// setup fake queue channel
			mockQueueChannel := constructorfakes.NewMockQueueChannel(mockController)
			mockQueueChannel.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(*zap.Logger, *v1willow.EnqueueQueueItem) *errors.ServerError {
				return &errors.ServerError{Message: "failed to enqueue item"}
			}).Times(1)

			// setup fake constructor
			mockConstructor := constructorfakes.NewMockQueueChannelsConstrutor(mockController)
			mockConstructor.EXPECT().New(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ func(), _ string, _ datatypes.KeyValues) constructor.QueueChannel { return mockQueueChannel }).Times(1)

			// setup queue channel client local
			queueChannelClentLocal := NewLocalQueueChannelsClient(mockConstructor)

			err := queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", defaultEnqueueItem(g))
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("failed to enqueue item"))
		})
	})

	t.Run("Context enquing an item to a channel that already exists", func(t *testing.T) {
		t.Run("It updates the channel items", func(t *testing.T) {
			// setup fake constructor
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			// setup fake queue channel
			mockQueueChannel := constructorfakes.NewMockQueueChannel(mockController)
			mockQueueChannel.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(*zap.Logger, *v1willow.EnqueueQueueItem) *errors.ServerError { return nil }).Times(5)
			mockQueueChannel.EXPECT().Dequeue().DoAndReturn(func() (channel <-chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())) {
				return nil
			}).Times(1)

			// setup fake constructor
			mockConstructor := constructorfakes.NewMockQueueChannelsConstrutor(mockController)
			mockConstructor.EXPECT().New(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ func(), _ string, _ datatypes.KeyValues) constructor.QueueChannel { return mockQueueChannel }).Times(1)

			// setup queue channel client local
			queueChannelClentLocal := NewLocalQueueChannelsClient(mockConstructor)

			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", defaultEnqueueItem(g)))
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", defaultEnqueueItem(g)))
		})
	})
}

// This could all be tested with a Mock, but I think it makes more sense to test with the
// actual memroy client right now to ensure all the logic works together
func Test_queueChannelsClientLocal_DequeueQueueItem(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context when a client is waiting", func(t *testing.T) {
		t.Run("It unblocks any requests on a server shutdown", func(t *testing.T) {
			mockController, constructor := setupConstuctor(t, g)
			defer mockController.Finish()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			done := make(chan struct{})
			var dequeueItem *v1willow.DequeueQueueItem
			var success func()
			var failure func()
			var dequeueErr *errors.ServerError

			// run the server in async mode
			executeCtx, executeCancel := context.WithCancel(context.Background())
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			go func() {
				defer close(done)
				dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), context.Background(), "test queue", query)
			}()

			g.Consistently(done).ShouldNot(BeClosed())

			executeCancel()
			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueItem).To(BeNil())
			g.Expect(success).To(BeNil())
			g.Expect(failure).To(BeNil())
			g.Expect(dequeueErr).ToNot(BeNil())
			g.Expect(dequeueErr.Error()).To(ContainSubstring("Server is shutting down. Retry the request"))
		})

		t.Run("It unblocks the request if the cancelContext is closed", func(t *testing.T) {
			mockController, constructor := setupConstuctor(t, g)
			defer mockController.Finish()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			done := make(chan struct{})
			var dequeueItem *v1willow.DequeueQueueItem
			var success func()
			var failure func()
			var dequeueErr *errors.ServerError

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				defer close(done)
				dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), ctx, "test queue", query)
			}()

			g.Consistently(done).ShouldNot(BeClosed())

			cancel()
			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueItem).To(BeNil())
			g.Expect(success).To(BeNil())
			g.Expect(failure).To(BeNil())
			g.Expect(dequeueErr).To(HaveOccurred())
			g.Expect(dequeueErr.Error()).To(Equal("Client closed"))
		})

		t.Run("It can dequeue an newly Enqueued item", func(t *testing.T) {
			mockController, constructor := setupConstuctor(t, g)
			defer mockController.Finish()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)
			// run the server in async mode
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// setup dequeue query
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			testZapCore, testLogs := observer.New(zap.DebugLevel)
			testLgger := zap.New(testZapCore)

			// run dequeue and wait till it has checked all current channels
			done := make(chan struct{})
			var dequeueItem *v1willow.DequeueQueueItem
			var success func()
			var failure func()
			var dequeueErr *errors.ServerError
			go func() {
				defer close(done)
				dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(testLgger, context.Background(), "test queue", query)
			}()
			g.Eventually(func() string {
				if testLogs.Len() == 0 {
					return ""
				}
				return testLogs.All()[testLogs.Len()-1].Message
			}).Should(ContainSubstring("waiting for available item"))

			// enqueue a new item
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", defaultEnqueueItem(g)))

			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueItem).ToNot(BeNil())
			g.Expect(dequeueItem.Item).To(Equal([]byte(`item to queue`)))
			g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
			g.Expect(dequeueItem.KeyValues).To(Equal(datatypes.KeyValues{"one": datatypes.Int(1)}))
			g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
			g.Expect(success).ToNot(BeNil())
			g.Expect(failure).ToNot(BeNil())
			g.Expect(dequeueErr).To(BeNil())

			// call success or fail
			success()
		})

		t.Run("It only dequeues an item that matches the original clients request", func(t *testing.T) {
			mockController, constructor := setupConstuctor(t, g)
			defer mockController.Finish()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)
			// run the server in async mode
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// setup dequeue query
			twoValue := datatypes.Int(2)
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
						"two": datatypes.Value{Value: &twoValue, ValueComparison: datatypes.EqualsPtr()},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			testZapCore, testLogs := observer.New(zap.DebugLevel)
			testLgger := zap.New(testZapCore)

			// run dequeue and wait till it has checked all current channels
			done := make(chan struct{})
			var dequeueItem *v1willow.DequeueQueueItem
			var success func()
			var failure func()
			var dequeueErr *errors.ServerError
			go func() {
				defer close(done)
				dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(testLgger, context.Background(), "test queue", query)
			}()
			g.Eventually(func() string {
				if testLogs.Len() == 0 {
					return ""
				}
				return testLogs.All()[testLogs.Len()-1].Message
			}).Should(ContainSubstring("waiting for available item"))

			// enqueue an item that triggers the request
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", defaultEnqueueItem(g)))
			g.Consistently(done).ShouldNot(BeClosed())

			// enqueue an item that should not trigger the request
			enqueueItem := &v1willow.EnqueueQueueItem{
				Item: []byte(`data to pull`),
				KeyValues: datatypes.KeyValues{
					"one":   datatypes.Float32(3.2),
					"two":   datatypes.Int(2),
					"three": datatypes.String("other"),
				},
				Updateable:      false,
				RetryAttempts:   2,
				RetryPosition:   "back",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", enqueueItem))

			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueItem).ToNot(BeNil())
			g.Expect(dequeueItem.Item).To(Equal([]byte(`data to pull`)))
			g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
			g.Expect(dequeueItem.KeyValues).To(Equal(datatypes.KeyValues{
				"one":   datatypes.Float32(3.2),
				"two":   datatypes.Int(2),
				"three": datatypes.String("other"),
			}))
			g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
			g.Expect(success).ToNot(BeNil())
			g.Expect(failure).ToNot(BeNil())
			g.Expect(dequeueErr).To(BeNil())

			// call success or fail
			success()
		})
	})

	t.Run("Context when items are enqued before a client is waiting", func(t *testing.T) {
		t.Run("It immediately returns a found item that matches the query", func(t *testing.T) {
			mockController, constructor := setupConstuctor(t, g)
			defer mockController.Finish()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)
			// run the server in async mode
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// enqueue an item
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", defaultEnqueueItem(g)))

			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			dequeueItem, success, failure, dequeueErr := queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), context.Background(), "test queue", query)
			g.Expect(dequeueItem).ToNot(BeNil())
			g.Expect(dequeueItem.Item).To(Equal([]byte(`item to queue`)))
			g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
			g.Expect(dequeueItem.KeyValues).To(Equal(datatypes.KeyValues{"one": datatypes.Int(1)}))
			g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
			g.Expect(success).ToNot(BeNil())
			g.Expect(failure).ToNot(BeNil())
			g.Expect(dequeueErr).To(BeNil())
		})

		t.Run("It immediately returns only a single item that matches the query", func(t *testing.T) {
			mockController, constructor := setupConstuctor(t, g)
			defer mockController.Finish()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)
			// run the server in async mode
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// enqueue an item that will dequeue
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", defaultEnqueueItem(g)))

			// enqueue an item that will not dequeue
			enqueueItem := &v1willow.EnqueueQueueItem{
				Item: []byte(`data to pull`),
				KeyValues: datatypes.KeyValues{
					"one":   datatypes.Int(1),
					"two":   datatypes.Int(2),
					"three": datatypes.String("other"),
				},
				Updateable:      false,
				RetryAttempts:   2,
				RetryPosition:   "back",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", enqueueItem))

			keyLimit := 1
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
					},
					Limits: &datatypes.KeyLimits{
						NumberOfKeys: &keyLimit,
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			dequeueItem, success, failure, dequeueErr := queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), context.Background(), "test queue", query)
			g.Expect(dequeueItem).ToNot(BeNil())
			g.Expect(dequeueItem.Item).To(Equal([]byte(`item to queue`)))
			g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
			g.Expect(dequeueItem.KeyValues).To(Equal(datatypes.KeyValues{"one": datatypes.Int(1)}))
			g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
			g.Expect(success).ToNot(BeNil())
			g.Expect(failure).ToNot(BeNil())
			g.Expect(dequeueErr).To(BeNil())

			// call success or fail
			success()
		})
	})

	t.Run("Context when the dequeue function fails to return anything", func(t *testing.T) {
		setupFailConstructor := func(g *GomegaWithT) (*gomock.Controller, constructor.QueueChannelsConstrutor, *constructorfakes.MockQueueChannel) {
			// setup fake constructor
			mockController := gomock.NewController(t)

			// setup the fake constructor and queue channel
			fakeQueueChannel := constructorfakes.NewMockQueueChannel(mockController)

			fakeConstructor := constructorfakes.NewMockQueueChannelsConstrutor(mockController)
			fakeConstructor.EXPECT().New(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ func(), _ string, _ datatypes.KeyValues) constructor.QueueChannel {
				return fakeQueueChannel
			}).AnyTimes()

			return mockController, fakeConstructor, fakeQueueChannel
		}

		t.Run("It continues to dequeue untill a different channels item is received", func(t *testing.T) {
			mockController, constructor, fakeQueueChannel := setupFailConstructor(g)
			defer mockController.Finish()

			// setup the channel
			count := 0
			dequeueChanOne := make(chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func()))
			dequeueChanTwo := make(chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func()))

			fakeQueueChannel.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(*zap.Logger, *v1willow.EnqueueQueueItem) *errors.ServerError { return nil }).AnyTimes()
			fakeQueueChannel.EXPECT().Dequeue().DoAndReturn(func() <-chan (func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())) {
				// IMPORTANT TO HAVE THIS BE 2. the enqueue calls dequeue 1 time each to update any clients currently waiting
				if count <= 2 {
					count++
					return dequeueChanOne
				}
				return dequeueChanTwo
			}).AnyTimes()
			fakeQueueChannel.EXPECT().Execute(gomock.Any()).DoAndReturn(func(ctx context.Context) error { return nil }).AnyTimes()

			// setup the response for the dequeue chan
			go func() {
				// first reponse is all nil (mimic the limiter blocking a request)
				dequeueChanOne <- func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func()) { return nil, nil, nil }

				// second response over same channel now should pass
				dequeueChanTwo <- func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func()) {
					return &v1willow.DequeueQueueItem{
						KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Float32(2.0)},
						ItemID:          "something",
						Item:            []byte(`doesn't matter 2`),
						TimeoutDuration: time.Second,
					}, func() {}, func() {}
				}
			}()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

			// enqueue our fake chan twice
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", &v1willow.EnqueueQueueItem{
				Item:            []byte(`doesn't matter 1`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
				Updateable:      false,
				RetryAttempts:   0,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			})).ToNot(HaveOccurred())

			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", &v1willow.EnqueueQueueItem{
				Item:            []byte(`doesn't matter 2`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Float32(2.0)},
				Updateable:      false,
				RetryAttempts:   0,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			})).ToNot(HaveOccurred())

			// run the server in async mode
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// setup dequeue query
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			// run dequeue and wait till it has checked all current channels
			done := make(chan struct{})
			var dequeueItem *v1willow.DequeueQueueItem
			var success func()
			var failure func()
			var dequeueErr *errors.ServerError
			go func() {
				defer close(done)
				dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), context.Background(), "test queue", query)
			}()

			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueItem).ToNot(BeNil())
			g.Expect(dequeueItem.Item).To(Equal([]byte(`doesn't matter 2`)))
			g.Expect(dequeueItem.ItemID).To(Equal("something"))
			g.Expect(dequeueItem.KeyValues).To(Equal(datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Float32(2.0)}))
			g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
			g.Expect(success).ToNot(BeNil())
			g.Expect(failure).ToNot(BeNil())
			g.Expect(dequeueErr).To(BeNil())

			// call success or fail
			success()
		})

		t.Run("It continues to dequeue untill the failed channel eventually returns an item", func(t *testing.T) {
			mockController, constructor, fakeQueueChannel := setupFailConstructor(g)
			defer mockController.Finish()

			// setup the channel
			dequeueChan := make(chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func()))
			fakeQueueChannel.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(*zap.Logger, *v1willow.EnqueueQueueItem) *errors.ServerError { return nil }).Times(1)
			fakeQueueChannel.EXPECT().Dequeue().DoAndReturn(func() <-chan (func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())) {
				return dequeueChan
			}).AnyTimes()
			fakeQueueChannel.EXPECT().Execute(gomock.Any()).DoAndReturn(func(ctx context.Context) error { return nil }).Times(1)

			// setup the response for the dequeue chan
			go func() {
				// first reponse is all nil (mimic the limiter blocking a request)
				empty := func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func()) { return nil, nil, nil }
				dequeueChan <- empty

				// second response over same channel now should pass
				item := func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func()) {
					return &v1willow.DequeueQueueItem{
						KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
						ItemID:          "something",
						Item:            []byte(`data`),
						TimeoutDuration: time.Second,
					}, func() {}, func() {}
				}
				dequeueChan <- item
			}()

			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

			// enqueue our fake chan
			g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "test queue", &v1willow.EnqueueQueueItem{
				Item:            []byte(`doesn't matter`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
				Updateable:      false,
				RetryAttempts:   0,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			})).ToNot(HaveOccurred())

			// run the server in async mode
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// setup dequeue query
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{
					KeyValues: map[string]datatypes.Value{
						"one": datatypes.Value{Exists: &checkTrue},
					},
				},
			}
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			// run dequeue and wait till it has checked all current channels
			done := make(chan struct{})
			var dequeueItem *v1willow.DequeueQueueItem
			var success func()
			var failure func()
			var dequeueErr *errors.ServerError
			go func() {
				defer close(done)
				dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), context.Background(), "test queue", query)
			}()

			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueItem).ToNot(BeNil())
			g.Expect(dequeueItem.Item).To(Equal([]byte(`data`)))
			g.Expect(dequeueItem.ItemID).To(Equal("something"))
			g.Expect(dequeueItem.KeyValues).To(Equal(datatypes.KeyValues{"one": datatypes.Int(1)}))
			g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
			g.Expect(success).ToNot(BeNil())
			g.Expect(failure).ToNot(BeNil())
			g.Expect(dequeueErr).To(BeNil())

			// call success or fail
			success()
		})
	})
}

func Test_queueChannelsClientLocal_ACK(t *testing.T) {
	g := NewGomegaWithT(t)

	setupConstuctor := func(g *GomegaWithT) (*gomock.Controller, constructor.QueueChannelsConstrutor) {
		// setup fake constructor
		mockController := gomock.NewController(t)

		// setup the limiter client to always pass
		fakeLimiterClient := limiterclientfakes.NewMockLimiterClient(mockController)
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter) error { return nil }).AnyTimes()

		constructor, err := constructor.NewQueueChannelConstructor("memory", fakeLimiterClient)
		g.Expect(err).ToNot(HaveOccurred())

		return mockController, constructor
	}

	t.Run("It returns an error if the channel KeyValues cannot be found", func(t *testing.T) {
		mockController, constructor := setupConstuctor(g)
		defer mockController.Finish()

		ack := &v1willow.ACK{
			ItemID: "some id",
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Passed: true,
		}
		g.Expect(ack.Validate()).ToNot(HaveOccurred())

		queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

		err := queueChannelClentLocal.ACK(zap.NewNop(), "not found", ack)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Failed to find channel by key values"))
	})

	t.Run("Context when an item exists to be acked", func(t *testing.T) {
		setupQueueChannelClient := func(g *GomegaWithT) (*gomock.Controller, *queueChannelsClientLocal) {
			mockController, constructor := setupConstuctor(g)
			queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

			return mockController, queueChannelClentLocal
		}

		enqueueItem := func(g *GomegaWithT, queueChannelClentLocal *queueChannelsClientLocal) {
			enqueuItem := &v1willow.EnqueueQueueItem{
				Item: []byte(`item to queue`),
				KeyValues: datatypes.KeyValues{
					"one": datatypes.Int(1),
				},
				Updateable:      false,
				RetryAttempts:   0,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueuItem.Validate()).ToNot(HaveOccurred())

			err := queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", enqueuItem)
			g.Expect(err).ToNot(HaveOccurred())
		}

		dequeueItem := func(g *GomegaWithT, queueChannelClentLocal *queueChannelsClientLocal) *v1willow.DequeueQueueItem {
			query := datatypes.AssociatedKeyValuesQuery{} // select all
			g.Expect(query.Validate()).ToNot(HaveOccurred())

			done := make(chan struct{})
			var dequeueItem *v1willow.DequeueQueueItem
			var success func()
			var failure func()
			var dequeueErr *errors.ServerError
			go func() {
				defer close(done)
				dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), context.Background(), "queue name", query)
			}()

			g.Eventually(done).Should(BeClosed())
			g.Expect(dequeueItem).ToNot(BeNil())
			g.Expect(success).ToNot(BeNil())
			g.Expect(failure).ToNot(BeNil())
			g.Expect(dequeueErr).To(BeNil())

			// call successful dequeue
			success()

			return dequeueItem
		}

		t.Run("It destroys the channel if it is the last item deleted", func(t *testing.T) {
			mockController, queueChannelClentLocal := setupQueueChannelClient(g)
			defer mockController.Finish()

			// run the queue channel client async
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// enqueue a single item
			enqueueItem(g, queueChannelClentLocal)

			// dequeue the single item
			item := dequeueItem(g, queueChannelClentLocal)

			// ack the item dequeued with true
			ack := &v1willow.ACK{
				ItemID:    item.ItemID,
				KeyValues: item.KeyValues,
				Passed:    true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queueChannelClentLocal.ACK(zap.NewNop(), "queue name", ack)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the channel is eventually deleted.
			g.Eventually(func() bool {
				foundAny := false
				onIterate := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
					foundAny = true
					return true
				}
				queueChannelClentLocal.queueChannels.Query("queue name", datatypes.AssociatedKeyValuesQuery{}, onIterate)

				return foundAny
			}).Should(BeFalse())
		})

		t.Run("It does not destroy the channel if there is another item enqueued", func(t *testing.T) {
			mockController, queueChannelClentLocal := setupQueueChannelClient(g)
			defer mockController.Finish()

			// run the queue channel client async
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// enqueue multiple items
			enqueueItem(g, queueChannelClentLocal)
			enqueueItem(g, queueChannelClentLocal)

			// dequeue the single item
			item := dequeueItem(g, queueChannelClentLocal)

			// ack the item dequeued with true
			ack := &v1willow.ACK{
				ItemID:    item.ItemID,
				KeyValues: item.KeyValues,
				Passed:    true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queueChannelClentLocal.ACK(zap.NewNop(), "queue name", ack)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the channel is not deleted
			g.Consistently(func() bool {
				foundAny := false
				onIterate := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
					foundAny = true
					return true
				}
				queueChannelClentLocal.queueChannels.Query("queue name", datatypes.AssociatedKeyValuesQuery{}, onIterate)

				return foundAny
			}).Should(BeTrue())
		})

		t.Run("It does not destroy the channel if there is another item processing", func(t *testing.T) {
			mockController, queueChannelClentLocal := setupQueueChannelClient(g)
			defer mockController.Finish()

			// run the queue channel client async
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// enqueue multiple items
			enqueueItem(g, queueChannelClentLocal)
			enqueueItem(g, queueChannelClentLocal)

			// dequeue the single item
			item1 := dequeueItem(g, queueChannelClentLocal)
			item2 := dequeueItem(g, queueChannelClentLocal)

			// ack the item dequeued with true
			ack := &v1willow.ACK{
				ItemID:    item1.ItemID,
				KeyValues: item1.KeyValues,
				Passed:    true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err := queueChannelClentLocal.ACK(zap.NewNop(), "queue name", ack)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the channel is not deleted
			g.Consistently(func() bool {
				foundAny := false
				onIterate := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
					foundAny = true
					return true
				}
				queueChannelClentLocal.queueChannels.Query("queue name", datatypes.AssociatedKeyValuesQuery{}, onIterate)

				return foundAny
			}).Should(BeTrue())

			// ack the second item dequeued with true
			ack2 := &v1willow.ACK{
				ItemID:    item2.ItemID,
				KeyValues: item2.KeyValues,
				Passed:    true,
			}
			g.Expect(ack.Validate()).ToNot(HaveOccurred())

			err = queueChannelClentLocal.ACK(zap.NewNop(), "queue name", ack2)
			g.Expect(err).ToNot(HaveOccurred())

			// ensure the channel is eventually deleted
			g.Eventually(func() bool {
				foundAny := false
				onIterate := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
					foundAny = true
					return true
				}
				queueChannelClentLocal.queueChannels.Query("queue name", datatypes.AssociatedKeyValuesQuery{}, onIterate)

				return foundAny
			}).Should(BeFalse())
		})
	})
}

func Test_queueChannelsClientLocal_Heartbeat(t *testing.T) {
	g := NewGomegaWithT(t)

	setupConstuctor := func(g *GomegaWithT) (*gomock.Controller, constructor.QueueChannelsConstrutor) {
		// setup fake constructor
		mockController := gomock.NewController(t)

		// setup the limiter client to always pass
		fakeLimiterClient := limiterclientfakes.NewMockLimiterClient(mockController)
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter) error { return nil }).AnyTimes()
		fakeLimiterClient.EXPECT().SetCounters(gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter) error { return nil }).AnyTimes()

		constructor, err := constructor.NewQueueChannelConstructor("memory", fakeLimiterClient)
		g.Expect(err).ToNot(HaveOccurred())

		return mockController, constructor
	}

	setupQueueChannelClient := func(g *GomegaWithT) (*gomock.Controller, *queueChannelsClientLocal) {
		mockController, constructor := setupConstuctor(g)

		// setup queue channel client local
		queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

		return mockController, queueChannelClentLocal
	}

	enqueueItem := func(g *GomegaWithT, queueChannelClentLocal *queueChannelsClientLocal) {
		enqueuItem := &v1willow.EnqueueQueueItem{
			Item: []byte(`item to queue`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      false,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueuItem.Validate()).ToNot(HaveOccurred())

		err := queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", enqueuItem)
		g.Expect(err).ToNot(HaveOccurred())
	}

	dequeueItem := func(g *GomegaWithT, queueChannelClentLocal *queueChannelsClientLocal) *v1willow.DequeueQueueItem {
		query := datatypes.AssociatedKeyValuesQuery{} // select all
		g.Expect(query.Validate()).ToNot(HaveOccurred())

		done := make(chan struct{})
		var dequeueItem *v1willow.DequeueQueueItem
		var success func()
		var failure func()
		var dequeueErr *errors.ServerError
		go func() {
			defer close(done)
			dequeueItem, success, failure, dequeueErr = queueChannelClentLocal.DequeueQueueItem(zap.NewNop(), context.Background(), "queue name", query)
		}()

		g.Eventually(done, 2*time.Second).Should(BeClosed())
		g.Expect(dequeueItem).ToNot(BeNil())
		g.Expect(success).ToNot(BeNil())
		g.Expect(failure).ToNot(BeNil())
		g.Expect(dequeueErr).To(BeNil())

		// call successful dequeue
		success()

		return dequeueItem
	}

	t.Run("It returns an error if the ItemID cannot be found", func(t *testing.T) {
		mockController, constructor := setupConstuctor(g)
		defer mockController.Finish()

		hearbeat := &v1willow.Heartbeat{
			ItemID: "not found",
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
		}
		g.Expect(hearbeat.Validate()).ToNot(HaveOccurred())

		queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

		err := queueChannelClentLocal.Heartbeat(zap.NewNop(), "queue name", hearbeat)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Failed to find channel for item by key values"))
	})

	t.Run("It can continue to heartbeat an item that has not been acked or timed out", func(t *testing.T) {
		mockController, queueChannelClentLocal := setupQueueChannelClient(g)
		defer mockController.Finish()

		// run the queue channel client async
		executeCtx, executeCancel := context.WithCancel(context.Background())
		defer executeCancel()
		go func() {
			_ = queueChannelClentLocal.Execute(executeCtx)
		}()

		// enqueue a single item
		enqueueItem(g, queueChannelClentLocal)

		// dequeue the single item
		item := dequeueItem(g, queueChannelClentLocal)

		// heartbeat the item
		hearbeat := &v1willow.Heartbeat{
			ItemID:    item.ItemID,
			KeyValues: item.KeyValues,
		}
		g.Expect(hearbeat.Validate()).ToNot(HaveOccurred())

		for i := 0; i < 5; i++ {
			err := queueChannelClentLocal.Heartbeat(zap.NewNop(), "queue name", hearbeat)
			g.Expect(err).ToNot(HaveOccurred())

			time.Sleep(300 * time.Millisecond)
		}

		// can still ack the item
		ack := &v1willow.ACK{
			ItemID:    item.ItemID,
			KeyValues: item.KeyValues,
			Passed:    true,
		}
		g.Expect(ack.Validate()).ToNot(HaveOccurred())

		err := queueChannelClentLocal.ACK(zap.NewNop(), "queue name", ack)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("Context when an item has timed out", func(t *testing.T) {
		t.Run("It destroys the channel if it is the only item", func(t *testing.T) {
			mockController, queueChannelClentLocal := setupQueueChannelClient(g)
			defer mockController.Finish()

			// run the queue channel client async
			executeCtx, executeCancel := context.WithCancel(context.Background())
			defer executeCancel()
			go func() {
				_ = queueChannelClentLocal.Execute(executeCtx)
			}()

			// enqueue a single item
			enqueueItem(g, queueChannelClentLocal)

			// dequeue the single item and let it time out twice to trigger the retry
			g.Eventually(func() *v1willow.DequeueQueueItem { return dequeueItem(g, queueChannelClentLocal) }, 2*time.Second).ShouldNot(BeNil())
			g.Eventually(func() *v1willow.DequeueQueueItem { return dequeueItem(g, queueChannelClentLocal) }, 2*time.Second).ShouldNot(BeNil())

			// ensure channel is destroyed
			g.Eventually(func() int {
				foundItems := 0
				queueChannelClentLocal.queueChannels.Query("queue name", datatypes.AssociatedKeyValuesQuery{}, func(oneToManyItem btreeonetomany.OneToManyItem) bool {
					foundItems++
					return true
				})

				return foundItems
			}, 2*time.Second).Should(Equal(0))
		})

		t.Run("It errors trying to heartbeat an item that has timed out", func(t *testing.T) {
			// This does not work as expected. I want to ensure a new itemID is returned for each dequeue
		})
	})
}

func Test_queueChannelsClientLocal_DestroyChannelsForQueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It performs a no-op if the channel name does not exist", func(t *testing.T) {
		mockController, constructor := setupConstuctor(t, g)
		defer mockController.Finish()

		queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

		err := queueChannelClentLocal.DestroyChannelsForQueue(zap.NewNop(), "does not matter")
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It destroys all channels associated with the queue", func(t *testing.T) {
		mockController, constructor := setupConstuctor(t, g)
		defer mockController.Finish()

		queueChannelClentLocal := NewLocalQueueChannelsClient(constructor)

		// enqueue a number of items all to 1 queue
		for i := 0; i < 10; i++ {
			enqueuItem := &v1willow.EnqueueQueueItem{
				Item: []byte(`item to queue`),
				KeyValues: datatypes.KeyValues{
					fmt.Sprintf("%d", i): datatypes.Int(i),
				},
				Updateable:      false,
				RetryAttempts:   1,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueuItem.Validate()).ToNot(HaveOccurred())

			err := queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name", enqueuItem)
			g.Expect(err).ToNot(HaveOccurred())
		}

		// enque another item to another queue
		enqueuItem := &v1willow.EnqueueQueueItem{
			Item: []byte(`item to queue`),
			KeyValues: datatypes.KeyValues{
				"one": datatypes.Int(1),
			},
			Updateable:      false,
			RetryAttempts:   1,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueuItem.Validate()).ToNot(HaveOccurred())
		g.Expect(queueChannelClentLocal.EnqueueQueueItem(zap.NewNop(), "queue name 2", enqueuItem)).ToNot(HaveOccurred())

		err := queueChannelClentLocal.DestroyChannelsForQueue(zap.NewNop(), "queue name")
		g.Expect(err).ToNot(HaveOccurred())

		// should only have 1 item left in the second queue's channels
		foundItems := 0
		queueChannelClentLocal.queueChannels.Query("queue name", datatypes.AssociatedKeyValuesQuery{}, func(oneToManyItem btreeonetomany.OneToManyItem) bool {
			foundItems++
			return true
		})
		g.Expect(foundItems).To(Equal(0))

		queueChannelClentLocal.queueChannels.Query("queue name 2", datatypes.AssociatedKeyValuesQuery{}, func(oneToManyItem btreeonetomany.OneToManyItem) bool {
			foundItems++
			return true
		})
		g.Expect(foundItems).To(Equal(1))
	})
}
