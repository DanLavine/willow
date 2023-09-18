package memory

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/idgenerator"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/internal/datastructures/btree"
)

type tagGroup struct {
	done     chan struct{}
	doneOnce *sync.Once
	lock     *sync.Mutex

	idGenerator    idgenerator.UniqueIDs
	items          btree.BTree
	availableItems []string

	notifier       *gonotify.Notify
	dequeueChannel chan any // any -> func(logger *zap.Logger) *v1willow.DequeueItemResponse

	tags                datatypes.StringMap
	itemReadyCount      *atomic.Uint64
	itemProcessingCount *atomic.Uint64
}

func newTagGroup(tags datatypes.StringMap) *tagGroup {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &tagGroup{
		// shutdown
		lock:     new(sync.Mutex),
		done:     make(chan struct{}),
		doneOnce: new(sync.Once),

		// keeping track of items
		idGenerator:    idgenerator.UUID(),
		items:          tree,
		availableItems: []string{},

		// communication
		notifier:       gonotify.New(),
		dequeueChannel: make(chan any),

		// counters and info
		tags:                tags,
		itemReadyCount:      new(atomic.Uint64),
		itemProcessingCount: new(atomic.Uint64),
	}
}

// Handled by GoAsync to constantly read items from the queue and handle shutdown
func (tg *tagGroup) Execute(ctx context.Context) error {
	// ensure we close the channel for this tag group
	defer func() {
		defer tg.stop()
		close(tg.dequeueChannel)
		tg.notifier.ForceStop()
	}()

	// setup all possible casees
	cases := []reflect.SelectCase{
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tg.done)},
		{Dir: reflect.SelectSend, Chan: reflect.ValueOf(tg.dequeueChannel), Send: reflect.ValueOf(tg.dequeue)},
	}

	for {
		select {
		case <-tg.notifier.Ready():
			// we have an item for processing, so we can proceed to the select case
		case <-tg.done:
			// no more messages in this tag group, so closing the thread
			return nil
		case <-ctx.Done():
			// shutdown signal received
			return nil
		}

		// DSL: We cannot check here since there is a lot in flight before this is actually pulled

		_, _, shutdown := reflect.Select(cases)
		if shutdown {
			return nil
		}
	}
}

// Open questions:
// 1. When dequeueing an item, the caller should be able to "fail" an item if sending to the client
//    fails. This is important in a horizontal scalang problem, because a client could be connected
//    to multiple willow nodes, but only 1 can dequeue an item.
// 2. When an item is dequeued and it fails to send, should the "updates" for any messages come in
//    during that time processes, or are they considered a new queue item. This is a super small use
//    case, but I think it highlights something is off on how long it takes to process.
// 3. Still need to think about checking the "limits for things to process. Where does that happen?
// 4. What happens during a shutdown operation, where we need to know that we can send a response?
//
// Should this take in a "callback" to do the actuall write operation, that way we can handle all the fail
// logic at once. Then also we can check the "limiter" before the callback. On a Fail of the limiter, we
// would need to do what... re-queue the item, but pause the "notifier" and try again from the client's
// perspective? There are so many possible options.

// dequeue is called by the queue when a client pulls a message for processing
//
//	RETURNS:
//	- *v1willow.DequeueItemResponse - dequeue item
//	- func() - onSuccess callback if the caller was able to process the request
//	- func() - onFailure callback if the caller faild to process the request
//	- *v1willow.Error - error with the queue
func (tg *tagGroup) dequeue(logger *zap.Logger) (*v1willow.DequeueItemResponse, func(), func(), *api.Error) {
	logger = logger.Named("dequeue")

	// pull the first item to process
	tg.lock.Lock()
	itemID := tg.availableItems[0]

	var dequeueItem *v1willow.DequeueItemResponse
	onFind := func(item any) {
		enqueuedItem := item.(*v1willow.EnqueueItemRequest)

		// DSLL On this find, do we have to check the rules set if it canbe processed?
		// What happens if that is false, we woould want the client to still know to check other tag groups

		dequeueItem = &v1willow.DequeueItemResponse{
			BrokerInfo: v1willow.BrokerInfo{
				Name: enqueuedItem.Name,
				Tags: tg.tags,
			},
			ID:   itemID,
			Data: enqueuedItem.Data,
		}
	}

	if err := tg.items.Find(datatypes.String(itemID), onFind); err != nil {
		logger.Error("Failed to dequeue tag group item", zap.Error(err))
		panic(err)
	}

	if dequeueItem == nil {
		logger.Error("failed to dequeue the available item id. Dropping the ID", zap.String("id", itemID), zap.Uint64("ready count", tg.itemReadyCount.Load()), zap.Uint64("processing count", tg.itemProcessingCount.Load()))
		panic("shouldn't happen")
	}

	onSuccess := func() {
		tg.availableItems = tg.availableItems[1:]

		tg.itemProcessingCount.Add(1)
		tg.itemReadyCount.Add(^uint64(0))
		tg.lock.Unlock()
	}

	onFailure := func() {
		tg.notifier.Add() // re-add to the notifier so the item can be processed again
		tg.lock.Unlock()
	}

	return dequeueItem, onSuccess, onFailure, nil
}

// Enqueue a new item onto the tag group.
// If this was to be used in a "trigger", would need to return the tag group itself, to pull an item?
func (tg *tagGroup) Enqueue(logger *zap.Logger, totalQueueCounter *Counter, queueItem *v1willow.EnqueueItemRequest) *api.Error {
	logger = logger.Named("Enqueue")
	tg.lock.Lock()
	defer tg.lock.Unlock()

	if len(tg.availableItems) >= 1 {
		lastItemId := tg.availableItems[len(tg.availableItems)-1]

		updated := false
		onFind := func(item any) {
			lastQueueItem := item.(*v1willow.EnqueueItemRequest)

			// update the last item if we can. In this case, just return
			if lastQueueItem.Updateable == true {
				lastQueueItem.Data = queueItem.Data
				updated = true
			}
		}

		if err := tg.items.Find(datatypes.String(lastItemId), onFind); err != nil {
			logger.Error("failed to update an enqueued item", zap.Error(err))
		}

		if updated {
			return nil
		}
	}

	if totalQueueCounter.Add() {
		tg.itemReadyCount.Add(1)
		// record the new id
		newID := tg.idGenerator.ID()
		tg.availableItems = append(tg.availableItems, newID)

		// add the item to the btree for processing and inform the notifier that we have something to process
		tg.items.CreateOrFind(datatypes.String(newID), func() any { return queueItem }, func(item any) { panic("shouldn't call") })
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
// - bool - bool to indicate if the tag group should be removed (True if there are no more items processing or waiting to process)
// - *v1willow.Error - internal error encountered
func (tg *tagGroup) ACK(logger *zap.Logger, totalQueueCounter *Counter, ackItem *v1willow.ACK) *api.Error {
	logger = logger.Named("ACK")

	processed := false
	canDelete := func(item any) bool {
		defer func() { processed = true }()

		enqueuedItem := item.(*v1willow.EnqueueItemRequest)
		tg.itemProcessingCount.Add(^uint64(0))

		if ackItem.Passed {
			// remove the item from the queue
			totalQueueCounter.Decrement()
			return true
		} else {
			// requeue a failed item
			switch ackItem.RequeueLocation {
			case v1willow.RequeueFront:
				tg.lock.Lock()
				defer tg.lock.Unlock()
				tg.availableItems = append([]string{ackItem.ID}, tg.availableItems...)

				_ = tg.notifier.Add()
				tg.itemReadyCount.Add(1)
				return false
			case v1willow.RequeueEnd:
				tg.lock.Lock()
				defer tg.lock.Unlock()
				tg.availableItems = append(tg.availableItems, enqueuedItem.Name)

				_ = tg.notifier.Add()
				tg.itemReadyCount.Add(1)
				return false
			default:
				totalQueueCounter.Decrement()

				// not going to requeue
				return true
			}
		}
	}

	if err := tg.items.Delete(datatypes.String(ackItem.ID), canDelete); err != nil {
		logger.Error("failed to ack item", zap.String("id", ackItem.ID), zap.Error(err))
	}

	if !processed {
		return &api.Error{Message: fmt.Sprintf("ID %s does not exist for tag group: %v", ackItem.ID, ackItem.BrokerInfo.Tags), StatusCode: http.StatusBadRequest}
	}

	return nil
}

func (tg *tagGroup) Metrics() *v1willow.TagMetricsResponse {
	return &v1willow.TagMetricsResponse{
		Tags:       tg.tags,
		Ready:      tg.itemReadyCount.Load(),
		Processing: tg.itemProcessingCount.Load(),
	}
}

func (tg *tagGroup) stop() {
	tg.doneOnce.Do(func() {
		close(tg.done)
	})
}
