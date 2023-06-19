package memory

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/internal/brokers/tags"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/errors"
)

type Queue struct {
	doneOnce *sync.Once
	done     chan struct{}

	// queue information and limits
	name datatypes.String

	// reader for all tag groups in this queue
	globalChannel chan tags.Tag

	// items that are enqueued and ready to be processed
	// Each element in this tree is of tyoe *tagGroup
	tagGroups btreeassociated.BTreeAssociated

	// manage all tag groups and their associated readers
	taskManager goasync.AsyncTaskManager

	// queue metrics information
	counter *Counter
}

func NewQueue(create *v1.Create) *Queue {
	return &Queue{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		name: create.Name,

		globalChannel: make(chan tags.Tag),

		tagGroups: btreeassociated.New(),

		taskManager: goasync.NewTaskManager(goasync.RelaxedConfig()),

		counter: NewCounter(create.QueueMaxSize),
	}
}

// Execute is a managment function used by the queue manager to shutdown and cleanup any managed goroutines
func (q *Queue) Execute(ctx context.Context) error {
	_ = q.taskManager.Run(ctx)
	return nil
}

// Enqueue an item onto the message queue
func (q *Queue) Enqueue(logger *zap.Logger, enqueueItemRequest *v1.EnqueueItemRequest) *v1.Error {
	logger = logger.Named("Enqueue")

	// try to be fast and find the tag group if it already exists
	item, err := q.tagGroups.Get(enqueueItemRequest.BrokerInfo.Tags, q.tagNodeLock)
	if err != nil {
		logger.Error("", zap.Error(err))
		return errors.InternalServerError.With("", err.Error())
	}

	// item might already exist
	if item != nil {
		node := item.(*tagNode)

		// item already exists, so just process that
		if node.tagGroup != nil {
			defer node.lock.Unlock()
			return node.tagGroup.Enqueue(q.counter, enqueueItemRequest)
		}

		// need to unlock here since create will lock again
		node.lock.Unlock()
	}

	// tag group not found, need to create all the readers and tag group
	channels := []chan tags.Tag{q.globalChannel}
	tagPairs := enqueueItemRequest.BrokerInfo.GenerateTagPairs()

	// create all readers and new tag group
	for index, tagPair := range tagPairs {
		if index == len(tagPairs)-1 {
			// we are on the last index, so create the actual queue
			// NOTE: this never returns an error we would care about
			item, _ = q.tagGroups.CreateOrFind(tagPair, q.newTagNode(tagPair, &channels), q.tagNodeUpdateTagGroup(tagPair, &channels))
		} else {
			// we only need to create/add the channel to all possible channels for the tag group
			_, _ = q.tagGroups.CreateOrFind(tagPair, q.tagNodeNewGeneralChannel(&channels), q.tagNodeGetGeneralChannel(&channels))
		}
	}

	node := item.(*tagNode)
	defer node.lock.Unlock()

	return node.tagGroup.Enqueue(q.counter, enqueueItemRequest)
}

// Readers is used by any clients to obtain possible readers for tag groups
func (q *Queue) Readers(logger *zap.Logger, readerSelect *v1.ReaderSelect) ([]<-chan tags.Tag, *v1.Error) {
	logger = logger.Named("Readers")
	var channels []<-chan tags.Tag

	if readerSelect == nil || readerSelect.Queries == nil {
		// this is the case were we want everything
		channels = append(channels, q.globalChannel)
	} else {
		// get all requested readers
		for _, readerSelection := range readerSelect.Queries {
			switch readerSelection.Type {
			case v1.ReaderExactly:
				// won't return an error
				_, _ = q.tagGroups.CreateOrFind(readerSelection.Tags, q.tagNodeNewStrictChannel(&channels), q.tagNodeGetStrictChannel(&channels))
			case v1.ReaderMatches:
				// won't return an error
				_, _ = q.tagGroups.CreateOrFind(readerSelection.Tags, q.tagNodeNewGeneralChannelRead(&channels), q.tagNodeGetGeneralChannelRead(&channels))
			}
		}
	}

	return channels, nil
}

func (q *Queue) ACK(logger *zap.Logger, ackItem *v1.ACK) *v1.Error {
	logger = logger.Named("ACK")

	called, deleteTagGroup := false, false
	var ackError *v1.Error
	ack := func(item any) {
		tagNode := item.(*tagNode)
		//tagNode.lock.RLock()
		//defer tagNode.lock.RUnlock()

		if tagNode.tagGroup != nil {
			deleteTagGroup, ackError = tagNode.tagGroup.ACK(q.counter, ackItem)
		} else {
			ackError = &v1.Error{Message: "tag group not found", StatusCode: http.StatusBadRequest}
		}

		called = true
	}

	_, err := q.tagGroups.Get(ackItem.Tags, ack)
	if err != nil {
		return errors.InternalServerError.With("", err.Error())
	} else if called == false {
		return &v1.Error{Message: "tag group not found", StatusCode: http.StatusBadRequest}
	}

	if deleteTagGroup {
		// need to also // delete all the combinations
		for _, tagPair := range ackItem.BrokerInfo.GenerateTagPairs() {
			fmt.Println("DSL deleteing tagPair:", tagPair)
			q.tagGroups.Delete(tagPair, q.canDeleteTagNode)
		}
	}

	return ackError
}

func (q *Queue) Metrics() *v1.QueueMetricsResponse {
	metrics := &v1.QueueMetricsResponse{
		Name:                   q.name,
		Max:                    q.counter.max,
		Total:                  q.counter.Total(),
		DeadLetterQueueMetrics: nil,
	}

	metricsFunc := func(_ datatypes.CompareType, value any) {
		tagNode := value.(*tagNode)
		tagNode.lock.RLock()
		defer tagNode.lock.RUnlock()

		if tagNode.tagGroup != nil {
			metrics.Tags = append(metrics.Tags, tagNode.tagGroup.Metrics())
		}
	}

	q.tagGroups.Iterate(metricsFunc)

	return metrics
}

func (q *Queue) Stop() {
	q.doneOnce.Do(func() {
		close(q.done)
	})
}
