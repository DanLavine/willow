package memory

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/tasks"
	"github.com/DanLavine/willow/internal/v1/tags"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	. "github.com/onsi/gomega"
)

var (
	chan1 = make(chan tags.Tag)
	chan2 = make(chan tags.Tag)
	chan3 = make(chan tags.Tag)

	tagGroupChans = []chan tags.Tag{chan1, chan2, chan3}
)

func TestMemoryTagGroup_Enqueue(t *testing.T) {
	g := NewGomegaWithT(t)

	queueTags := []string{"one", "two", "three"}
	taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())

	// start task manager
	started := make(chan struct{})
	taskManager.AddTask("start", tasks.Running(started))

	ctx, stop := context.WithCancel(context.Background())
	go func() {
		_ = taskManager.Run(ctx)
	}()
	defer stop()

	g.Eventually(started).Should(BeClosed())

	t.Run("allows a message to be processed from all readers", func(t *testing.T) {
		counter := new(atomic.Uint64)
		chans := []chan<- tags.Tag{chan1, chan2, chan3}
		tagsGroup := newTagGroup(chans)
		defer tagsGroup.Stop()

		// allow for message processing
		g.Expect(taskManager.AddRunningTask("", tagsGroup)).ToNot(HaveOccurred())

		var dequeueFunc tags.Tag
		for index, channel := range tagGroupChans {
			enqueueItem := &v1.EnqueueItem{Name: "name", Tags: queueTags, Data: []byte(fmt.Sprintf("%d", index)), Updateable: false}

			g.Expect(tagsGroup.Enqueue(enqueueItem, counter)).ToNot(HaveOccurred())

			cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
			g.Eventually(func() tags.Tag {
				select {
				case <-cdl.Done():
				case dequeueFunc = <-channel:
				}

				return dequeueFunc
			}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
			cancel()

			dequeueMessage := dequeueFunc()
			g.Expect(dequeueMessage.Name).To(Equal("name"))
			g.Expect(dequeueMessage.ID).To(Equal(uint64(index + 1)))
			g.Expect(dequeueMessage.Data).To(Equal([]byte(fmt.Sprintf("%d", index))))
			g.Expect(dequeueMessage.Tags).To(Equal(queueTags))
		}

		g.Expect(counter.Load()).To(Equal(uint64(3)))
	})

	t.Run("when a message is updateable, they collapse when not being processed", func(t *testing.T) {
		counter := new(atomic.Uint64)
		chans := []chan<- tags.Tag{chan1}
		tagsGroup := newTagGroup(chans)
		defer tagsGroup.Stop()

		// allow for message processing
		g.Expect(taskManager.AddRunningTask("", tagsGroup)).ToNot(HaveOccurred())

		// only last enqueu should be valid
		g.Expect(tagsGroup.Enqueue(&v1.EnqueueItem{Name: "name", Tags: queueTags, Data: []byte(`1`), Updateable: true}, counter)).ToNot(HaveOccurred())
		g.Expect(tagsGroup.Enqueue(&v1.EnqueueItem{Name: "name", Tags: queueTags, Data: []byte(`2`), Updateable: true}, counter)).ToNot(HaveOccurred())
		g.Expect(tagsGroup.Enqueue(&v1.EnqueueItem{Name: "name", Tags: queueTags, Data: []byte(`3`), Updateable: true}, counter)).ToNot(HaveOccurred())
		g.Expect(tagsGroup.Enqueue(&v1.EnqueueItem{Name: "name", Tags: queueTags, Data: []byte(`4`), Updateable: true}, counter)).ToNot(HaveOccurred())

		var dequeueFunc tags.Tag
		cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		g.Eventually(func() tags.Tag {
			select {
			case <-cdl.Done():
			case dequeueFunc = <-chan1:
			}

			return dequeueFunc
		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
		cancel()

		dequeueMessage := dequeueFunc()
		g.Expect(dequeueMessage.Name).To(Equal("name"))
		g.Expect(dequeueMessage.ID).To(Equal(uint64(1)))
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`4`)))
		g.Expect(dequeueMessage.Tags).To(Equal(queueTags))

		g.Expect(counter.Load()).To(Equal(uint64(1)))
	})

	t.Run("when a message is updateable and is processing, the next message is added as a new message", func(t *testing.T) {
		counter := new(atomic.Uint64)
		chans := []chan<- tags.Tag{chan1}
		tagsGroup := newTagGroup(chans)
		defer tagsGroup.Stop()

		// allow for message processing
		g.Expect(taskManager.AddRunningTask("", tagsGroup)).ToNot(HaveOccurred())

		// only last enqueu should be valid
		g.Expect(tagsGroup.Enqueue(&v1.EnqueueItem{Name: "name", Tags: queueTags, Data: []byte(`1`), Updateable: true}, counter)).ToNot(HaveOccurred())

		var dequeueFunc tags.Tag
		cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		g.Eventually(func() tags.Tag {
			select {
			case <-cdl.Done():
			case dequeueFunc = <-chan1:
			}

			return dequeueFunc
		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
		cancel()

		dequeueMessage := dequeueFunc()
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`1`)))

		// enqueu another message should process
		g.Expect(tagsGroup.Enqueue(&v1.EnqueueItem{Name: "name", Tags: queueTags, Data: []byte(`2`), Updateable: true}, counter)).ToNot(HaveOccurred())

		cdl, cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Second))
		g.Eventually(func() tags.Tag {
			select {
			case <-cdl.Done():
			case dequeueFunc = <-chan1:
			}

			return dequeueFunc
		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
		cancel()

		dequeueMessage = dequeueFunc()
		g.Expect(dequeueMessage.Data).To(Equal([]byte(`2`)))
		g.Expect(counter.Load()).To(Equal(uint64(2)))
	})
}
