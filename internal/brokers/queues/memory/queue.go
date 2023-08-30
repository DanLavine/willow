package memory

import (
	"context"
	"net/http"
	"sync"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/pkg/models/query"
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
	name string

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

		tagGroups: btreeassociated.NewThreadSafe(),

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

	// callback to save off the tag group when found
	var foundTagGroup *tagGroup
	setTagGroup := func(item any) {
		tagNode := item.(*tagNode)
		if tagNode.tagGroup != nil {
			foundTagGroup = tagNode.tagGroup
		}
	}

	if err := q.tagGroups.Find(enqueueItemRequest.Tags, setTagGroup); err != nil {
		// always check find first to be fast
		logger.Error("fast finding queue failed", zap.Error(err))
		return errors.InternalServerError.With("", err.Error())
	} else if foundTagGroup == nil {
		// will need to either create, or update the tag group
		channels := []chan tags.Tag{q.globalChannel}
		tagPairs := enqueueItemRequest.BrokerInfo.Tags.GenerateTagPairs()

		for index, tagPair := range tagPairs {
			if index == len(tagPairs)-1 {
				// we are on the last index, so create the actual queue
				// NOTE: this never returns an error we would care about
				if err := q.tagGroups.CreateOrFind(tagPair, q.newTagNode(tagPair, &channels, setTagGroup), q.tagNodeUpdateTagGroup(tagPair, &channels, setTagGroup)); err != nil {
					logger.Error("CreateOrFind of tag group failed", zap.Error(err))
					return errors.InternalServerError.With("", err.Error())
				}
			} else {
				// we only need to create/add the channel to all possible channels for the tag group
				if err := q.tagGroups.CreateOrFind(tagPair, q.tagNodeNewGeneralChannel(&channels), q.tagNodeGetGeneralChannel(&channels)); err != nil {
					logger.Error("CreateOrFind of tag group subset failed", zap.Error(err))
					return errors.InternalServerError.With("", err.Error())
				}
			}
		}
	}

	return foundTagGroup.Enqueue(q.counter, enqueueItemRequest)
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
				if err := q.tagGroups.CreateOrFind(readerSelection.Tags, q.tagNodeNewStrictChannel(&channels), q.tagNodeGetStrictChannel(&channels)); err != nil {
					logger.Error("Failed to find exact readers", zap.Error(err))
					return nil, errors.InternalServerError.With("", err.Error())
				}
			case v1.ReaderMatches:
				// won't return an error
				if err := q.tagGroups.CreateOrFind(readerSelection.Tags, q.tagNodeNewGeneralChannelRead(&channels), q.tagNodeGetGeneralChannelRead(&channels)); err != nil {
					logger.Error("Failed to find  matches readers", zap.Error(err))
					return nil, errors.InternalServerError.With("", err.Error())
				}
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

		if tagNode.tagGroup != nil {
			deleteTagGroup, ackError = tagNode.tagGroup.ACK(q.counter, ackItem)
		} else {
			ackError = &v1.Error{Message: "tag group not found", StatusCode: http.StatusBadRequest}
		}

		called = true
	}

	if err := q.tagGroups.Find(ackItem.Tags, ack); err != nil {
		return errors.InternalServerError.With("", err.Error())
	} else if called == false {
		return &v1.Error{Message: "tag group not found", StatusCode: http.StatusBadRequest}
	} else if ackError != nil {
		return ackError
	}

	if deleteTagGroup {
		// need to also delete all tag pair combinations
		for _, tagPair := range ackItem.BrokerInfo.Tags.GenerateTagPairs() {
			q.tagGroups.Delete(tagPair, q.canDeleteTagNode)
		}
	}

	return nil
}

func (q *Queue) Metrics() *v1.QueueMetricsResponse {
	metrics := &v1.QueueMetricsResponse{
		Name:                   q.name,
		Max:                    q.counter.max,
		Total:                  q.counter.Total(),
		DeadLetterQueueMetrics: nil,
	}

	metricsFunc := func(value any) bool {
		tagNode := value.(*tagNode)
		tagNode.lock.RLock()
		defer tagNode.lock.RUnlock()

		if tagNode.tagGroup != nil {
			metrics.Tags = append(metrics.Tags, tagNode.tagGroup.Metrics())
		}

		return true
	}

	// find all items in the tree
	q.tagGroups.Query(query.Select{}, metricsFunc)

	return metrics
}

func (q *Queue) Stop() {
	q.doneOnce.Do(func() {
		close(q.done)
	})
}
