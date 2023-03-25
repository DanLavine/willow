package memory

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/v1/tags"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type TagGroup interface {
	// Call for processing messages from GoAsync
	Execute(ctx context.Context) error

	// Enqueue a new message or updatte the last message waiting to be processed
	Enqueue(queueItem *v1.EnqueueItem) *v1.Error

	// Stop this queue
	Stop()
}

type tagGroup struct {
	lock     *sync.Mutex
	done     chan struct{}
	doneOnce *sync.Once

	items          *datastructures.IDTree
	availableItems []uint64

	notifier *gonotify.Notify
	channels []chan<- tags.Tag
}

func newTagGroup(channels []chan<- tags.Tag) *tagGroup {
	return &tagGroup{
		// shutdown
		lock:     new(sync.Mutex),
		done:     make(chan struct{}),
		doneOnce: new(sync.Once),

		// keeping track of items
		items:          datastructures.NewIDTree(),
		availableItems: []uint64{},

		// communication
		notifier: gonotify.New(),
		channels: channels,
	}
}

func (tg *tagGroup) OnFind() {
	tg.lock.Lock()
}

func (tg *tagGroup) Unlock() {
	tg.lock.Unlock()
}

// Handled by GoAsync to constantly read items from the queue and handle shutdown
func (tg *tagGroup) Execute(ctx context.Context) error {
	defer func() {
		tg.Stop()
		tg.notifier.ForceStop()
	}()

	cases := []reflect.SelectCase{
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tg.done)}}

	// setup all possible channels
	for _, channel := range tg.channels {
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(channel), Send: reflect.ValueOf(tg.process)})
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

// process is called from any clients that pull messages from any of the channels passed into 'newTagGroup'
func (tg *tagGroup) process() *v1.DequeueItemResponse {
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
func (tg *tagGroup) Enqueue(queueItem *v1.EnqueueItem, readyCount *atomic.Uint64) *v1.Error {
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

	readyCount.Add(uint64(1))
	tg.availableItems = append(tg.availableItems, tg.items.Add(queueItem))
	_ = tg.notifier.Add() // don't care about the error. on a shutdown message will be dropped anyways

	return nil
}

func (tg *tagGroup) Stop() {
	tg.doneOnce.Do(func() {
		close(tg.done)
	})
}
