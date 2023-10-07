package memory

import (
	"context"
	"net/http"
	"sync"

	"github.com/DanLavine/channelops"
	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
	"go.uber.org/zap"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/errors"
)

type clientWaiting struct {
	selection  query.AssociatedKeyValuesQuery
	channelOPS channelops.MergeReadChannelOps
}

type Queue struct {
	// process management
	done        chan struct{}
	doneOnce    *sync.Once
	lock        *sync.RWMutex
	taskManager goasync.AsyncTaskManager

	// queue information and limits
	name string
	// queue metrics information
	counter *Counter

	// items that are enqueued and ready to be processed
	// Each element in this tree is of tyoe *tagGroup
	tagGroups btreeassociated.BTreeAssociated

	clientsLock    *sync.RWMutex
	clientsWaiting []clientWaiting

	shutdownContext       context.Context
	shutdownContextCancel context.CancelFunc
}

func NewQueue(create *v1willow.Create) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		done:        make(chan struct{}),
		doneOnce:    new(sync.Once),
		lock:        new(sync.RWMutex),
		taskManager: goasync.NewTaskManager(goasync.RelaxedConfig()),

		name:    create.Name,
		counter: NewCounter(create.QueueMaxSize),

		tagGroups: btreeassociated.NewThreadSafe(),

		clientsLock:    new(sync.RWMutex),
		clientsWaiting: []clientWaiting{},

		shutdownContext:       ctx,
		shutdownContextCancel: cancel,
	}
}

// Execute is a managment function used by the queue manager to shutdown and cleanup any managed goroutines
func (q *Queue) Execute(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		q.shutdownContextCancel() // close any waiting clients
	}()

	_ = q.taskManager.Run(ctx) // waits for all tag groups to stop processing
	q.Stop()

	return nil
}

// Enqueue an item onto the message queue
func (q *Queue) Enqueue(logger *zap.Logger, enqueueItemRequest *v1willow.EnqueueItemRequest) *api.Error {
	logger = logger.Named("Enqueue")
	var returnErr *api.Error

	// if a tag group already exists, enqueue an item
	onFind := func(item any) {
		tagGroup := item.(*btreeassociated.AssociatedKeyValues).Value().(*tagGroup)
		returnErr = tagGroup.Enqueue(logger, q.counter, enqueueItemRequest)
	}

	// create a new tag group and manage the processin a new goasync task
	onCreate := func() any {
		// create the new tag group
		tagGroup := newTagGroup(enqueueItemRequest.Tags)

		// Enqueue the item
		returnErr = tagGroup.Enqueue(logger, q.counter, enqueueItemRequest)

		// start processing the tag group, if there is an error here, we are shutting down so who cares
		_ = q.taskManager.AddExecuteTask("", tagGroup)

		// update any know tag groups where the query matches this channel
		q.updateClientWating(enqueueItemRequest.Tags, tagGroup.dequeueChannel)

		// return the new tagGroup to save in the AssociatedBTree``
		return tagGroup
	}

	if _, err := q.tagGroups.CreateOrFind(enqueueItemRequest.Tags, onCreate, onFind); err != nil {
		logger.Error("failed to create or find the tag group", zap.Error(err))
		return errors.InternalServerError.With("", err.Error())
	}

	return returnErr
}

// Dequeue an item from the queue where the query selection matches a tag group
//
//	PARAMS:
//	- logger - logger to record any errors
//	- cancelContext - context from the http client to indicate if a client disconnects
//	- selection - query to use when searching for tag groups
func (q *Queue) Dequeue(logger *zap.Logger, cancelContext context.Context, selection query.AssociatedKeyValuesQuery) (*v1willow.DequeueItemResponse, func(), func(), *api.Error) {
	logger = logger.Named("DequeueItem")

	var dequeueResponse func(logger *zap.Logger) (*v1willow.DequeueItemResponse, func(), func(), *api.Error)
	channelOperations, reader := channelops.NewMergeRead(cancelContext, q.shutdownContext)

	// add the channel operations if we don't find any values, or a new tag group is added during iteration
	q.addClientWaiting(selection, channelOperations)
	// remove the channelOperations when this function returns
	defer q.removeClientWaiting(channelOperations)

	onFindPagination := func(item any) bool {
		tagGroup := item.(*btreeassociated.AssociatedKeyValues).Value().(*tagGroup)

		select {
		case response := <-tagGroup.dequeueChannel:
			if response != nil {
				dequeueResponse = response.(func(logger *zap.Logger) (*v1willow.DequeueItemResponse, func(), func(), *api.Error))
				return false
			}
		// Could add this optimization but its hard to test right here. So is there a better way to set evereything up?
		//case response := <-reader:
		default:
			channelOperations.MergeOrToOneIgnoreDuplicates(tagGroup.dequeueChannel)
		}

		return true
	}

	if err := q.tagGroups.Query(selection, onFindPagination); err != nil {
		logger.Error("failed to dequeue item", zap.Error(err))
		panic(err)
	}

	// found an item that was already waiting to be processed
	if dequeueResponse != nil {
		// found an item
		return dequeueResponse(logger)
	}

	// no items were ready for the client. Need to be notified when something is available
	readerVal := <-reader
	if readerVal != nil {
		// something was found
		return readerVal.(func(logger *zap.Logger) (*v1willow.DequeueItemResponse, func(), func(), *api.Error))(logger)
	}

	return nil, nil, nil, errors.QueueClosed
}

func (q *Queue) ACK(logger *zap.Logger, ackItem *v1willow.ACK) *api.Error {
	logger = logger.Named("ACK")
	var ackError *api.Error

	called := false
	ack := func(item any) bool {
		defer func() { called = true }()
		tagGroup := item.(*btreeassociated.AssociatedKeyValues).Value().(*tagGroup)

		if err := tagGroup.ACK(logger, q.counter, ackItem); err != nil {
			ackError = err
			return false
		}

		if tagGroup.itemReadyCount.Load()+tagGroup.itemProcessingCount.Load() == 0 {
			tagGroup.stop()
			return true
		}

		return false
	}

	if err := q.tagGroups.Delete(ackItem.Tags, ack); err != nil {
		return errors.InternalServerError.With("", err.Error())
	} else if !called {
		return &api.Error{Message: "tag group not found", StatusCode: http.StatusBadRequest}
	}

	return ackError
}

func (q *Queue) Metrics() *v1willow.QueueMetricsResponse {
	metrics := &v1willow.QueueMetricsResponse{
		Name:                   q.name,
		Max:                    q.counter.max,
		Total:                  q.counter.Total(),
		DeadLetterQueueMetrics: nil,
	}

	metricsFunc := func(item any) bool {
		tagGroup := item.(*btreeassociated.AssociatedKeyValues).Value().(*tagGroup)
		metrics.Tags = append(metrics.Tags, tagGroup.Metrics())

		return true
	}

	// find all items in the tree
	q.tagGroups.Query(query.AssociatedKeyValuesQuery{}, metricsFunc)

	return metrics
}

func (q *Queue) Stop() {
	q.doneOnce.Do(func() {
		close(q.done)
	})
}

func (q *Queue) addClientWaiting(selection query.AssociatedKeyValuesQuery, channelOps channelops.MergeReadChannelOps) {
	q.clientsLock.Lock()
	defer q.clientsLock.Unlock()

	q.clientsWaiting = append(q.clientsWaiting, clientWaiting{selection: selection, channelOPS: channelOps})
}

func (q *Queue) removeClientWaiting(channelOps channelops.MergeReadChannelOps) {
	q.clientsLock.Lock()
	defer q.clientsLock.Unlock()

	// find and remove the clients waiting
	for index, clientWaiting := range q.clientsWaiting {
		if clientWaiting.channelOPS == channelOps {
			q.clientsWaiting[index] = q.clientsWaiting[len(q.clientsWaiting)-1]
			q.clientsWaiting = q.clientsWaiting[:len(q.clientsWaiting)-1]
			return
		}
	}
}

func (q *Queue) updateClientWating(tags datatypes.StringMap, channel chan any) {
	q.clientsLock.RLock()
	defer q.clientsLock.RUnlock()

	// find and remove the clients waiting
	for _, clientWaiting := range q.clientsWaiting {
		if clientWaiting.selection.MatchTags(tags) {
			clientWaiting.channelOPS.MergeOrToOneIgnoreDuplicates(channel)
		}
	}
}
