package memory

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	fakelimiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client/limiterclientfakes"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
	"github.com/DanLavine/willow/testhelpers"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	. "github.com/onsi/gomega"
)

func defaultKeyValues(g *GomegaWithT) datatypes.KeyValues {
	keyValues := datatypes.KeyValues{
		"one": datatypes.Int(1),
	}

	g.Expect(keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType)).ToNot(HaveOccurred())
	return keyValues
}

func setupSuccessFakeLimiter(t *testing.T, count int) (*gomock.Controller, *fakelimiterclient.MockLimiterClient) {
	mockController := gomock.NewController(t)
	mockClient := fakelimiterclient.NewMockLimiterClient(mockController)

	mockClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).Times(count)
	mockClient.EXPECT().SetCounters(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).AnyTimes()

	return mockController, mockClient
}

func fakeLimiterClient(t *testing.T) (*gomock.Controller, *fakelimiterclient.MockLimiterClient) {
	mockController := gomock.NewController(t)
	return mockController, fakelimiterclient.NewMockLimiterClient(mockController)
}

func Test_memoryQueueChannel_Delete(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns true if there are no enqueued items and closes the dequeue chan", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		go func() {
			_ = memeoryQueueChannel.Execute(context.Background())
		}()

		successfulDelte := memeoryQueueChannel.Delete()
		g.Expect(successfulDelte).To(BeTrue())
		g.Eventually(memeoryQueueChannel.Dequeue()).Should(BeClosed())
	})

	t.Run("It returns false if there are enqueued items and does not closes the dequeue chan", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = memeoryQueueChannel.Execute(ctx)
		}()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).Times(1)

		// create and enqueue an item
		enqueueItem := &v1willow.EnqueueQueueItem{
			Item:            []byte(`data`),
			KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
			Updateable:      true,
			RetryAttempts:   2,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

		err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
		g.Expect(err).ToNot(HaveOccurred())

		successfulDelte := memeoryQueueChannel.Delete()
		g.Expect(successfulDelte).To(BeFalse())
		g.Consistently(memeoryQueueChannel.Dequeue()).ShouldNot(BeClosed())
	})
}

func Test_memoryQueueChannel_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can add a new item to the queue", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).Times(1)

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// create and enqueue the item
		enqueueItem := &v1willow.EnqueueQueueItem{
			Item:            []byte(`data`),
			KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
			Updateable:      true,
			RetryAttempts:   2,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

		err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("It updates the last item in the queue if it is not yet processing and updateable", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).Times(1)

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// create and enqueue the item
		enqueueItem := &v1willow.EnqueueQueueItem{
			Item:            []byte(`data`),
			KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
			Updateable:      true,
			RetryAttempts:   2,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

		err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
		g.Expect(err).ToNot(HaveOccurred())

		// update the first item in the queue
		enqueueItem2 := &v1willow.EnqueueQueueItem{
			Item:            []byte(`data 2`),
			KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
			Updateable:      true,
			RetryAttempts:   0,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueueItem2.Validate()).ToNot(HaveOccurred())

		err = memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem2)
		g.Expect(err).ToNot(HaveOccurred())

		// check the available items len
		g.Expect(len(memeoryQueueChannel.itemIDsEnqueued)).To(Equal(1))
	})

	t.Run("It appends the last item in the queue if it is not yet processing and not updateable", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).Times(2)

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// create and enqueue the item
		enqueueItem := &v1willow.EnqueueQueueItem{
			Item:            []byte(`data`),
			KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
			Updateable:      false,
			RetryAttempts:   2,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

		err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
		g.Expect(err).ToNot(HaveOccurred())

		// update the first item in the queue
		enqueueItem2 := &v1willow.EnqueueQueueItem{
			Item:            []byte(`data 2`),
			KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
			Updateable:      true,
			RetryAttempts:   0,
			RetryPosition:   "front",
			TimeoutDuration: time.Second,
		}
		g.Expect(enqueueItem2.Validate()).ToNot(HaveOccurred())

		err = memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem2)
		g.Expect(err).ToNot(HaveOccurred())

		// check the available items len
		g.Expect(len(memeoryQueueChannel.itemIDsEnqueued)).To(Equal(2))
	})

	t.Run("Context when the limits are already reached", func(t *testing.T) {
		t.Run("It returns an error", func(t *testing.T) {
			mockController, fakeLimiterClient := fakeLimiterClient(t)
			defer mockController.Finish()

			// set limiter client to fail
			fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return fmt.Errorf("failed to update counter") }).Times(1)

			// create queue channel
			memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

			// create and enqueue the item
			enqueueItem := &v1willow.EnqueueQueueItem{
				Item:            []byte(`data`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
				Updateable:      true,
				RetryAttempts:   2,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

			err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("Queue has reached the total number of allowed queue items"))
		})

		t.Run("It can update the last item already in the queue without countng towards the limit", func(t *testing.T) {
			mockController, fakeLimiterClient := fakeLimiterClient(t)
			defer mockController.Finish()

			// set limiter client to fail
			counter := 0
			fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
				if counter == 0 {
					counter++
					return nil
				}
				return fmt.Errorf("failed to update counter")
			}).Times(1)

			// create queue channel
			memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

			// create and enqueue the item
			enqueueItem := &v1willow.EnqueueQueueItem{
				Item:            []byte(`data`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
				Updateable:      true,
				RetryAttempts:   2,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

			err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
			g.Expect(err).ToNot(HaveOccurred())

			// update the first item in the queue
			enqueueItem2 := &v1willow.EnqueueQueueItem{
				Item:            []byte(`data 2`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
				Updateable:      true,
				RetryAttempts:   0,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueueItem2.Validate()).ToNot(HaveOccurred())

			err = memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem2)
			g.Expect(err).ToNot(HaveOccurred())
		})
	})
}

func Test_memoryQueueChannel_Dequeue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It closes the dequeue chan when Execute stops processing", func(t *testing.T) {
		// set limiter client to pass
		mockController, fakeLimiterClient := setupSuccessFakeLimiter(t, 0)
		defer mockController.Finish()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			_ = memeoryQueueChannel.Execute(ctx)
		}()

		// grab the dequeu channel
		dequeueChan := memeoryQueueChannel.Dequeue()
		g.Consistently(dequeueChan).ShouldNot(Receive())

		// close the deueChan by stopping taskmanager
		cancel()
		g.Eventually(dequeueChan).Should(BeClosed())
	})

	t.Run("It closes the deueue chan and stops Executing when ForceDelete is called", func(t *testing.T) {
		// set limiter client to pass
		mockController, fakeLimiterClient := setupSuccessFakeLimiter(t, 0)
		defer mockController.Finish()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		go func() {
			_ = memeoryQueueChannel.Execute(context.Background())
		}()

		// grab the dequeu channel
		dequeueChan := memeoryQueueChannel.Dequeue()
		g.Consistently(dequeueChan).ShouldNot(Receive())

		// close the deueChan by stopping taskmanager
		memeoryQueueChannel.ForceDelete(testhelpers.NewContextWithLogger())
		g.Eventually(dequeueChan).Should(BeClosed())
	})

	t.Run("Context when the dequeue callback function succeeds", func(t *testing.T) {
		t.Run("Context successful dequeues", func(t *testing.T) {
			t.Run("It can dequeue only a single item untill success is called", func(t *testing.T) {
				// set limiter client to pass
				mockController, fakeLimiterClient := setupSuccessFakeLimiter(t, 7) // 5 for enqueue, 2 for success
				defer mockController.Finish()

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// execute like the task manager
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go func() {
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// create and enqueue multiple items
				for i := 1; i <= 5; i++ {
					enqueueItem := &v1willow.EnqueueQueueItem{
						Item:            []byte(fmt.Sprintf(`data %d`, i)),
						KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
						Updateable:      false,
						RetryAttempts:   uint64(i),
						RetryPosition:   "front",
						TimeoutDuration: time.Duration(i) * time.Second,
					}
					g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

					err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
					g.Expect(err).ToNot(HaveOccurred())
				}

				// grab the first dequeu channel
				var dequeueItem *v1willow.DequeueQueueItem
				var success func()
				dequeueChan := memeoryQueueChannel.Dequeue()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, _ = dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).ToNot(BeNil())

					g.Expect(dequeueItem.Item).To(Equal([]byte(`data 1`)))
					g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
					g.Expect(dequeueItem.KeyValues).To(Equal(defaultKeyValues(g)))
					g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}

				// ensure that we cannot dequeue another item
				g.Consistently(dequeueChan).ShouldNot(Receive())

				// call success and get the 2nd item
				success()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, _ = dequeueFunc(testhelpers.NewContextWithLogger())
					defer success()
					g.Expect(dequeueItem).ToNot(BeNil())

					g.Expect(dequeueItem.Item).To(Equal([]byte(`data 2`)))
					g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
					g.Expect(dequeueItem.KeyValues).To(Equal(defaultKeyValues(g)))
					g.Expect(dequeueItem.TimeoutDuration).To(Equal(2 * time.Second))
				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}
			})
		})

		t.Run("Context fail dequeues", func(t *testing.T) {
			t.Run("It dequeues the same item multiple times if fail is called", func(t *testing.T) {
				// set limiter client to pass
				mockController, fakeLimiterClient := setupSuccessFakeLimiter(t, 9) // 5 for enqueue, 2 for dequeue(), 2 times for fail()
				defer mockController.Finish()

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// execute like the task manager
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go func() {
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// create and enqueue multiple items
				for i := 1; i <= 5; i++ {
					enqueueItem := &v1willow.EnqueueQueueItem{
						Item:            []byte(fmt.Sprintf(`data %d`, i)),
						KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
						Updateable:      false,
						RetryAttempts:   uint64(i),
						RetryPosition:   "front",
						TimeoutDuration: time.Duration(i) * time.Second,
					}
					g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

					err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
					g.Expect(err).ToNot(HaveOccurred())
				}

				// grab the first dequeu channel
				var dequeueItem *v1willow.DequeueQueueItem
				var fail func()
				dequeueChan := memeoryQueueChannel.Dequeue()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, _, fail = dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).ToNot(BeNil())

					g.Expect(dequeueItem.Item).To(Equal([]byte(`data 1`)))
					g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
					g.Expect(dequeueItem.KeyValues).To(Equal(defaultKeyValues(g)))
					g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}

				// ensure that we cannot dequeue another item
				g.Consistently(dequeueChan).ShouldNot(Receive())

				// call success and get the 2nd item
				fail()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, _, fail = dequeueFunc(testhelpers.NewContextWithLogger())
					defer fail()
					g.Expect(dequeueItem).ToNot(BeNil())

					g.Expect(dequeueItem.Item).To(Equal([]byte(`data 1`)))
					g.Expect(dequeueItem.ItemID).ToNot(Equal(""))
					g.Expect(dequeueItem.KeyValues).To(Equal(defaultKeyValues(g)))
					g.Expect(dequeueItem.TimeoutDuration).To(Equal(time.Second))
				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}
			})
		})
	})

	t.Run("Context when the dequeue callback function fails", func(t *testing.T) {
		t.Run("It stops the limit checks if Execute context is canceled", func(t *testing.T) {
			// set limiter client to pass
			mockController, fakeLimiterClient := fakeLimiterClient(t)
			defer mockController.Finish()

			count := 0
			fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
				if count == 0 {
					// allow for a single enqueue
					count++
					return nil
				}

				return fmt.Errorf("error, limit has been reached")
			}).Times(2)

			ruleCount := 0
			fakeLimiterClient.EXPECT().MatchRules(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Rules, error) {
				if ruleCount == 0 {
					ruleCount++

					return v1limiter.Rules{
						&v1limiter.Rule{
							Name: "block rule",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
							},
							Limit: 0,
						},
					}, nil
				}

				return nil, nil
			}).Times(1)

			fakeLimiterClient.EXPECT().MatchOverrides(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ string, _ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Overrides, error) {
				return nil, nil
			}).Times(1)

			// create queue channel
			memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

			// execute like the task manager
			doneExecuting := make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				defer close(doneExecuting)
				_ = memeoryQueueChannel.Execute(ctx)
			}()

			// create and enqueue an item
			enqueueItem := &v1willow.EnqueueQueueItem{
				Item:            []byte(`data`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
				Updateable:      false,
				RetryAttempts:   1,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

			err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
			g.Expect(err).ToNot(HaveOccurred())

			// grab the first dequeu channel
			dequeueChan := memeoryQueueChannel.Dequeue()
			select {
			case dequeueFunc := <-dequeueChan:
				dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
				g.Expect(dequeueItem).To(BeNil())
				g.Expect(success).To(BeNil())
				g.Expect(fail).To(BeNil())

			case <-time.After(time.Second):
				g.Fail("failed to dequeue item")
			}

			cancel()
			g.Eventually(doneExecuting).Should(BeClosed())
		})

		t.Run("It stops the limit checks if ForceDelete is called", func(t *testing.T) {
			// set limiter client to pass
			mockController, fakeLimiterClient := fakeLimiterClient(t)
			defer mockController.Finish()

			fakeLimiterClient.EXPECT().SetCounters(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).AnyTimes()

			count := 0
			fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
				if count == 0 {
					// allow for a single enqueue
					count++
					return nil
				}

				return fmt.Errorf("error, limit has been reached")
			}).Times(2)

			ruleCount := 0
			fakeLimiterClient.EXPECT().MatchRules(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Rules, error) {
				if ruleCount == 0 {
					ruleCount++

					return v1limiter.Rules{
						&v1limiter.Rule{
							Name: "block rule",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
							},
							Limit: 0,
						},
					}, nil
				}

				return nil, nil
			}).Times(1)

			fakeLimiterClient.EXPECT().MatchOverrides(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ string, _ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Overrides, error) {
				return nil, nil
			}).Times(1)

			// create queue channel
			memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

			// execute like the task manager
			doneExecuting := make(chan struct{})
			go func() {
				defer close(doneExecuting)
				_ = memeoryQueueChannel.Execute(context.Background())
			}()

			// create and enqueue an item
			enqueueItem := &v1willow.EnqueueQueueItem{
				Item:            []byte(`data`),
				KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
				Updateable:      false,
				RetryAttempts:   1,
				RetryPosition:   "front",
				TimeoutDuration: time.Second,
			}
			g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

			err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
			g.Expect(err).ToNot(HaveOccurred())

			// grab the first dequeu channel
			dequeueChan := memeoryQueueChannel.Dequeue()
			select {
			case dequeueFunc := <-dequeueChan:
				dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
				g.Expect(dequeueItem).To(BeNil())
				g.Expect(success).To(BeNil())
				g.Expect(fail).To(BeNil())
			case <-time.After(time.Second):
				g.Fail("failed to dequeue item")
			}

			memeoryQueueChannel.ForceDelete(testhelpers.NewContextWithLogger())
			g.Eventually(doneExecuting).Should(BeClosed())
		})

		t.Run("Context when checking to know when the limit is no longer an issue", func(t *testing.T) {
			t.Run("It stops when there are no rules", func(t *testing.T) {
				// set limiter client to pass
				mockController, fakeLimiterClient := fakeLimiterClient(t)
				defer mockController.Finish()

				count := 0
				fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
					if count == 1 {
						count++
						return fmt.Errorf("error, limit has been reached")
					}

					count++
					return nil
				}).Times(3)

				ruleCount := 0
				fakeLimiterClient.EXPECT().MatchRules(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Rules, error) {
					if ruleCount == 0 {
						ruleCount++

						return v1limiter.Rules{
							&v1limiter.Rule{
								Name: "block rule",
								GroupByKeyValues: datatypes.KeyValues{
									"one": datatypes.Any(),
								},
								Limit: 0,
							},
						}, nil
					}

					return nil, nil
				}).Times(2)

				fakeLimiterClient.EXPECT().MatchOverrides(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ string, _ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Overrides, error) {
					return nil, nil
				}).Times(1)

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// execute like the task manager
				doneExecuting := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					defer close(doneExecuting)
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// create and enqueue an item
				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(`data`),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      false,
					RetryAttempts:   1,
					RetryPosition:   "front",
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

				err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
				g.Expect(err).ToNot(HaveOccurred())

				// grab the first dequeu channel
				// first call should return nil to setup the channel is blocked
				dequeueChan := memeoryQueueChannel.Dequeue()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).To(BeNil())
					g.Expect(success).To(BeNil())
					g.Expect(fail).To(BeNil())

				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}

				// second call should return an item as the limiter client now passes
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).ToNot(BeNil())
					g.Expect(success).ToNot(BeNil())
					g.Expect(fail).ToNot(BeNil())

					success()
				case <-time.After(6 * time.Second):
					g.Fail("failed to dequeue item")
				}

				cancel()
				g.Eventually(doneExecuting).Should(BeClosed())
			})

			t.Run("It stops when all Rules are unlimited", func(t *testing.T) {
				// set limiter client to pass
				mockController, fakeLimiterClient := fakeLimiterClient(t)
				defer mockController.Finish()

				count := 0
				fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
					if count == 0 {
						// allow for a single enqueue
						count++
						return nil
					}

					return fmt.Errorf("error, limit has been reached")
				}).Times(2)

				fakeLimiterClient.EXPECT().MatchRules(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Rules, error) {
					return v1limiter.Rules{
						&v1limiter.Rule{
							Name: "rule1",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
							},
							Limit: -1,
						},
						&v1limiter.Rule{
							Name: "rule2",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
								"two": datatypes.Any(),
							},
							Limit: -1,
						},
					}, nil
				}).Times(1)

				fakeLimiterClient.EXPECT().MatchOverrides(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ string, _ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Overrides, error) {
					return nil, nil
				}).Times(2)

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// execute like the task manager
				doneExecuting := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					defer close(doneExecuting)
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// create and enqueue an item
				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(`data`),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      false,
					RetryAttempts:   1,
					RetryPosition:   "front",
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

				err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
				g.Expect(err).ToNot(HaveOccurred())

				// grab the first dequeu channel
				dequeueChan := memeoryQueueChannel.Dequeue()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).To(BeNil())
					g.Expect(success).To(BeNil())
					g.Expect(fail).To(BeNil())

				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}

				cancel()
				g.Eventually(doneExecuting).Should(BeClosed())
			})

			t.Run("It stops when all Rule's Overrides are unlimited", func(t *testing.T) {
				// set limiter client to pass
				mockController, fakeLimiterClient := fakeLimiterClient(t)
				defer mockController.Finish()

				count := 0
				fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
					if count == 0 {
						// allow for a single enqueue
						count++
						return nil
					}

					return fmt.Errorf("error, limit has been reached")
				}).Times(2)

				fakeLimiterClient.EXPECT().MatchRules(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Rules, error) {
					return v1limiter.Rules{
						&v1limiter.Rule{
							Name: "rule1",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
							},
							Limit: 1,
						},
						&v1limiter.Rule{
							Name: "rule2",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
								"two": datatypes.Any(),
							},
							Limit: 0,
						},
					}, nil
				}).Times(1)

				overrideCounter := 0
				fakeLimiterClient.EXPECT().MatchOverrides(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ string, _ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Overrides, error) {
					if overrideCounter == 0 {
						overrideCounter++
						return v1limiter.Overrides{
							&v1limiter.Override{
								Name:      "override1",
								KeyValues: datatypes.KeyValues{"one": datatypes.Int(1)},
								Limit:     -1,
							},
						}, nil
					}

					return v1limiter.Overrides{
						&v1limiter.Override{
							Name:      "override2",
							KeyValues: datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Int(2)},
							Limit:     -1,
						},
					}, nil
				}).Times(2)

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// execute like the task manager
				doneExecuting := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					defer close(doneExecuting)
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// create and enqueue an item
				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(`data`),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      false,
					RetryAttempts:   1,
					RetryPosition:   "front",
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

				err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
				g.Expect(err).ToNot(HaveOccurred())

				// grab the first dequeu channel

				dequeueChan := memeoryQueueChannel.Dequeue()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).To(BeNil())
					g.Expect(success).To(BeNil())
					g.Expect(fail).To(BeNil())

				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}

				cancel()
				g.Eventually(doneExecuting).Should(BeClosed())
			})

			t.Run("It stops when the rule is less than the counters", func(t *testing.T) {
				mockController, fakeLimiterClient := fakeLimiterClient(t)
				defer mockController.Finish()

				count := 0
				fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
					if count == 0 {
						// allow for a single enqueue
						count++
						return nil
					}

					return fmt.Errorf("error, limit has been reached")
				}).Times(2)

				fakeLimiterClient.EXPECT().MatchRules(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Rules, error) {
					return v1limiter.Rules{
						&v1limiter.Rule{
							Name: "rule1",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
							},
							Limit: 8,
						},
						&v1limiter.Rule{
							Name: "rule2",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
								"two": datatypes.Any(),
							},
							Limit: 12,
						},
					}, nil
				}).Times(1)

				fakeLimiterClient.EXPECT().MatchOverrides(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ string, _ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Overrides, error) {
					return nil, nil
				}).Times(2)

				fakeLimiterClient.EXPECT().QueryCounters(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *queryassociatedaction.AssociatedActionQuery, _ http.Header) (v1limiter.Counters, error) {
					return v1limiter.Counters{
						&v1limiter.Counter{
							Counters:  3,
							KeyValues: datatypes.KeyValues{"one": datatypes.Int(1), "zone": datatypes.String("west")},
						},
						&v1limiter.Counter{
							Counters:  4,
							KeyValues: datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Int(2), "zone": datatypes.String("west")},
						},
					}, nil
				}).Times(2) // called for each rule

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// execute like the task manager
				doneExecuting := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					defer close(doneExecuting)
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// create and enqueue an item
				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(`data`),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      false,
					RetryAttempts:   1,
					RetryPosition:   "front",
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

				err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
				g.Expect(err).ToNot(HaveOccurred())

				// grab the first dequeu channel

				dequeueChan := memeoryQueueChannel.Dequeue()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).To(BeNil())
					g.Expect(success).To(BeNil())
					g.Expect(fail).To(BeNil())

				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}

				cancel()
				g.Eventually(doneExecuting).Should(BeClosed())
			})

			t.Run("It stops when all Rule's Overrides are less than the counters", func(t *testing.T) {
				// set limiter client to pass
				mockController, fakeLimiterClient := fakeLimiterClient(t)
				defer mockController.Finish()

				count := 0
				fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error {
					if count == 0 {
						// allow for a single enqueue
						count++
						return nil
					}

					return fmt.Errorf("error, limit has been reached")
				}).Times(2)

				fakeLimiterClient.EXPECT().MatchRules(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Rules, error) {
					return v1limiter.Rules{
						&v1limiter.Rule{
							Name: "rule1",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
							},
							Limit: 1,
						},
						&v1limiter.Rule{
							Name: "rule2",
							GroupByKeyValues: datatypes.KeyValues{
								"one": datatypes.Any(),
								"two": datatypes.Any(),
							},
							Limit: 0,
						},
					}, nil
				}).Times(1)

				overrideCounter := 0
				fakeLimiterClient.EXPECT().MatchOverrides(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ string, _ *querymatchaction.MatchActionQuery, _ http.Header) (v1limiter.Overrides, error) {
					if overrideCounter == 0 {
						overrideCounter++

						return v1limiter.Overrides{
							&v1limiter.Override{
								Name:      "override1",
								KeyValues: datatypes.KeyValues{"one": datatypes.Int(1)},
								Limit:     8,
							},
						}, nil
					}

					return v1limiter.Overrides{
						&v1limiter.Override{
							Name:      "override2",
							KeyValues: datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Int(2)},
							Limit:     12,
						},
						&v1limiter.Override{
							Name:      "override3",
							KeyValues: datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Int(2), "zone": datatypes.String("west")},
							Limit:     9,
						},
					}, nil
				}).Times(2)

				fakeLimiterClient.EXPECT().QueryCounters(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *queryassociatedaction.AssociatedActionQuery, _ http.Header) (v1limiter.Counters, error) {
					return v1limiter.Counters{
						&v1limiter.Counter{
							Counters:  3,
							KeyValues: datatypes.KeyValues{"one": datatypes.Int(1), "zone": datatypes.String("west")},
						},
						&v1limiter.Counter{
							Counters:  4,
							KeyValues: datatypes.KeyValues{"one": datatypes.Int(1), "two": datatypes.Int(2), "zone": datatypes.String("west")},
						},
					}, nil
				}).Times(3) // called for each override

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// execute like the task manager
				doneExecuting := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					defer close(doneExecuting)
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// create and enqueue an item
				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(`data`),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      false,
					RetryAttempts:   1,
					RetryPosition:   "front",
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

				err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
				g.Expect(err).ToNot(HaveOccurred())

				// grab the first dequeu channel

				dequeueChan := memeoryQueueChannel.Dequeue()
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, fail := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).To(BeNil())
					g.Expect(success).To(BeNil())
					g.Expect(fail).To(BeNil())

				case <-time.After(time.Second):
					g.Fail("failed to dequeue item")
				}

				cancel()
				g.Eventually(doneExecuting).Should(BeClosed())
			})
		})
	})
}

func Test_memoryQueueChannel_ACK(t *testing.T) {
	g := NewGomegaWithT(t)

	enqueue := func(g *GomegaWithT, memeoryQueueChannel *memoryQueueChannel, updateable bool, retryAttempts int, retryPosition string, enqueueCount int) func() {
		return func() {
			for i := 0; i < enqueueCount; i++ {
				// create and enqueue an items
				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(fmt.Sprintf(`data %d`, i)),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      updateable,
					RetryAttempts:   uint64(retryAttempts),
					RetryPosition:   retryPosition,
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())

				err := memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)
				g.Expect(err).ToNot(HaveOccurred())
			}
		}
	}

	enqueueAndDequeue := func(g *GomegaWithT, memeoryQueueChannel *memoryQueueChannel, enqueue func()) *v1willow.DequeueQueueItem {
		enqueue()

		// grab the first item
		var dequeueItem *v1willow.DequeueQueueItem
		var success func()

		dequeueChan := memeoryQueueChannel.Dequeue()
		select {
		case dequeueFunc := <-dequeueChan:
			dequeueItem, success, _ = dequeueFunc(testhelpers.NewContextWithLogger())
			g.Expect(dequeueItem).ToNot(BeNil())

			// call success
			success()
		case <-time.After(time.Second):
			g.Fail("failed to dequeue item")
		}

		return dequeueItem
	}

	t.Run("It returns an error if the tree is empty or item id cannot be found", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		ack := &v1willow.ACK{
			ItemID:    "item not found",
			KeyValues: datatypes.KeyValues{"one": datatypes.Int(1)},
			Passed:    true,
		}
		g.Expect(ack.Validate()).ToNot(HaveOccurred())

		delete, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ack)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to find processing item by id"))
		g.Expect(delete).To(BeTrue())
	})

	t.Run("Context when the ack success is true", func(t *testing.T) {
		t.Run("Context when there is only 1 item in the queue", func(t *testing.T) {
			t.Run("It removes the item without re-queuing and signals to destroy the channel", func(t *testing.T) {
				mockController, fakeLimiterClient := fakeLimiterClient(t)
				defer mockController.Finish()

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// 1 for enqueue, 1 for dequeue(). 2 for ack
				fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(v1counter *v1limiter.Counter, headers http.Header) error {
					return nil
				}).Times(4)

				// execute like the task manager
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go func() {
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// deqeue an item successfully
				dequeueItem := enqueueAndDequeue(g, memeoryQueueChannel, enqueue(g, memeoryQueueChannel, true, 2, "front", 1))

				// ack the dequeued item
				ack := &v1willow.ACK{
					ItemID:    dequeueItem.ItemID,
					KeyValues: dequeueItem.KeyValues,
					Passed:    true,
				}
				g.Expect(ack.Validate()).ToNot(HaveOccurred())

				destroyChannel, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ack)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(destroyChannel).To(BeTrue())

				// ensure nothing else is dequeued
				dequeueChan := memeoryQueueChannel.Dequeue()
				g.Consistently(dequeueChan).ShouldNot(Receive())
			})
		})

		t.Run("Context when there are multiple items in the queue", func(t *testing.T) {
			t.Run("It removes the item without re-queuing and does not signal to destroy the channel", func(t *testing.T) {
				mockController, fakeLimiterClient := fakeLimiterClient(t)
				defer mockController.Finish()

				// create queue channel
				memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

				// 2 for enqueue, 1 for dequeue(). 2 for ack
				fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(v1counter *v1limiter.Counter, headers http.Header) error {
					return nil
				}).Times(5)

				// execute like the task manager
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go func() {
					_ = memeoryQueueChannel.Execute(ctx)
				}()

				// deqeue an item successfully
				dequeueItem := enqueueAndDequeue(g, memeoryQueueChannel, enqueue(g, memeoryQueueChannel, false, 2, "front", 2))

				// ack the dequeued item
				ack := &v1willow.ACK{
					ItemID:    dequeueItem.ItemID,
					KeyValues: dequeueItem.KeyValues,
					Passed:    true,
				}
				g.Expect(ack.Validate()).ToNot(HaveOccurred())

				destroyChannel, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ack)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(destroyChannel).To(BeFalse())
			})
		})
	})

	t.Run("Context when the ack success is false", func(t *testing.T) {
		t.Run("It re-queues the item when under the requeue limit", func(t *testing.T) {
			mockController, fakeLimiterClient := fakeLimiterClient(t)
			defer mockController.Finish()

			// create queue channel
			memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

			// 1 for enqueue, 2 for dequeue(), 1 for failHeartbeat(), 2 for ack
			fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(v1counter *v1limiter.Counter, headers http.Header) error {
				return nil
			}).Times(6)

			// execute like the task manager
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				_ = memeoryQueueChannel.Execute(ctx)
			}()

			// deqeue an item successfully
			dequeueItem := enqueueAndDequeue(g, memeoryQueueChannel, enqueue(g, memeoryQueueChannel, true, 1, "front", 1))

			// ack the dequeued item
			ackFalse := &v1willow.ACK{
				ItemID:    dequeueItem.ItemID,
				KeyValues: dequeueItem.KeyValues,
				Passed:    false,
			}
			g.Expect(ackFalse.Validate()).ToNot(HaveOccurred())

			destroyChannel, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ackFalse)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(destroyChannel).To(BeFalse())

			// ensure we can dequeue the same item again
			dequeueChan := memeoryQueueChannel.Dequeue()
			select {
			case dequeueFunc := <-dequeueChan:
				var success func()
				dequeueItem, success, _ = dequeueFunc(testhelpers.NewContextWithLogger())
				g.Expect(dequeueItem).ToNot(BeNil())

				// call success for dequeuing the item
				success()
			case <-time.After(time.Second):
				g.Fail("failed to dequeue item")
			}

			// 2nd call to ack can remove the item since there are no more retry attempts
			ackTrue := &v1willow.ACK{
				ItemID:    dequeueItem.ItemID,
				KeyValues: dequeueItem.KeyValues,
				Passed:    false,
			}
			g.Expect(ackTrue.Validate()).ToNot(HaveOccurred())

			destroyChannel, err = memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ackTrue)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(destroyChannel).To(BeTrue())

			// ensure we can not dequeue the same item again
			dequeueChan = memeoryQueueChannel.Dequeue()
			g.Consistently(dequeueChan).ShouldNot(Receive())
		})
	})
}

func Test_memoryQueueChannel_async(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can enqueue many items at once", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).AnyTimes()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = memeoryQueueChannel.Execute(ctx)
		}()

		// create and enqueue many items
		wg := new(sync.WaitGroup)
		for i := 0; i < 1_024; i++ {
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				location := "front"
				if count%2 == 0 {
					location = "back"
				}

				updateable := false
				if count%3 == 0 {
					updateable = true
				}

				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(fmt.Sprintf(`data %d`, count)),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      updateable,
					RetryAttempts:   2,
					RetryPosition:   location,
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())
				g.Expect(memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)).ToNot(HaveOccurred())
			}(i)
		}

		wg.Wait()
	})

	t.Run("It can dequeue all items when success is called", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).AnyTimes()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = memeoryQueueChannel.Execute(ctx)
		}()

		// create and enqueue many items
		wg := new(sync.WaitGroup)
		for i := 0; i < 1_024; i++ {
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				location := "front"
				if count%2 == 0 {
					location = "back"
				}

				updateable := false
				if count%3 == 0 {
					updateable = true
				}

				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(fmt.Sprintf(`data %d`, count)),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      updateable,
					RetryAttempts:   2,
					RetryPosition:   location,
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())
				g.Expect(memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		// the updateable items are collapsed
		dequeueChan := memeoryQueueChannel.Dequeue()
		for i := 0; i < 682; i++ {
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				// ensure we can dequeue the same item again
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, _ := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).ToNot(BeNil())

					// call success for dequeuing the item and trigger another item to process
					success()

					wg.Add(1)
					go func() {
						defer wg.Done()

						// ack the item that was dequeued
						ackTrue := &v1willow.ACK{
							ItemID:    dequeueItem.ItemID,
							KeyValues: dequeueItem.KeyValues,
							Passed:    true,
						}
						g.Expect(ackTrue.Validate()).ToNot(HaveOccurred())

						_, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ackTrue)
						g.Expect(err).ToNot(HaveOccurred())
					}()
				case <-time.After(5 * time.Second):
					g.Fail(fmt.Sprintf("failed to dequeue item at index: %d", count))
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("It can dequeue all items when failed is sometimes false", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).AnyTimes()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = memeoryQueueChannel.Execute(ctx)
		}()

		// create and enqueue many items
		wg := new(sync.WaitGroup)
		for i := 0; i < 1_024; i++ {
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				location := "front"
				if count%2 == 0 {
					location = "back"
				}

				updateable := false
				if count%3 == 0 {
					updateable = true
				}

				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(fmt.Sprintf(`data %d`, count)),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      updateable,
					RetryAttempts:   2,
					RetryPosition:   location,
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())
				g.Expect(memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		// dequeue through all the retries as well
		dequeueChan := memeoryQueueChannel.Dequeue()
		for i := 0; i < 910; i++ { // (1024 - (1024/3)) + ((1024 - (1024/3))/4) + (((1024 - (1024/3))/4)/3)
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				// ensure we can dequeue the same item again
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, failed := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).ToNot(BeNil())

					// call success for dequeuing the item and trigger another item to process
					if count%4 == 0 {
						failed()
					} else {
						success()

						wg.Add(1)
						go func() {
							defer wg.Done()

							// ack the item that was dequeued
							ackTrue := &v1willow.ACK{
								ItemID:    dequeueItem.ItemID,
								KeyValues: dequeueItem.KeyValues,
								Passed:    true,
							}
							g.Expect(ackTrue.Validate()).ToNot(HaveOccurred())

							_, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ackTrue)
							g.Expect(err).ToNot(HaveOccurred())
						}()
					}
				case <-time.After(5 * time.Second):
					// just timeout the fact that some can requeue at the front /w out knowing if there are items in the queue
					// makes it unkown how long we will actaully loop. at most will be 910 times
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("It can dequeue all items when failed is sometimes called and timeouts occur", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).AnyTimes()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = memeoryQueueChannel.Execute(ctx)
		}()

		// create and enqueue many (682) since collapse will trigger
		wg := new(sync.WaitGroup)
		for i := 0; i < 1_024; i++ {
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				updateable := false
				if count%3 == 0 {
					updateable = true
				}

				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(fmt.Sprintf(`data %d`, count)),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      updateable,
					RetryAttempts:   2,
					RetryPosition:   "front",
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())
				g.Expect(memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)).ToNot(HaveOccurred())
			}(i)
		}
		wg.Wait()

		// dequeue through all the retries as well
		dequeueChan := memeoryQueueChannel.Dequeue()
		for i := 0; i < 921; i++ { // (1024 - (1024/3)) + ((1024 - (1024/3))/4) + (((1024 - (1024/3))/4)/3) + ((((1024 - (1024/3))/5)/4)/3)
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				// ensure we can dequeue the same item again
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, failed := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).ToNot(BeNil())

					// call success for dequeuing the item and trigger another item to process
					if count%4 == 0 {
						failed()
					} else if count%5 == 0 {
						success()
						// but then timeout because no ack
					} else {
						success()

						wg.Add(1)
						go func() {
							defer wg.Done()

							// ack the item that was dequeued
							ackTrue := &v1willow.ACK{
								ItemID:    dequeueItem.ItemID,
								KeyValues: dequeueItem.KeyValues,
								Passed:    true,
							}
							g.Expect(ackTrue.Validate()).ToNot(HaveOccurred())

							_, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ackTrue)
							g.Expect(err).ToNot(HaveOccurred())
						}()
					}
				case <-time.After(5 * time.Second):
					// just timeout the fact that some can requeue at the front /w out knowing if there are items in the queue
					// makes it unkown how long we will actaully loop. at most will be 921 times
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("It can run everything all at once", func(t *testing.T) {
		mockController, fakeLimiterClient := fakeLimiterClient(t)
		defer mockController.Finish()

		// set limiter client to pass
		fakeLimiterClient.EXPECT().UpdateCounter(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *v1limiter.Counter, _ http.Header) error { return nil }).AnyTimes()

		// create queue channel
		memeoryQueueChannel := New(fakeLimiterClient, func() {}, "test", defaultKeyValues(g))

		// execute like the task manager
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			_ = memeoryQueueChannel.Execute(ctx)
		}()

		dequeueChan := memeoryQueueChannel.Dequeue()

		wg := new(sync.WaitGroup)
		for i := 0; i < 1_024; i++ {
			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				location := "front"
				if count%2 == 0 {
					location = "back"
				}

				updateable := false
				if count%3 == 0 {
					updateable = true
				}

				enqueueItem := &v1willow.EnqueueQueueItem{
					Item:            []byte(fmt.Sprintf(`data %d`, count)),
					KeyValues:       datatypes.KeyValues{"one": datatypes.Int(1)},
					Updateable:      updateable,
					RetryAttempts:   2,
					RetryPosition:   location,
					TimeoutDuration: time.Second,
				}
				g.Expect(enqueueItem.Validate()).ToNot(HaveOccurred())
				g.Expect(memeoryQueueChannel.Enqueue(testhelpers.NewContextWithLogger(), enqueueItem)).ToNot(HaveOccurred())
			}(i)

			wg.Add(1)
			go func(count int) {
				defer wg.Done()

				// ensure we can dequeue the same item again
				select {
				case dequeueFunc := <-dequeueChan:
					dequeueItem, success, failed := dequeueFunc(testhelpers.NewContextWithLogger())
					g.Expect(dequeueItem).ToNot(BeNil())

					// call success for dequeuing the item and trigger another item to process
					if count%4 == 0 {
						failed()
					} else if count%5 == 0 {
						success()
						// but then timeout because no ack
					} else {
						success()

						wg.Add(1)
						go func() {
							defer wg.Done()

							// ack the item that was dequeued
							ackTrue := &v1willow.ACK{
								ItemID:    dequeueItem.ItemID,
								KeyValues: dequeueItem.KeyValues,
								Passed:    true,
							}
							g.Expect(ackTrue.Validate()).ToNot(HaveOccurred())

							_, err := memeoryQueueChannel.ACK(testhelpers.NewContextWithLogger(), ackTrue)
							g.Expect(err).ToNot(HaveOccurred())
						}()
					}
				case <-time.After(time.Second):
					// This is fine to timeout since we don't know 100% how many items we are pulling off the queue.
					// in some cases updated will be called, and some cases those will be collapsed
				}
			}(i)
		}
		wg.Wait()
	})
}
