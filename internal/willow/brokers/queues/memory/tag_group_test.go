package memory

//import (
//	"context"
//	"fmt"
//	"testing"
//	"time"
//
//	"github.com/DanLavine/goasync"
//	"github.com/DanLavine/willow/internal/brokers/tags"
//	"github.com/DanLavine/willow/pkg/models/datatypes"
//	v1 "github.com/DanLavine/willow/pkg/models/v1"
//	. "github.com/onsi/gomega"
//)
//
//var (
//	chan1 = make(chan tags.Tag)
//	chan2 = make(chan tags.Tag)
//	chan3 = make(chan tags.Tag)
//
//	tagGroupChans = []chan tags.Tag{chan1, chan2, chan3}
//
//	queueTags = datatypes.Strings{"one", "two", "three"}
//)
//
//func TestMemoryTagGroup_Enqueue(t *testing.T) {
//	g := NewGomegaWithT(t)
//	taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	go func() {
//		_ = taskManager.Run(ctx)
//	}()
//
//	t.Run("allows a message to be processed from all readers", func(t *testing.T) {
//		// setup tags group
//		counter := NewCounter(5)
//		tagsGroup := newTagGroup(queueTags, []chan<- tags.Tag{chan1, chan2, chan3})
//		defer tagsGroup.Stop()
//
//		// allow for message processing
//		g.Expect(taskManager.AddExecuteTask("", tagsGroup)).ToNot(HaveOccurred())
//
//		var dequeueFunc tags.Tag
//		for index, channel := range tagGroupChans {
//			enqueueItem := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "test", BrokerType: v1.Queue, Tags: datatypes.Strings{""}}, Data: []byte(fmt.Sprintf("%d", index)), Updateable: false}
//			g.Expect(tagsGroup.Enqueue(counter, enqueueItem)).ToNot(HaveOccurred())
//
//			cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
//			g.Eventually(func() tags.Tag {
//				select {
//				case <-cdl.Done():
//				case dequeueFunc = <-channel:
//				}
//
//				return dequeueFunc
//			}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
//			cancel()
//
//			dequeueMessage := dequeueFunc()
//			g.Expect(dequeueMessage.BrokerInfo.Name).To(Equal(datatypes.String("test")))
//			g.Expect(dequeueMessage.BrokerInfo.Tags).To(Equal(queueTags))
//			g.Expect(dequeueMessage.ID).To(Equal(uint64(index + 1)))
//			g.Expect(dequeueMessage.Data).To(Equal([]byte(fmt.Sprintf("%d", index))))
//		}
//	})
//
//	t.Run("when a message is updateable, they collapse when not being processed", func(t *testing.T) {
//		counter := NewCounter(5)
//		chans := []chan<- tags.Tag{chan1}
//		tagsGroup := newTagGroup(queueTags, chans)
//		defer tagsGroup.Stop()
//
//		// allow for message processing
//		g.Expect(taskManager.AddExecuteTask("", tagsGroup)).ToNot(HaveOccurred())
//
//		// enqueue 4 items that shoudl collapse to the last
//		enqueueItem1 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "name", BrokerType: v1.Queue, Tags: datatypes.Strings{""}}, Data: []byte(`1`), Updateable: true}
//		enqueueItem2 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "name", BrokerType: v1.Queue, Tags: datatypes.Strings{""}}, Data: []byte(`2`), Updateable: true}
//		enqueueItem3 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "name", BrokerType: v1.Queue, Tags: datatypes.Strings{""}}, Data: []byte(`3`), Updateable: true}
//		enqueueItem4 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "name", BrokerType: v1.Queue, Tags: datatypes.Strings{""}}, Data: []byte(`4`), Updateable: true}
//		g.Expect(tagsGroup.Enqueue(counter, enqueueItem1)).ToNot(HaveOccurred())
//		g.Expect(tagsGroup.Enqueue(counter, enqueueItem2)).ToNot(HaveOccurred())
//		g.Expect(tagsGroup.Enqueue(counter, enqueueItem3)).ToNot(HaveOccurred())
//		g.Expect(tagsGroup.Enqueue(counter, enqueueItem4)).ToNot(HaveOccurred())
//
//		// only last enqueu should be valid
//		var dequeueFunc tags.Tag
//		cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
//		g.Eventually(func() tags.Tag {
//			select {
//			case <-cdl.Done():
//			case dequeueFunc = <-chan1:
//			}
//
//			return dequeueFunc
//		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
//		cancel()
//
//		dequeueMessage := dequeueFunc()
//		g.Expect(dequeueMessage.BrokerInfo.Name).To(Equal(datatypes.String("name")))
//		g.Expect(dequeueMessage.BrokerInfo.Tags).To(Equal(queueTags))
//		g.Expect(dequeueMessage.ID).To(Equal(uint64(1)))
//		g.Expect(dequeueMessage.Data).To(Equal([]byte(`4`)))
//	})
//
//	t.Run("when a message is updateable and is processing, the next message is added as a new message", func(t *testing.T) {
//		counter := NewCounter(5)
//		chans := []chan<- tags.Tag{chan1}
//		tagsGroup := newTagGroup(queueTags, chans)
//		defer tagsGroup.Stop()
//
//		// allow for message processing
//		g.Expect(taskManager.AddExecuteTask("", tagsGroup)).ToNot(HaveOccurred())
//
//		// enqueue first item
//		enqueueItem1 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "name", BrokerType: v1.Queue, Tags: datatypes.Strings{""}}, Data: []byte(`1`), Updateable: true}
//		g.Expect(tagsGroup.Enqueue(counter, enqueueItem1)).ToNot(HaveOccurred())
//
//		var dequeueFunc tags.Tag
//		cdl, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
//		g.Eventually(func() tags.Tag {
//			select {
//			case <-cdl.Done():
//			case dequeueFunc = <-chan1:
//			}
//
//			return dequeueFunc
//		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
//		cancel()
//
//		dequeueMessage := dequeueFunc()
//		g.Expect(dequeueMessage.Data).To(Equal([]byte(`1`)))
//
//		// enqueu another message should process
//		enqueueItem2 := &v1.EnqueueItemRequest{BrokerInfo: v1.BrokerInfo{Name: "name", BrokerType: v1.Queue, Tags: datatypes.Strings{""}}, Data: []byte(`2`), Updateable: true}
//		g.Expect(tagsGroup.Enqueue(counter, enqueueItem2)).ToNot(HaveOccurred())
//
//		cdl, cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Second))
//		g.Eventually(func() tags.Tag {
//			select {
//			case <-cdl.Done():
//			case dequeueFunc = <-chan1:
//			}
//
//			return dequeueFunc
//		}, time.Second, 100*time.Millisecond, cdl).ShouldNot(BeNil())
//		cancel()
//
//		dequeueMessage = dequeueFunc()
//		g.Expect(dequeueMessage.Data).To(Equal([]byte(`2`)))
//	})
//}
//
//func TestMemoryTagGroup_Metrics(t *testing.T) {
//	g := NewGomegaWithT(t)
//	taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
//	context, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	go func() {
//		_ = taskManager.Run(context)
//	}()
//
//	t.Run("returns the tags group properties", func(t *testing.T) {
//		chans := []chan<- tags.Tag{chan1, chan2, chan3}
//		tagsGroup := newTagGroup(queueTags, chans)
//
//		metrics := tagsGroup.Metrics()
//		g.Expect(metrics.Tags).To(Equal(queueTags))
//	})
//}
