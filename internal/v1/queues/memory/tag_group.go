package memory

import (
	"context"
	"reflect"
	"sync"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/v1/tags"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type TagGroup interface {
	Execute(ctx context.Context) error

	Process() *v1.DequeueItemResponse

	Enqueue(queueItem *v1.EnqueueItem) *v1.Error

	StrictTag() <-chan tags.Tag

	Stop()
}

type tagGroup struct {
	lock     *sync.Mutex
	done     chan struct{}
	doneOnce *sync.Once

	items          *datastructures.IDTree
	availableItems []uint64

	notifier      *gonotify.Notify
	channels      []tags.Tag
	strictChannel chan tags.Tag
}

func newTagGroup(channels []tags.Tag) *tagGroup {
	return &tagGroup{
		lock:           new(sync.Mutex),
		done:           make(chan struct{}),
		doneOnce:       new(sync.Once),
		items:          datastructures.NewIDTree(),
		availableItems: []uint64{},
		notifier:       gonotify.New(),
		channels:       channels,
		strictChannel:  make(chan tags.Tag),
	}
}

// Handled by GoAsync to constantly read items from the queue and handle shutdown
func (tg *tagGroup) Execute(ctx context.Context) error {
	defer func() {
		tg.Stop()
		close(tg.strictChannel)
		tg.notifier.ForceStop()
	}()

	cases := []reflect.SelectCase{
		reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(tg.strictChannel), Send: reflect.ValueOf(tg.Process)},
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tg.done)}}

	// setup all possible channels
	for _, channel := range tg.channels {
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(channel), Send: reflect.ValueOf(tg.Process)})
	}

	for {
		select {
		case <-tg.notifier.Ready():
			// ready to process a message on the queue
		case <-tg.done:
			// no more message and all clients are closed. so cleanup
			return nil
		case <-ctx.Done():
			// shutdown signal recieved
			return nil
		}

		_, _, shutdown := reflect.Select(cases)
		if shutdown {
			return nil
		}
	}
}

// Process is called from any clients that pull messages from any of the channels passed into 'newTagGroup'
func (tg *tagGroup) Process() *v1.DequeueItemResponse {
	tg.lock.Lock()
	defer tg.lock.Unlock()

	index := tg.availableItems[0]
	tg.availableItems = tg.availableItems[1:]
	enqueuedItem := tg.items.Get(index).(*v1.EnqueueItem)

	return &v1.DequeueItemResponse{
		ID:   index,
		Name: enqueuedItem.Name,
		Tags: enqueuedItem.Tags,
		Data: enqueuedItem.Data,
	}
}

// Enqueue a new item onto the tag group.
func (tg *tagGroup) Enqueue(queueItem *v1.EnqueueItem) *v1.Error {
	tg.lock.Lock()
	defer tg.lock.Unlock()

	if len(tg.availableItems) >= 1 {
		lastItemId := tg.availableItems[len(tg.availableItems)-1]
		lastItem := tg.items.Get(lastItemId)
		lastQueueItem := lastItem.(*v1.EnqueueItem)

		// update the last item if we can. In this case, just return
		if lastQueueItem.Updateable == true {
			lastQueueItem.Data = queueItem.Data
			return nil
		}
	}

	tg.availableItems = append(tg.availableItems, tg.items.Add(queueItem))
	_ = tg.notifier.Add() // don't care about the error. on a shutdown message will be dropped anyways

	return nil
}

func (tg *tagGroup) StrictTag() <-chan tags.Tag {
	return tg.strictChannel
}

func (tg *tagGroup) Stop() {
	tg.doneOnce.Do(func() {
		close(tg.done)
	})
}
