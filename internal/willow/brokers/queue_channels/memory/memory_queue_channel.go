package memory

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/idgenerator"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type memoryQueueChannel struct {
	// heartbeat manager to run the tasks async
	asyncManager goasync.AsyncTaskManager

	// callback to delete this channel from the channel client's perspective
	deleteChan     chan struct{}
	deleteOnce     *sync.Once
	deleteCallback func()

	// need these all over the place so save them off
	queueName        string
	channelKeyValues datatypes.KeyValues

	// check the limiter when:
	// 1. enquing an item to ensure we are under the max queue limit
	// 2. dequeuing to ensure the user didn't set up some limit
	limiterClient limiterclient.LimiterClient

	// notifieer is used to indicate that there is something to process
	notifier *gonotify.Notify

	// channel all items will be dequeued from
	dequeueChan         chan func(ctx context.Context) (*v1willow.Item, func(), func()) // func() (*v1willow.DequeueQueueItem, func(), func())
	dequeueResponseChan chan bool                                                       // indicates if there is an issue with the limiter on the last request. need to wait for this to be lower

	// all the saved items are a QueueItem
	idGenerator idgenerator.UniqueIDs
	items       btree.BTree

	itemsLock       *sync.RWMutex
	itemIDsEnqueued []string
}

func New(limiterClient limiterclient.LimiterClient, deleteCallback func(), queueName string, channelKeyValues datatypes.KeyValues) *memoryQueueChannel {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	if deleteCallback == nil {
		panic("delete callback can not be nil")
	}

	return &memoryQueueChannel{
		asyncManager: goasync.NewTaskManager(goasync.RelaxedConfig()),

		deleteChan:     make(chan struct{}),
		deleteOnce:     new(sync.Once),
		deleteCallback: deleteCallback,

		queueName:        queueName,
		channelKeyValues: channelKeyValues,

		limiterClient:       limiterClient,
		notifier:            gonotify.New(),
		dequeueChan:         make(chan func(ctx context.Context) (*v1willow.Item, func(), func())),
		dequeueResponseChan: make(chan bool),

		idGenerator: idgenerator.UUID(),
		items:       tree,

		itemsLock:       new(sync.RWMutex),
		itemIDsEnqueued: []string{},
	}
}

// this is write loked from the client in a "Destroy" call
func (mqc *memoryQueueChannel) Delete() bool {
	if mqc.items.Empty() {
		mqc.deleteOnce.Do(func() {
			close(mqc.deleteChan)
		})

		return true
	}

	return false
}

// force delete is used when a channel is being destroyed and we do not care about the channel being empty.
// in this case, the channel should always just be destroyed
func (mqc *memoryQueueChannel) ForceDelete(ctx context.Context) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "ForceDelete")

	// stop processing on this channel
	mqc.deleteOnce.Do(func() {
		close(mqc.deleteChan)
	})

	// delete the running and enqueued counters
	if err := mqc.setLimiterEnqueuedValue(ctx); err != nil {
		logger.Fatal("TODO fix this panic error with a queue of Limiter values to retry", zap.Error(err))
	}

	if err := mqc.setLimterRunningValue(ctx); err != nil {
		logger.Fatal("TODO fix this panic error with a queue of Limiter values to retry", zap.Error(err))
	}

	canDelete := func(key datatypes.EncapsulatedValue, treeItem any) bool {
		queueItem := treeItem.(*item)

		// always try to just stop any running heartbeaters. The clients should be responsible for stopping
		// items that are processing
		_ = queueItem.StopHeartbeater()

		return true
	}

	if err := mqc.items.DestroyAll(canDelete); err != nil {
		panic(err)
	}
}

// Handler for the GoAsync manager to ensure this process stops processing when the server is shutdown
func (mqc *memoryQueueChannel) Execute(ctx context.Context) error {
	// run our own async manager for the heartbeats
	asyncCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stoppedHeartbeating := make(chan struct{})
	go func() {
		defer close(stoppedHeartbeating)
		_ = mqc.asyncManager.Run(asyncCtx)
	}()

	for {
		select {
		case <-ctx.Done():
			// server told to shutdown
			goto BREAK_DEQUEUE
		case <-mqc.deleteChan:
			// delete of the channel itself because there are no more items, or the queue is being destroyed
			cancel()
			goto BREAK_DEQUEUE
		case <-mqc.notifier.Ready():
			// we received an item to enqueue, so try to send on this channel
		}

		select {
		case <-ctx.Done():
			// server was told to shutdown while waiting for a client to process
			goto BREAK_DEQUEUE
		case <-mqc.deleteChan:
			// delete of the channel itself because there are no more items, or the queue is being destroyed
			cancel()
			goto BREAK_DEQUEUE
		case mqc.dequeueChan <- mqc.dequeue:
			// sent an item that is going to be dequeued by a client

			// 1. need to wait for failure to dequeue, success or failure to process
			limitReached := <-mqc.dequeueResponseChan

			// 2. if there was an issue with the limit reached. just be simple for now and spin here.
			// TODO DSL: add a 'watch' api on the Limiter to know if this can succeed again
			if limitReached {
				erroredKeyValues := datatypes.KeyValues{
					"_willow_queue_name": datatypes.String(mqc.queueName),
					"_willow_running":    datatypes.String("true"),
				}
				for key, value := range mqc.channelKeyValues {
					erroredKeyValues[key] = value
				}

				for {
					// find any rules we might be at the limit for
					rules, err := mqc.limiterClient.MatchRules(context.Background(), querymatchaction.KeyValuesToAnyMatchActionQuery(erroredKeyValues))
					if err != nil {
						panic(err)
					}

					// for each rule, query the counters to know if there is an issue
					underLimit := true
					for _, rule := range rules {
						overrides, err := mqc.limiterClient.MatchOverrides(context.Background(), *rule.Spec.DBDefinition.Name, querymatchaction.KeyValuesToAnyMatchActionQuery(erroredKeyValues))
						if err != nil {
							panic(err)
						}

						switch len(overrides) {
						case 0:
							// check the rule itsel

							// somethign is at a limit of 0 so just stop processing
							if *rule.Spec.Properties.Limit == 0 {
								underLimit = false
								break
							}

							// unlimited so skip this one
							if *rule.Spec.Properties.Limit == -1 {
								continue
							}

							// setup query for the counters based off of the Rule's key values we are looking at
							query := &queryassociatedaction.AssociatedActionQuery{
								Selection: &queryassociatedaction.Selection{
									KeyValues: queryassociatedaction.SelectionKeyValues{},
								},
							}

							for key, value := range rule.Spec.DBDefinition.GroupByKeyValues {
								query.Selection.KeyValues[key] = queryassociatedaction.ValueQuery{
									Value:      value,
									Comparison: v1common.Equals,
									TypeRestrictions: v1common.TypeRestrictions{
										MinDataType: value.Type,
										MaxDataType: value.Type,
									},
								}
							}

							counters, err := mqc.limiterClient.QueryCounters(context.Background(), query)
							if err != nil {
								panic(err)
							}

							totalCount := int64(0)
							for _, counter := range counters {
								totalCount += *counter.Spec.Properties.Counters
								if totalCount >= *rule.Spec.Properties.Limit {
									underLimit = false
									break
								}
							}
						default:
							// check all the overrides

							for _, override := range overrides {
								// somethign is at a limit of 0 so just stop processing
								if *override.Spec.Properties.Limit == 0 {
									underLimit = false
									break
								}

								// unlimited so skip this one
								if *override.Spec.Properties.Limit == -1 {
									continue
								}

								// setup query for the counters based off of the Rule's key values we are looking at
								query := &queryassociatedaction.AssociatedActionQuery{
									Selection: &queryassociatedaction.Selection{
										KeyValues: queryassociatedaction.SelectionKeyValues{},
									},
								}

								for key, value := range override.Spec.DBDefinition.GroupByKeyValues {
									query.Selection.KeyValues[key] = queryassociatedaction.ValueQuery{
										Value:      value,
										Comparison: v1common.Equals,
										TypeRestrictions: v1common.TypeRestrictions{
											MinDataType: value.Type,
											MaxDataType: value.Type,
										},
									}
								}

								counters, err := mqc.limiterClient.QueryCounters(context.Background(), query)
								if err != nil {
									panic(err)
								}

								totalCount := int64(0)
								for _, counter := range counters {
									totalCount += *counter.Spec.Properties.Counters
									if totalCount >= *override.Spec.Properties.Limit {
										underLimit = false
										break
									}
								}
							}
						}

						// can stop processing, still over the limit
						if !underLimit {
							break
						}
					}

					// if under the limit, break out of this loop and try to dequeue an item again
					if underLimit {
						break
					}

					// check to delete before attempting another retry
					select {
					case <-ctx.Done():
						// server was told to shutdown while checking
						goto BREAK_DEQUEUE
					case <-mqc.deleteChan:
						// delete of the channel itself because there are no more items, or the queue is being destroyed
						cancel()
						goto BREAK_DEQUEUE
					case <-time.After(5 * time.Second):
					}
				}
			}
		}
	}

BREAK_DEQUEUE:

	close(mqc.dequeueChan)
	mqc.notifier.ForceStop()

	// wait for any heartbeat operations to finish
	<-stoppedHeartbeating

	return nil
}

// Try to Enqueue an item and record on the limiter what is being saved
func (mqc *memoryQueueChannel) Enqueue(ctx context.Context, enqueueItem *v1willow.Item) *errors.ServerError {
	ctx, _ = middleware.GetNamedMiddlewareLogger(ctx, "Enqueue")

	mqc.itemsLock.Lock() // need this lock so multiple enqueue requests can all be squashed into 1
	defer mqc.itemsLock.Unlock()

	lastItemIndex := len(mqc.itemIDsEnqueued) - 1

	// attempt to update the last item enqueued
	if lastItemIndex >= 0 {
		lastItemID := mqc.itemIDsEnqueued[lastItemIndex]

		updated := false
		onFind := func(key datatypes.EncapsulatedValue, treeItem any) bool {
			queueItem := treeItem.(*item)
			queueItem.lock.Lock()
			defer queueItem.lock.Unlock()

			// update the last item
			if queueItem.updateable {
				updated = true

				queueItem.data = enqueueItem.Spec.Properties.Data
				queueItem.updateable = *enqueueItem.Spec.Properties.Updateable
				queueItem.maxRetryAttempts = *enqueueItem.Spec.Properties.RetryAttempts
				queueItem.retryPosition = *enqueueItem.Spec.Properties.RetryPosition
				queueItem.retryCount = 0
			}

			return false
		}

		// try to update the last item enqueud
		if err := mqc.items.Find(datatypes.String(lastItemID), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
			panic(err)
		}

		if updated {
			return nil
		}
	}

	// need to create the new item and append it to the list

	// ensure the limits are not reached
	if err := mqc.limiterUpdateEnqueuedValue(ctx, 1); err != nil {
		return err
	}

	// create the new item in the channel
	newId := mqc.idGenerator.ID()
	onCreate := func() any {
		return newItem(
			enqueueItem.Spec.Properties.Data,
			*enqueueItem.Spec.Properties.Updateable,
			*enqueueItem.Spec.Properties.RetryAttempts,
			*enqueueItem.Spec.Properties.RetryPosition,
			*enqueueItem.Spec.Properties.TimeoutDuration,
		)
	}

	if err := mqc.items.Create(datatypes.String(newId), onCreate); err != nil {
		panic(err)
	}

	// add the item id to the list of processing items
	mqc.itemIDsEnqueued = append(mqc.itemIDsEnqueued, newId)

	// signal to the notifier that we have something to process
	_ = mqc.notifier.Add() // in the case of an error we are shutting down so just drop it

	return nil
}

//	PARAMETERS:
//	- *zapLogger - logger for the operation
//	- *ack - api model with all the detals for the ACK operation
//
//	RETURNS:
//	- bool - indicates if the entire tree can be removed
//	- *errors.ServerError - api error if one is encountered when acking an item
//
// ACK an item. On successful, the item is removed entierly. On a failure, the itemwill try to be requeued.
//
// NOTE: Write locked from the queue_channel_client
func (mqc *memoryQueueChannel) ACK(ctx context.Context, ack *v1willow.ACK) (bool, *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "ACK")
	ackErr := &errors.ServerError{Message: "failed to find processing item by id", StatusCode: http.StatusNotFound}

	switch ack.Passed {
	case true:
		canDelete := func(_ datatypes.EncapsulatedValue, treeItem any) bool {
			queueItem := treeItem.(*item)

			// if the queue item was stopped here, then we know there was no async timeout processed for this item
			if queueItem.StopHeartbeater() {
				// 2. always update counters that the item is no loger running
				if err := mqc.limterUpdateRunningValue(ctx, -1); err != nil {
					// what should be the actual course of action here?
					panic(err)
				}

				// 3. update counters that an enqueued item completed
				if err := mqc.limiterUpdateEnqueuedValue(ctx, -1); err != nil {
					// what should we really do here? the limiter would be out of sync in this case
					panic(err)
				}

				logger.Debug("removed item from the channel")
				ackErr = nil
				return true
			}

			logger.Debug("failed to remove the item since it is not processing")
			return false
		}

		if err := mqc.items.Delete(datatypes.String(ack.ItemID), canDelete); err != nil {
			logger.Error("failed to ack delete an item", zap.Error(err))
			panic(err)
		}
	default:
		if mqc.failItem(ctx, ack.ItemID, false) {
			ackErr = nil
		}
	}

	return mqc.items.Empty(), ackErr
}

func (mqc *memoryQueueChannel) failItem(ctx context.Context, itemID string, timedOut bool) bool {
	ctx, _ = middleware.GetNamedMiddlewareLogger(ctx, "failItem")

	mqc.itemsLock.Lock()
	defer mqc.itemsLock.Unlock()

	attemptedDelete := false
	backID := ""

	// attempt to delete or requeue the item
	canDelete := func(_ datatypes.EncapsulatedValue, treeItem any) bool {
		attemptedDelete = true
		queueItem := treeItem.(*item)

		if timedOut {
			queueItem.UnsetHeartbeater()
		}

		// if the queue item was stopped here, then we know there was no async timeout processed for this item
		if timedOut || queueItem.StopHeartbeater() {
			// always update counters that the item is no longer running
			if err := mqc.limterUpdateRunningValue(ctx, -1); err != nil {
				// what should be the actual course of action here?
				panic(err)
			}

			queueItemToDelete := treeItem.(*item)
			queueItemToDelete.lock.Lock()
			defer queueItemToDelete.lock.Unlock()

			queueItemToDelete.retryCount++

			// hit the max retry attempts for the queue item, so remove the item from the queue
			if queueItemToDelete.retryCount > queueItemToDelete.maxRetryAttempts {
				// when removing an item. we need to delete the total number of enqueued item
				if err := mqc.limiterUpdateEnqueuedValue(ctx, -1); err != nil {
					// what should we really do here? the limiter would be out of sync in this case
					panic(err)
				}

				return true
			}

			// must requeue the item for processing
			switch queueItemToDelete.retryPosition {
			case "front":
				if len(mqc.itemIDsEnqueued) >= 1 {
					// just delete the item. since it is updateable, we want the next item in the queue to run anyways
					if queueItemToDelete.updateable {
						return true
					}
				}

				// always append to the front
				mqc.itemIDsEnqueued = append([]string{itemID}, mqc.itemIDsEnqueued...)
				mqc.notifier.Add()

				return false
			case "back":
				if len(mqc.itemIDsEnqueued) >= 1 {
					backID = mqc.itemIDsEnqueued[len(mqc.itemIDsEnqueued)-1]
				} else {
					mqc.itemIDsEnqueued = append(mqc.itemIDsEnqueued, itemID)
					mqc.notifier.Add()
				}
			}
		}

		return false
	}

	if err := mqc.items.Delete(datatypes.String(itemID), canDelete); err != nil {
		panic(err)
	}

	// need to check the last enqueued item to see if it can be dropped
	if backID != "" {
		// check to see if we can delete the previous item in the enqueued list. Logicialy
		// this is the same as updating the last enueued item
		canDeleteLastItem := func(_ datatypes.EncapsulatedValue, treeItem any) bool {
			queueItemToCheck := treeItem.(*item)
			queueItemToCheck.lock.Lock()
			defer queueItemToCheck.lock.Unlock()

			if queueItemToCheck.updateable {
				// "update" the last item by simply dropping it
				mqc.itemIDsEnqueued[len(mqc.itemIDsEnqueued)-1] = itemID
				return true
			} else {
				// "append" to the list the item that failed
				mqc.itemIDsEnqueued = append(mqc.itemIDsEnqueued, itemID)
				mqc.notifier.Add()
				return false
			}
		}

		if err := mqc.items.Delete(datatypes.String(backID), canDeleteLastItem); err != nil {
			panic(err)
		}
	}

	return attemptedDelete
}

// Dequeue returns a read only channel that can be cast to 'func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())'
// From here, we can make some assumptions aboout the reurned values
//
//	RETURNS:
//	- *v1willow.DequeueQueueItem - will be present IFF there were no limits encountered with this item fom the Limiter service
//	- func() - success callback to run if we successfully respond to the client through the http.ResponseWriter
//	- func() - failure callback to run if we fail to responde to the client through the http.ResponseWriter
//
// The callback functions are used to ensure that the item being processed is gurranted to at least once be sent
// to a client
func (mqc *memoryQueueChannel) Dequeue() <-chan func(ctx context.Context) (*v1willow.Item, func(), func()) {
	return mqc.dequeueChan
}

// callback passed to the 'dequeueChan' when there is something to dequeue
func (mqc *memoryQueueChannel) dequeue(ctx context.Context) (*v1willow.Item, func(), func()) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "dequeue")

	// 1. ensure that the item can be dequeued when running. This just forwards the key values that define the channel
	if err := mqc.limterUpdateRunningValue(ctx, 1); err != nil {
		logger.Error("failed to update the counter for the queue item", zap.Error(err))

		// re-add to the notifier since we failed to process this item properly
		_ = mqc.notifier.Add()

		// resond to the queue forwarder that there is work to do checking that this limit is no longer reached
		mqc.dequeueResponseChan <- true
		return nil, nil, nil
	}

	// drop the first index since we are now processing it
	mqc.itemsLock.Lock()
	firtItemID := mqc.itemIDsEnqueued[0]
	mqc.itemIDsEnqueued = mqc.itemIDsEnqueued[1:]
	mqc.itemsLock.Unlock()

	// 2. successfully incremented the counters, pull an item off for the client
	dequeueItem := &v1willow.Item{}

	onFind := func(key datatypes.EncapsulatedValue, treeItem any) bool {
		queueItem := treeItem.(*item)

		dequeueItem = &v1willow.Item{
			Spec: &v1willow.ItemSpec{
				DBDefinition: &v1willow.ItemDBDefinition{
					KeyValues: mqc.channelKeyValues,
				},
				Properties: &v1willow.ItemProperties{
					Data:            queueItem.data,
					Updateable:      &queueItem.updateable,
					RetryAttempts:   &queueItem.maxRetryAttempts,
					RetryPosition:   &queueItem.retryPosition,
					TimeoutDuration: &queueItem.heartbeatTimeout,
				},
			},
			State: &v1willow.ItemState{
				ID: firtItemID,
			},
		}

		// create the heartbeat operation in the background

		// the shutdown function should remove the limiter values since this is in memory only
		onShutdown := func() {
			// TODO. what should this really be?
		}

		// The timeout function is the same behavior as a failed ACK operation + the parent callback to try and destroy this queue channel
		onTimeout := func() {
			_ = mqc.failItem(reporting.StripedContext(logger), firtItemID, true)

			// if this times out, call the client to try and delete this channel
			mqc.deleteCallback()
		}

		if err := mqc.asyncManager.AddExecuteTask(firtItemID, queueItem.CreateHeartbeater(onShutdown, onTimeout)); err != nil {
			// failing to add happens if shutting down the server or destroying the queue happens async. In that case
			// we want to ensure our resources are cleaned up properly
			queueItem.UnsetHeartbeater()
		}

		return false
	}

	if err := mqc.items.Find(datatypes.String(firtItemID), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		panic(err)
	}

	return dequeueItem, mqc.successfulDequeue(ctx, firtItemID), mqc.failedDequeue(ctx, firtItemID)
}

// callback passed to the 'dequeueChan' and called when the client successfully recieved the item
func (mqc *memoryQueueChannel) successfulDequeue(ctx context.Context, itemID string) func() {
	return func() {
		_, logger := middleware.GetNamedMiddlewareLogger(ctx, "successfulDequeue")

		// start the heartbeater process
		onfindBTree := func(key datatypes.EncapsulatedValue, treeItem any) bool {
			queueItem := treeItem.(*item)

			if queueItem.StartHeartbeater() {
				logger.Debug("started the heartbeat process")
			} else {
				// This can happen if a call to ACK happens and processes before the success callback is called in the API controller
				logger.Debug("failed to start the heartbeat process")
			}

			return false
		}

		if err := mqc.items.Find(datatypes.String(itemID), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onfindBTree); err != nil {
			panic(err)
		}

		// resond to the queue forwarder that it can start processing the next item in the queue
		mqc.dequeueResponseChan <- false
	}
}

// callback passed to the 'dequeueChan' and called wheh the client failed to send the dequeue response to the client
func (mqc *memoryQueueChannel) failedDequeue(ctx context.Context, itemID string) func() {
	return func() {
		ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "failedDequeue")

		// update the running counter that we are no longer processing since we failed to dequeue to the remote client
		if err := mqc.limterUpdateRunningValue(ctx, -1); err != nil {
			panic(err)
		}

		mqc.itemsLock.Lock()
		defer mqc.itemsLock.Unlock()

		// requeue the item
		canDelete := func(_ datatypes.EncapsulatedValue, treeItem any) bool { // i think this should be canDestroy?
			queueItem := treeItem.(*item)
			queueItem.lock.Lock() // don't think I need this, but leave it here for now
			defer queueItem.lock.Unlock()

			// stop the heartbeater process
			if queueItem.StopHeartbeater() {
				logger.Debug("stopped the heartbeat process")

				// if the queue item is updateable, check to see if there is something else in the queeu
				if queueItem.updateable {
					if len(mqc.itemIDsEnqueued) >= 1 {
						// in this case there is something else in the queue that would have updated the item. so just toss this item away
						if err := mqc.limiterUpdateEnqueuedValue(ctx, -1); err != nil {
							panic(err)
						}

						return true
					}
				}

				// always put the item at the front of the queue to process again
				mqc.itemIDsEnqueued = append([]string{itemID}, mqc.itemIDsEnqueued...)
				mqc.notifier.Add() // indicate to the notifier that there is something to process
			} else {
				// this should never happen!
				logger.Debug("failed to stop the heartbeat process")
			}

			return false
		}

		if err := mqc.items.Delete(datatypes.String(itemID), canDelete); err != nil {
			panic(err)
		}

		// resond to the queue forwarder that it can start processing the next item in the queue
		mqc.dequeueResponseChan <- false
	}
}

func (mqc *memoryQueueChannel) Heartbeat(ctx context.Context, heartbeat *v1willow.Heartbeat) *errors.ServerError {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "Heartbeat")
	heartbeatErr := &errors.ServerError{Message: "failed to find processing item by id", StatusCode: http.StatusNotFound}

	onFind := func(key datatypes.EncapsulatedValue, treeItem any) bool {
		queueItem := treeItem.(*item)

		if queueItem.Heartbeat() {
			heartbeatErr = nil
		}

		return false
	}

	if err := mqc.items.Find(datatypes.String(heartbeat.ItemID), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		logger.Fatal("failed to lookup item to heartbeat", zap.Error(err))
	}

	// at this point, if this is an error, there should be a debug log that this timed out previously
	return heartbeatErr
}

// limiterUpdateEnqueuedValue is used when an item is enqueued or removed from the channel. This keeps track
// of the total 'enqueued' items for a queue and rejects when to many items are being added to the queue
func (mqc *memoryQueueChannel) limiterUpdateEnqueuedValue(ctx context.Context, counterUpdate int64) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "limiterUpdateEnqueuedValue")

	enqueueKeyValues := datatypes.KeyValues{
		"_willow_queue_name": datatypes.String(mqc.queueName),
		"_willow_enqueued":   datatypes.String("true"),
	}
	for key, value := range mqc.channelKeyValues {
		enqueueKeyValues[fmt.Sprintf("_willow_%s", key)] = value
	}

	err := mqc.limiterClient.UpdateCounter(ctx, &v1limiter.Counter{
		Spec: &v1limiter.CounterSpec{
			DBDefinition: &v1limiter.CounterDBDefinition{
				KeyValues: enqueueKeyValues,
			},
			Properties: &v1limiter.CounteProperties{
				Counters: &counterUpdate,
			},
		},
	})

	if err != nil {
		logger.Warn("hit a limit with the total number of enqued items", zap.Error(err))
		return &errors.ServerError{Message: "Queue has reached the total number of allowed queue items", StatusCode: http.StatusConflict}
	}

	return nil
}

// limterUpdateRunningValue is used when an item is dequeued channel. This keeps track
// of the total 'running' items for a queue and rejects when a 3rd paarty rule has reached the limit setup from a user
func (mqc *memoryQueueChannel) limterUpdateRunningValue(ctx context.Context, counterUpdate int64) error {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "limiterUpdateRunningValue")

	counterKeyValues := datatypes.KeyValues{
		"_willow_queue_name": datatypes.String(mqc.queueName),
		"_willow_running":    datatypes.String("true"),
	}
	for key, value := range mqc.channelKeyValues {
		counterKeyValues[key] = value
	}

	counter := &v1limiter.Counter{
		Spec: &v1limiter.CounterSpec{
			DBDefinition: &v1limiter.CounterDBDefinition{
				KeyValues: counterKeyValues,
			},
			Properties: &v1limiter.CounteProperties{
				Counters: &counterUpdate,
			},
		},
	}

	if err := mqc.limiterClient.UpdateCounter(ctx, counter); err != nil {
		logger.Warn("hit a limit for the total number of runnable items", zap.Error(err))
		return err
	}

	return nil
}

func (mqc *memoryQueueChannel) setLimiterEnqueuedValue(ctx context.Context) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "setLimiterEnqueuedValue")

	enqueueKeyValues := datatypes.KeyValues{
		"_willow_queue_name": datatypes.String(mqc.queueName),
		"_willow_enqueued":   datatypes.String("true"),
	}
	for key, value := range mqc.channelKeyValues {
		enqueueKeyValues[fmt.Sprintf("_willow_%s", key)] = value
	}

	err := mqc.limiterClient.SetCounters(ctx, &v1limiter.Counter{
		Spec: &v1limiter.CounterSpec{
			DBDefinition: &v1limiter.CounterDBDefinition{
				KeyValues: enqueueKeyValues,
			},
			Properties: &v1limiter.CounteProperties{
				Counters: helpers.PointerOf[int64](0),
			},
		},
	})

	if err != nil {
		logger.Warn("failed to remove the enqueued counters values", zap.Error(err))
		return &errors.ServerError{Message: "Failed to set the counters properly for enqueued channel", StatusCode: http.StatusInternalServerError}
	}

	return nil
}

func (mqc *memoryQueueChannel) setLimterRunningValue(ctx context.Context) error {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "setLimiterRunningValue")

	counterKeyValues := datatypes.KeyValues{
		"_willow_queue_name": datatypes.String(mqc.queueName),
		"_willow_running":    datatypes.String("true"),
	}
	for key, value := range mqc.channelKeyValues {
		counterKeyValues[key] = value
	}

	counter := &v1limiter.Counter{
		Spec: &v1limiter.CounterSpec{
			DBDefinition: &v1limiter.CounterDBDefinition{
				KeyValues: counterKeyValues,
			},
			Properties: &v1limiter.CounteProperties{
				Counters: helpers.PointerOf[int64](0),
			},
		},
	}

	if err := mqc.limiterClient.SetCounters(ctx, counter); err != nil {
		logger.Warn("hit a limit for the total number of runnable items", zap.Error(err))
		return &errors.ServerError{Message: "Failed to set the counters properly for running channel", StatusCode: http.StatusInternalServerError}
	}

	return nil
}
