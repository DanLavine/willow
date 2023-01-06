package queues_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DanLavine/willow/internal/v1/queues"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

func TestDiskQueueManager_Create(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("creates a queue without a tag", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		g.Expect(dqm.Create(nil)).ToNot(HaveOccurred())

		metrics := dqm.Metrics()
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags).To(BeNil())
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("creates a queue with a tag", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())

		metrics := dqm.Metrics()
		g.Expect(len(metrics.Queues)).To(Equal(1))
		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"tag1"}))
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("can create a queue with multiple tags", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())
		g.Expect(dqm.Create([]string{"tag2", "tag1"})).ToNot(HaveOccurred())

		metrics := dqm.Metrics()
		g.Expect(len(metrics.Queues)).To(Equal(2))
		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"tag1"}))
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))

		g.Expect(metrics.Queues[1].Tags).To(Equal([]string{"tag1", "tag2"}))
		g.Expect(metrics.Queues[1].Ready).To(Equal(uint64(0)))
		g.Expect(metrics.Queues[1].Processing).To(Equal(uint64(0)))
	})

	t.Run("tags in any order are treated as the same queue", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		g.Expect(dqm.Create([]string{"tag1", "tag2"})).ToNot(HaveOccurred())
		g.Expect(dqm.Create([]string{"tag2", "tag1"})).ToNot(HaveOccurred())

		metrics := dqm.Metrics()
		g.Expect(len(metrics.Queues)).To(Equal(1))
	})
}

func TestDiskQueueManager_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("creates a ready item when the queue exists", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())
		g.Expect(dqm.Enqueue([]byte(`data`), false, []string{"tag1"})).ToNot(HaveOccurred())

		metrics := dqm.Metrics()
		g.Expect(metrics.Queues).ToNot(BeNil())
		g.Expect(metrics.Queues[0].Tags).To(Equal([]string{"tag1"}))
		g.Expect(metrics.Queues[0].Ready).To(Equal(uint64(1)))
		g.Expect(metrics.Queues[0].Processing).To(Equal(uint64(0)))
	})

	t.Run("returns an error if the queue does not exist", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		err := dqm.Enqueue([]byte(`data`), false, []string{"tag1"})
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("Failed to find queue"))
	})
}

func TestDiskQueueManager_Message(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("when the subscribe method is 'STRICT'", func(t *testing.T) {
		t.Run("returns an error if the queue does not exist", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dqm := queues.NewDiskQueueManager(testDir)
			dequeueMessage, err := dqm.Message(ctx, v1.STRICT, []string{"tag1"})
			g.Expect(err).To(HaveOccurred())
			g.Expect(dequeueMessage).To(BeNil())
		})

		t.Run("it waits untill a message is ready", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dqm := queues.NewDiskQueueManager(testDir)
			g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())

			g.Eventually(func() bool {
				reader := make(chan struct{})

				go func() {
					_, _ = dqm.Message(ctx, v1.STRICT, []string{"tag1"})
					close(reader)
				}()

				select {
				case <-time.After(100 * time.Millisecond):
					return false
				case <-reader:
					return true
				}
			}).Should(BeFalse())
		})

		t.Run("it breaks if the context is canceled", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			dqm := queues.NewDiskQueueManager(testDir)
			g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())

			var err error
			var dequeueMessage *v1.DequeueMessage
			g.Eventually(func() bool {
				dequeueMessage, err = dqm.Message(ctx, v1.STRICT, []string{"tag1"})
				return true
			}).Should(BeTrue())

			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(Equal("Context was canceled"))
			g.Expect(dequeueMessage).To(BeNil())
		})

		t.Run("it can retrieve a message for the exact queue", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			dqm := queues.NewDiskQueueManager(testDir)
			g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())
			g.Expect(dqm.Create([]string{"tag1", "tag2"})).ToNot(HaveOccurred())

			// read from tag 1
			g.Expect(dqm.Enqueue([]byte(`tag1 data`), false, []string{"tag1"})).ToNot(HaveOccurred())
			g.Expect(dqm.Enqueue([]byte(`multitag data`), false, []string{"tag1", "tag2"})).ToNot(HaveOccurred())

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
			var err error
			var dequeueMessage *v1.DequeueMessage
			g.Eventually(func() bool {
				reader := make(chan struct{})

				go func() {
					dequeueMessage, err = dqm.Message(ctx, v1.STRICT, []string{"tag1"})
					close(reader)
				}()

				select {
				case <-ctx.Done():
					return false
				case <-reader:
					return true
				}
			}, time.Second, 100*time.Millisecond, ctx).Should(BeTrue())
			cancel()

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dequeueMessage.ID).To(Equal(uint64(1)))
			g.Expect(dequeueMessage.BrokerTags).To(Equal([]string{"tag1"}))
			g.Expect(dequeueMessage.Data).To(Equal([]byte(`tag1 data`)))
		})
	})

	t.Run("when the subscribe method is 'SUBSET'", func(t *testing.T) {
		t.Run("it waits untill a message is ready", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dqm := queues.NewDiskQueueManager(testDir)
			g.Eventually(func() bool {
				reader := make(chan struct{})

				go func() {
					_, _ = dqm.Message(ctx, v1.SUBSET, []string{"a", "b", "c", "d"})
					close(reader)
				}()

				select {
				case <-time.After(100 * time.Millisecond):
					return false
				case <-reader:
					return true
				}
			}).Should(BeFalse())
		})

		t.Run("it breaks if the context is canceled", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			dqm := queues.NewDiskQueueManager(testDir)
			g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())

			var err error
			var dequeueMessage *v1.DequeueMessage
			g.Eventually(func() bool {
				dequeueMessage, err = dqm.Message(ctx, v1.STRICT, []string{"tag1"})
				return true
			}).Should(BeTrue())

			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(Equal("Context was canceled"))
			g.Expect(dequeueMessage).To(BeNil())
		})

		t.Run("it can recieve a subeset for a new queue", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			dqm := queues.NewDiskQueueManager(testDir)
			g.Expect(dqm.Create([]string{"a", "b", "c", "d"})).ToNot(HaveOccurred())
			g.Expect(dqm.Enqueue([]byte(`hello`), false, []string{"a", "b", "c", "d"})).ToNot(HaveOccurred())

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
			var err error
			var dequeueMessage *v1.DequeueMessage
			g.Eventually(func() bool {
				reader := make(chan struct{})

				go func() {
					dequeueMessage, err = dqm.Message(ctx, v1.SUBSET, []string{"a", "b"})
					close(reader)
				}()

				select {
				case <-ctx.Done():
					return false
				case <-reader:
					return true
				}
			}, time.Second, 100*time.Millisecond, ctx).Should(BeTrue())
			cancel()

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dequeueMessage.ID).To(Equal(uint64(1)))
			g.Expect(dequeueMessage.BrokerTags).To(Equal([]string{"a", "b", "c", "d"}))
			g.Expect(dequeueMessage.Data).To(Equal([]byte(`hello`)))
		})
	})

	t.Run("when the subscribe method is 'ANY'", func(t *testing.T) {
		t.Run("it waits untill a message is ready", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dqm := queues.NewDiskQueueManager(testDir)
			g.Eventually(func() bool {
				reader := make(chan struct{})

				go func() {
					_, _ = dqm.Message(ctx, v1.ANY, []string{"a", "b", "c", "d"})
					close(reader)
				}()

				select {
				case <-time.After(100 * time.Millisecond):
					return false
				case <-reader:
					return true
				}
			}).Should(BeFalse())
		})

		t.Run("it breaks if the context is canceled", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			dqm := queues.NewDiskQueueManager(testDir)
			g.Expect(dqm.Create([]string{"blam"})).ToNot(HaveOccurred())

			var err error
			var dequeueMessage *v1.DequeueMessage
			g.Eventually(func() bool {
				dequeueMessage, err = dqm.Message(ctx, v1.ANY, []string{"blam"})
				return true
			}).Should(BeTrue())

			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(Equal("Context was canceled"))
			g.Expect(dequeueMessage).To(BeNil())
		})

		t.Run("it can recieve a subeset for a new queue", func(t *testing.T) {
			testDir := os.TempDir()
			defer os.RemoveAll(testDir)

			dqm := queues.NewDiskQueueManager(testDir)
			g.Expect(dqm.Create([]string{"a", "b", "c"})).ToNot(HaveOccurred())
			g.Expect(dqm.Create([]string{"a", "b"})).ToNot(HaveOccurred())
			g.Expect(dqm.Enqueue([]byte(`hello`), false, []string{"a", "b", "c"})).ToNot(HaveOccurred())
			g.Expect(dqm.Enqueue([]byte(`nope`), false, []string{"a", "b"})).ToNot(HaveOccurred())

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
			var err error
			var dequeueMessage *v1.DequeueMessage
			g.Eventually(func() bool {
				reader := make(chan struct{})

				go func() {
					dequeueMessage, err = dqm.Message(ctx, v1.ANY, []string{"c"})
					close(reader)
				}()

				select {
				case <-ctx.Done():
					return false
				case <-reader:
					return true
				}
			}, time.Second, 100*time.Millisecond, ctx).Should(BeTrue())
			cancel()

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dequeueMessage.ID).To(Equal(uint64(1)))
			g.Expect(dequeueMessage.BrokerTags).To(Equal([]string{"a", "b", "c"}))
			g.Expect(dequeueMessage.Data).To(Equal([]byte(`hello`)))
		})
	})
}

func TestACK(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("it returns an error if the queue does not exist", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		g.Expect(dqm.ACK(0, true, []string{"tag1"})).To(HaveOccurred())
	})

	t.Run("it returns an erro if the message is not yet processing", func(t *testing.T) {
		testDir := os.TempDir()
		defer os.RemoveAll(testDir)

		dqm := queues.NewDiskQueueManager(testDir)
		g.Expect(dqm.Create([]string{"tag1"})).ToNot(HaveOccurred())
		g.Expect(dqm.Enqueue([]byte(`data-tag1`), false, []string{"tag1"})).ToNot(HaveOccurred())

		// ack a enqueue value
		err := dqm.ACK(0, true, []string{"tag1"})
		g.Expect(err).To(HaveOccurred())
	})

	//t.Run("when the message passed", func(t *testing.T) {
	//	passed := true

	//	t.Run("the message is removed from the queue", func(t *testing.T) {
	//		testDir := os.TempDir()
	//		defer os.RemoveAll(testDir)

	//		dqm := queues.NewDiskQueueManager(testDir)
	//		defer dqm.Close()

	//		g.Expect(dqm.Create("name1", "tag1")).ToNot(HaveOccurred())
	//		g.Expect(dqm.Enqueue([]byte(`data-tag1`), false, "name1", "tag1")).ToNot(HaveOccurred())

	//		context, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	//		defer cancel()

	//		var data deadlettermodels.Message
	//		g.Eventually(func() deadlettermodels.Message {
	//			queueItemChan := make(chan deadlettermodels.Message, 1)

	//			go func() {
	//				queueItem := dqm.Message("name1", nil)
	//				queueItemChan <- queueItem
	//			}()

	//			select {
	//			case <-context.Done():
	//				return deadlettermodels.Message{}
	//			case data = <-queueItemChan:
	//				return data
	//			}
	//		}, time.Second, 100*time.Millisecond, context).ShouldNot(BeNil())

	//		g.Expect(data.Data).To(Or(Equal([]byte(`data-tag2`)), Equal([]byte(`data-tag1`))))
	//	})
	//})

	//t.Run("when the message failed", func(t *testing.T) {
	//	passed := false

	//	t.Run("the message is requeued", func(t *testing.T) {

	//	})
	//})
}
