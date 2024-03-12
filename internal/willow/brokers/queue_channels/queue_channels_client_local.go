package queuechannels

import (
	"context"
	"net/http"
	"sync"

	"github.com/DanLavine/channelops"
	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/internal/willow/brokers/queue_channels/constructor"

	btreeonetomany "github.com/DanLavine/willow/internal/datastructures/btree_one_to_many"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type clientWaiting struct {
	// queue the client is waiting on
	queueName string
	// query is used to match any channels against to see if they can provide values for the client
	query datatypes.AssociatedKeyValuesQuery
	// channelOPS is the collection of channels attempting to be read
	channelOPS channelops.RepeatableMergeReadChannelOperator[func(ctx context.Context) (*v1willow.DequeueQueueItem, func(), func())]
}

type queueChannelsClientLocal struct {
	// used to cancel any waiting/blocked DequeueQueueItem clients
	shutdownCtx    context.Context
	shutdownCancel func()

	// channel manager to run the tasks async
	asyncManager goasync.AsyncTaskManager

	// queue constructor for creating and managing queues
	queueChannelsConstructor constructor.QueueChannelsConstrutor

	// all possible queues a user created
	queueChannels btreeonetomany.BTreeOneToMany

	// client wating resources
	clientsWaitingLock *sync.RWMutex
	clientsWaiting     []clientWaiting
}

func NewLocalQueueChannelsClient(queueChannelsConstructor constructor.QueueChannelsConstrutor) *queueChannelsClientLocal {
	shutdownContext, cancel := context.WithCancel(context.Background())

	return &queueChannelsClientLocal{
		shutdownCtx:              shutdownContext,
		shutdownCancel:           cancel,
		asyncManager:             goasync.NewTaskManager(goasync.RelaxedConfig()),
		queueChannelsConstructor: queueChannelsConstructor,
		queueChannels:            btreeonetomany.NewThreadSafe(),
		clientsWaitingLock:       new(sync.RWMutex),
		clientsWaiting:           []clientWaiting{},
	}
}

func (qccl *queueChannelsClientLocal) Execute(ctx context.Context) error {
	done := make(chan struct{})

	// allow all queue channels to be added to the channel manager when they are created
	go func() {
		defer close(done)
		_ = qccl.asyncManager.Run(ctx)
	}()

	// triggered on a server shutdown call
	<-ctx.Done()

	// close the shutdown context to inform any other calls that the services is stopping
	qccl.shutdownCancel()

	// wait for all queue channels to be cleaned up
	<-done

	return nil
}

func (qccl *queueChannelsClientLocal) DestroyChannelsForQueue(ctx context.Context, queueName string) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("DestroyChannelsForQueue")
	ctx = reporting.UpdateLogger(ctx, logger)

	deleteChannel := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
		queueChannel := oneToManyItem.Value().(constructor.QueueChannel)
		queueChannel.ForceDelete(ctx)

		return true
	}

	if err := qccl.queueChannels.DestroyOne(queueName, deleteChannel); err != nil {
		switch err {
		case btreeonetomany.ErrorManyIDDestroying:
			logger.Debug("Already destroying the queue's channels")
			return &errors.ServerError{Message: "queue channels already destroying", StatusCode: http.StatusNoContent}
		default:
			logger.Fatal("Failed to delete channels for queue", zap.Error(err))
		}
	}

	return nil
}

func (qccl *queueChannelsClientLocal) DeleteChannel(ctx context.Context, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("DeleteChannel").With(zap.String("queue_name", queueName), zap.Any("channel_key_values", channelKeyValues))
	ctx = reporting.UpdateLogger(ctx, logger)

	// delete a channel only if there are no enqueued items
	deleteChannel := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
		queueChannel := oneToManyItem.Value().(constructor.QueueChannel)
		queueChannel.ForceDelete(ctx)

		return true
	}

	if err := qccl.queueChannels.DeleteOneOfManyByKeyValues(queueName, channelKeyValues, deleteChannel); err != nil {
		switch err {
		case btreeonetomany.ErrorManyIDDestroying:
			logger.Debug("Already destroying the queue's channel")
			return &errors.ServerError{Message: "Already destroying the queue's channel", StatusCode: http.StatusConflict}
		default:
			logger.Fatal("Failed to delete channels for queue", zap.Error(err))
		}
	}

	return nil
}

func (qccl *queueChannelsClientLocal) attemptDeleteChannel(logger *zap.Logger, queueName string, channelKeyValues datatypes.KeyValues) {
	logger = logger.Named("attemptDeleteChannel").With(zap.String("queue_name", queueName), zap.Any("channel_key_values", channelKeyValues))

	// delete a channel only if there are no enqueued items
	deleteChannel := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
		queueChannel := oneToManyItem.Value().(constructor.QueueChannel)
		return queueChannel.Delete()
	}

	if err := qccl.queueChannels.DeleteOneOfManyByKeyValues(queueName, channelKeyValues, deleteChannel); err != nil {
		switch err {
		case btreeonetomany.ErrorManyIDDestroying:
			logger.Debug("Already destroying the queue's channel on a timeout")
		default:
			logger.Fatal("Failed to delete channels for queue", zap.Error(err))
		}
	}
}

func (qccl *queueChannelsClientLocal) EnqueueQueueItem(ctx context.Context, queueName string, enqueueItem *v1willow.EnqueueQueueItem) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("EnqueueQueueItem").With(zap.String("queue_name", queueName), zap.Any("channel_key_values", enqueueItem.KeyValues))
	ctx = reporting.UpdateLogger(ctx, logger)
	var enqueueError *errors.ServerError

	//create a new channel to enqueue items to
	bTreeOneToManyOnCreate := func() any {
		destroyCallback := func() {
			// on a timeout we can attempt to delete the channel
			qccl.attemptDeleteChannel(reporting.BaseLogger(logger), queueName, enqueueItem.KeyValues)
		}

		queueChannel := qccl.queueChannelsConstructor.New(destroyCallback, queueName, enqueueItem.KeyValues)
		enqueueError = queueChannel.Enqueue(ctx, enqueueItem)

		// break early because we failed to enqueue the item and return nil because nothing was saved
		if enqueueError != nil {
			return nil
		}

		// when a new channel is created, need to add it to the async task manager. if there is an error, that means the
		// server is shutting down so don't add it to any waiting clients.
		if err := qccl.asyncManager.AddExecuteTask(queueName, queueChannel); err == nil {
			// when a new channel is created, inform any clients currently waiting to process that there is something they might care about
			qccl.updateClientsWaiting(queueName, enqueueItem.KeyValues, queueChannel.Dequeue())
		}

		return queueChannel
	}

	// enqueue an item to an already existing channel
	bTreeOneToManyOnFind := func(item btreeonetomany.OneToManyItem) {
		queueChannel := item.Value().(constructor.QueueChannel)
		enqueueError = queueChannel.Enqueue(ctx, enqueueItem)
	}

	if _, err := qccl.queueChannels.CreateOrFind(queueName, enqueueItem.KeyValues, bTreeOneToManyOnCreate, bTreeOneToManyOnFind); err != nil {
		switch err {
		//case btreeonetomany.ErrorOneIDDestroying:
		// This shouldn't happen as the 'Queue' should ensure these don't process once it starts destroying
		default:
			logger.Error("failed to create or find the queue channels", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return enqueueError
}

//	PARAMETERS:
//	- logger - general logger for this operation
//	- cancelContext - context that can be canceled to stop processing this function
//	- queueName - name of the queue to find items from
//	- dequeueQuery - query to match any channels for
//
//	RETURNS
//	- *v1willow.DequeueQueueItem - item dequeued that can be returned to the client who made the original request
//	- func() - success callback that must be called when the dequeueItem is sent back to the client
//	- func() - failure callback that must be called when the dequeueItem fails to send back to the original client
//	- *errors.ServerError - any unexpected errors during the dequeue process
//
// Dequeue an item from the queue. This is a blocking operation until an item is found that matches the query. This will also start a heartbeating
// operation for any succeffully dequeued items
func (qccl *queueChannelsClientLocal) DequeueQueueItem(ctx context.Context, queueName string, dequeueQuery datatypes.AssociatedKeyValuesQuery) (*v1willow.DequeueQueueItem, func(), func(), *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("DequeueQueueItem").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)

	// var of the return values
	var dequeueItem *v1willow.DequeueQueueItem
	var successCallback func()
	var failureCallback func()

	// setup our client so that any possible channels created after these calls are automatically added.
	// this is important to do before we traverse the queues so we don't miss any duplicate channels.
	// Duplicate channels added the the channelops will be dropped
	channelOperations, reader := channelops.NewRepeatableMergeRead[func(ctx context.Context) (*v1willow.DequeueQueueItem, func(), func())](false, ctx, qccl.shutdownCtx)
	qccl.addClientWaiting(queueName, dequeueQuery, channelOperations)
	defer qccl.removeClientWaiting(channelOperations)

	// in the background go through all the channels and add them to the channel operator. This will break if the reader finds a valid item to read
	go func() {
		bTreeOneToManyOnIterate := func(item btreeonetomany.OneToManyItem) bool {
			queueChannel := item.Value().(constructor.QueueChannel)

			// error is returned only when the channel operator is closed. so something must have been read
			dequeueChan := queueChannel.Dequeue()
			if err := channelOperations.MergeOrToOneIgnoreDuplicates(dequeueChan); err != nil {
				return false
			}

			return true
		}

		if err := qccl.queueChannels.Query(queueName, dequeueQuery, bTreeOneToManyOnIterate); err != nil {
			switch err {
			//case btreeonetomany.ErrorOneIDDestroying
			// This shouldn't happen as the 'Queue' should ensure these don't process once it starts destroying
			default:
				logger.Error("failed to create or find the queue channels", zap.Error(err))
				panic(err)
			}
		}
	}()

	logger.Debug("waiting for available item")
	defer logger.Debug("found available item")

	for repeatableReader := range reader {
		dequeueItem, successCallback, failureCallback = repeatableReader.Value(ctx)
		if dequeueItem != nil {
			// pulled something from the queue
			repeatableReader.Stop()

			return dequeueItem, successCallback, failureCallback, nil
		}

		// continue to dequeue a valid item
		repeatableReader.Continue()
	}

	// reader was closed. Must have been canceled by the client or server shutdown
	select {
	case <-ctx.Done():
		// client was closed, but still attempt to return something?
		return nil, nil, nil, &errors.ServerError{Message: "Client closed", StatusCode: http.StatusConflict}
	default:
		// server must be shutting down
		return nil, nil, nil, errors.ServerShutdown
	}
}

// ACK the item in a channel
func (qccl *queueChannelsClientLocal) ACK(ctx context.Context, queueName string, ack *v1willow.ACK) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("ACK").With(zap.String("queue_name", queueName), zap.Any("item_id", ack.ItemID))
	ctx = reporting.UpdateLogger(ctx, logger)

	ackErr := &errors.ServerError{Message: "Failed to find channel by key values ", StatusCode: http.StatusNotFound}

	tryDelete := false
	performAck := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
		queueChannel := oneToManyItem.Value().(constructor.QueueChannel)
		tryDelete, ackErr = queueChannel.ACK(ctx, ack)

		return false
	}

	keyValuesQuery := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{},
		},
	}

	for key, value := range ack.KeyValues {
		tmpValue := value
		keyValuesQuery.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &tmpValue, ValueComparison: datatypes.EqualsPtr()}
	}

	// ack the item in the queue channel
	// NOTE: this could be a delete, but with the locks i think that will be slower than I want so use 1 find and then 1 delete
	if err := qccl.queueChannels.Query(queueName, keyValuesQuery, performAck); err != nil {
		panic(err)
	}

	// if the item was acked and there are no more items, attempt to delete the channel
	if tryDelete {
		deleteChannel := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
			queueChannel := oneToManyItem.Value().(constructor.QueueChannel)
			return queueChannel.Delete()
		}

		if err := qccl.queueChannels.DeleteOneOfManyByKeyValues(queueName, ack.KeyValues, deleteChannel); err != nil {
			switch err {
			case btreeonetomany.ErrorManyIDDestroying:
				logger.Debug("Not deleting the channel after ack as queue is being deleted")
			default:
				logger.Fatal("Failed to delete queue after ack", zap.Error(err))
			}
		}
	}

	return ackErr
}

// Heartbeat an item that has been pulled from the queue
func (qccl *queueChannelsClientLocal) Heartbeat(ctx context.Context, queueName string, heartbeat *v1willow.Heartbeat) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("Heartbeat").With(zap.String("queue_name", queueName), zap.Any("item_id", heartbeat.ItemID))
	ctx = reporting.UpdateLogger(ctx, logger)
	heartbeatErr := &errors.ServerError{Message: "Failed to find channel for item by key values", StatusCode: http.StatusNotFound}

	performHeartbeat := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
		queueChannel := oneToManyItem.Value().(constructor.QueueChannel)
		heartbeatErr = queueChannel.Heartbeat(ctx, heartbeat)

		return false
	}

	// setup the query
	keyValuesQuery := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{},
		},
	}
	for key, value := range heartbeat.KeyValues {
		tmpValue := value
		keyValuesQuery.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &tmpValue, ValueComparison: datatypes.EqualsPtr()}
	}

	// ack the item in the queue channel
	// NOTE: this could be a delete, but with the locks i think that will be slower than I want so use 1 find and then 1 delete
	if err := qccl.queueChannels.Query(queueName, keyValuesQuery, performHeartbeat); err != nil {
		panic(err)
	}

	return heartbeatErr
}

// on dequeue, we add a client waiting to capture any newly created channels
func (qccl *queueChannelsClientLocal) addClientWaiting(queueName string, channelQuery datatypes.AssociatedKeyValuesQuery, channelOps *channelops.RepeatableMergeReadChannelOps[func(ctx context.Context) (*v1willow.DequeueQueueItem, func(), func())]) {
	qccl.clientsWaitingLock.Lock()
	defer qccl.clientsWaitingLock.Unlock()

	qccl.clientsWaiting = append(qccl.clientsWaiting, clientWaiting{queueName: queueName, query: channelQuery, channelOPS: channelOps})
}

// when a client finishes dequeue, it removes itself from the clients waiting to process an item
func (qccl *queueChannelsClientLocal) removeClientWaiting(channelOps *channelops.RepeatableMergeReadChannelOps[func(ctx context.Context) (*v1willow.DequeueQueueItem, func(), func())]) {
	qccl.clientsWaitingLock.Lock()
	defer qccl.clientsWaitingLock.Unlock()

	for index, clientWaiting := range qccl.clientsWaiting {
		if clientWaiting.channelOPS == channelOps {
			qccl.clientsWaiting[index] = qccl.clientsWaiting[len(qccl.clientsWaiting)-1]
			qccl.clientsWaiting = qccl.clientsWaiting[:len(qccl.clientsWaiting)-1]
			return
		}
	}
}

// when a new channel is created, check any clients currently waiting that might be interested in the channel
func (qccl *queueChannelsClientLocal) updateClientsWaiting(queueName string, channelTags datatypes.KeyValues, channel <-chan func(ctx context.Context) (*v1willow.DequeueQueueItem, func(), func())) {
	qccl.clientsWaitingLock.Lock()
	defer qccl.clientsWaitingLock.Unlock()

	for _, clientWaiting := range qccl.clientsWaiting {
		if clientWaiting.queueName == queueName {
			if clientWaiting.query.MatchTags(channelTags) {
				clientWaiting.channelOPS.MergeOrToOne(channel)
			}
		}
	}
}

// read operration for the channel
func (qccl *queueChannelsClientLocal) Channels(ctx context.Context, queueName string, query v1common.AssociatedQuery) v1willow.Channels {
	logger := reporting.GetLogger(ctx).Named("Channels").With(zap.String("queue_name", queueName))
	channels := v1willow.Channels{}

	queryChannels := func(oneToManyItem btreeonetomany.OneToManyItem) bool {
		channels = append(channels, &v1willow.Channel{
			KeyValues:     oneToManyItem.ManyKeyValues(),
			EnqueuedItems: -1,
			RunningItems:  -1,
		})

		return true
	}

	if err := qccl.queueChannels.Query(queueName, query.AssociatedKeyValues, queryChannels); err != nil {
		switch err {
		case btreeonetomany.ErrorManyIDDestroying:
			logger.Debug("Already destroying the queue's channels")
		default:
			logger.Fatal("Failed to query channels", zap.Error(err))
		}
	}

	return channels
}
