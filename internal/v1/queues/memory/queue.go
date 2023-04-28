package memory

import (
	"context"
	"sync"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/tags"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"

	disjointtree "github.com/DanLavine/willow/internal/datastructures/disjoint_tree"
)

type Queue struct {
	doneOnce *sync.Once
	done     chan struct{}

	// queue information and limits
	name       datatypes.String
	maxSize    uint64
	retryLimit uint64

	// items that are enqueued and ready to be processed
	// each item in the tree is of type *tagGroup
	tagGroups  disjointtree.DisjointTree
	tagReaders tags.TagReaders // all readers for the various tag groups

	// manage all tag groups and their associated readers
	taskManager goasync.AsyncTaskManager

	// queue metrics information
	counter *Counter
}

func NewQueue(create *v1.Create) *Queue {
	return &Queue{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		name:       create.Name,
		maxSize:    create.QueueMaxSize,
		retryLimit: create.ItemRetryAttempts,

		tagGroups:  disjointtree.New(),
		tagReaders: tags.NewTagReaderTree(),

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
	// create the new tags group if it does not currently exist
	// TODO: need something for OnFind to know that we shouldn't delete. Not sure what that is yet
	tagsGroup, err := q.tagGroups.CreateOrFind(enqueueItemRequest.BrokerInfo.Tags, nil, q.setupTagsGroup(enqueueItemRequest.BrokerInfo.Tags))
	if err != nil {
		return errors.InternalServerError.With("", err.Error())
	}

	if err := tagsGroup.(*tagGroup).Enqueue(q.counter, enqueueItemRequest); err != nil {
		// TODO, need to try and remove tagGtoup/Readers?
		return err
	}

	return nil
}

// callbaack for setup Enqueue to use when setting up a new tags group
func (q *Queue) setupTagsGroup(tags datatypes.Strings) func() (any, error) {
	return func() (any, error) {
		allPossibleReaders := q.tagReaders.CreateGroup(tags)
		tagGroup := newTagGroup(tags, allPossibleReaders)

		_ = q.taskManager.AddExecuteTask("", tagGroup)
		return tagGroup, nil
	}
}

// Readers is used by any clients to obtain possible readers for tag groups
func (q *Queue) Readers(matchQuery *v1.MatchQuery) []<-chan tags.Tag {
	var channels []<-chan tags.Tag

	if matchQuery == nil {
		return channels
	}

	switch matchQuery.MatchTagsRestrictions {
	case v1.STRICT:
		// return the strict reader for the queue's tagGroup
		channels = append(channels, q.tagReaders.GetStrictReader(matchQuery.Tags))
	case v1.SUBSET:
		// return the specific subset reader
		channels = append(channels, q.tagReaders.GetSubsetReader(matchQuery.Tags))
	case v1.ANY:
		// return all readers that match
		channels = q.tagReaders.GetAnyReaders(matchQuery.Tags)
	case v1.ALL:
		// return the global reader
		channels = append(channels, q.tagReaders.GetGlobalReader())
	default:
		// nothing to do here. Caller should decide what to do with nil?
	}

	return channels
}

func (q *Queue) Metrics() *v1.QueueMetricsResponse {
	metrics := &v1.QueueMetricsResponse{
		Name:                   q.name,
		Max:                    q.maxSize,
		Total:                  q.counter.Total(),
		DeadLetterQueueMetrics: nil,
	}

	metricsFunc := func(value any) {
		metrics.Tags = append(metrics.Tags, value.(*tagGroup).Metrics())
	}

	q.tagGroups.Iterate(metricsFunc)

	return metrics
}

func (q *Queue) Stop() {
	q.doneOnce.Do(func() {
		close(q.done)
	})
}
