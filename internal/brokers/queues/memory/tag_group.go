package memory

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/brokers/tags"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"

	idtree "github.com/DanLavine/willow/internal/datastructures/id_tree"
)

type queuedItem struct {
	processing  bool
	enqueueItem *v1.EnqueueItemRequest
}

type TagGroup interface {
	// Call for processing messages from GoAsync
	Execute(ctx context.Context) error

	// Enqueue a new message or updatte the last message waiting to be processed
	Enqueue(queueItem *v1.EnqueueItemRequest) *v1.Error

	// Stop this queue
	Stop()
}

type tagGroup struct {
	lock     *sync.Mutex
	done     chan struct{}
	doneOnce *sync.Once

	items          *idtree.IDTree
	availableItems []uint64

	notifier *gonotify.Notify
	channels []chan tags.Tag

	tags                datatypes.StringMap
	itemReadyCount      *atomic.Uint64
	itemProcessingCount *atomic.Uint64
}

func newTagGroup(tags datatypes.StringMap, channels []chan tags.Tag) *tagGroup {
	return &tagGroup{
		// shutdown
		lock:     new(sync.Mutex),
		done:     make(chan struct{}),
		doneOnce: new(sync.Once),

		// keeping track of items
		items:          idtree.NewIDTree(),
		availableItems: []uint64{},

		// communication
		notifier: gonotify.New(),
		channels: channels,

		// counters and info
		tags:                tags,
		itemReadyCount:      new(atomic.Uint64),
		itemProcessingCount: new(atomic.Uint64),
	}
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
	enqueuedItem := tg.items.Get(index).(*queuedItem)
	enqueuedItem.processing = true

	// update counters
	tg.itemProcessingCount.Add(1)
	tg.itemReadyCount.Add(^uint64(0))

	return &v1.DequeueItemResponse{
		BrokerInfo: v1.BrokerInfo{
			Name: enqueuedItem.enqueueItem.BrokerInfo.Name,
			Tags: tg.tags,
		},
		ID:   index,
		Data: enqueuedItem.enqueueItem.Data,
	}
}

// Enqueue a new item onto the tag group.
func (tg *tagGroup) Enqueue(totalQueueCounter *Counter, queueItem *v1.EnqueueItemRequest) *v1.Error {
	tg.lock.Lock()
	defer tg.lock.Unlock()

	if len(tg.availableItems) >= 1 {
		lastItemId := tg.availableItems[len(tg.availableItems)-1]
		lastItem := tg.items.Get(lastItemId)
		lastQueueItem := lastItem.(*queuedItem)

		// update the last item if we can. In this case, just return
		if lastQueueItem.enqueueItem.Updateable == true {
			lastQueueItem.enqueueItem.Data = queueItem.Data
			return nil
		}
	}

	if totalQueueCounter.Add() {
		tg.itemReadyCount.Add(1)
		tg.availableItems = append(tg.availableItems, tg.items.Add(&queuedItem{processing: false, enqueueItem: queueItem}))
		_ = tg.notifier.Add() // don't care about the error. on a shutdown message will be dropped anyways

		return nil
	}

	return errors.MaxEnqueuedItems
}

// ACK is the response to a processing item.
//
// PARAMS:
// - totalQueueCounter - counter to keep track of the toatl processing items in the queue
// - ackItem - item to ack and all the details from the client
//
// RESPONSE:
// - bool - bool to indicate if the tag group should be removed (True if there are no more items processing)
// - *v1.Error - internal error encountered
func (tg *tagGroup) ACK(totalQueueCounter *Counter, ackItem *v1.ACK) (bool, *v1.Error) {
	tg.lock.Lock()
	defer tg.lock.Unlock()

	if item := tg.items.Get(ackItem.ID); item != nil {
		enqueuedItem := item.(*queuedItem)

		if enqueuedItem.processing {
			if ackItem.Passed {
				if item := tg.items.Remove(ackItem.ID); item != nil {
					// item was processed successfully
					totalQueueCounter.Decrement()
				}
			} else {
				//TODO
				// item failed, need to re-queue it at the begining
				if item := tg.items.Get(ackItem.ID); item != nil {
					tg.availableItems = append([]uint64{ackItem.ID}, tg.availableItems...)
					tg.itemReadyCount.Add(1)
				}
			}
		} else {
			return false, &v1.Error{Message: fmt.Sprintf("ID %d is not processing", ackItem.ID), StatusCode: http.StatusBadRequest}
		}
	} else {
		return false, &v1.Error{Message: fmt.Sprintf("ID %d does not exist for tag group: %v", ackItem.ID, ackItem.BrokerInfo.Tags), StatusCode: http.StatusBadRequest}
	}

	// both cases we remove the item from the processing count
	tg.itemProcessingCount.Add(^uint64(0))
	return (tg.itemReadyCount.Load() + tg.itemProcessingCount.Load()) == 0, nil
}

func (tg *tagGroup) Metrics() *v1.TagMetricsResponse {
	return &v1.TagMetricsResponse{
		Tags:       tg.tags,
		Ready:      tg.itemReadyCount.Load(),
		Processing: tg.itemProcessingCount.Load(),
	}
}

func (tg *tagGroup) Stop() {
	tg.doneOnce.Do(func() {
		close(tg.done)
	})
}

func tagGroupOnFindLock(item any) {
	tagGroup := item.(*tagGroup)
	tagGroup.lock.Lock()
}
