package memory

import (
	"context"
	"net/http"
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
	name    datatypes.String
	maxSize uint64

	// reader for all tag groups in this queue
	globalChannel chan tags.Tag

	// items that are enqueued and ready to be processed
	// Each element in this tree is of tyoe *tagNode
	tagGroups disjointtree.DisjointTree

	// manage all tag groups and their associated readers
	taskManager goasync.AsyncTaskManager

	// queue metrics information
	counter *Counter
}

func NewQueue(create *v1.Create) *Queue {
	return &Queue{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		name:    create.Name,
		maxSize: create.QueueMaxSize,

		globalChannel: make(chan tags.Tag),

		tagGroups: disjointtree.New(),

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
	// try and be fast to know if the queue already exist
	item, err := q.tagGroups.Find(enqueueItemRequest.BrokerInfo.SortedTags(), nil)
	if err != nil {
		return errors.InternalServerError.With("", err.Error())
	}

	// need to create the queue group
	if item != nil && item.(*tagNode).tagGroup != nil {
		// generate all possible tag combinations
		queueTags := enqueueItemRequest.BrokerInfo.GenerateTagPairs()

		// channel to be updated as part of insertion
		channels := []chan tags.Tag{}
		channels = append(channels, q.globalChannel)

		// NOTE: the queue tags are guaranteed to have the full queue tags as the last index
		for index, tagGroup := range queueTags {
			if index == len(queueTags)-1 {
				// we are on the last index, so create the actual queue
				item, err = q.tagGroups.CreateOrFind(queueTags, q.updateTagNode(queueTags, channels), q.newTagNode(queueTags, channels))
				if err != nil {
					return errors.InternalServerError.With("", err.Error())
				}
			} else {
				// we only need to create the channels
				if _, err := q.tagGroups.CreateOrFind(queueTags, q.findChannels(channels), q.newChannels(channels)); err != nil {
					return errors.InternalServerError.With("", err.Error())
				}
			}
		}
	}

	node := item.(*tagNode)
	return node.tagGroup.Enqueue(q.counter, enqueueItemRequest)
}

// TODO NEXT: I'M implementing all the search features for this
// Readers is used by any clients to obtain possible readers for tag groups
func (q *Queue) Readers(logger *zap.Logger, query *v1.Query) ([]<-chan tags.Tag, *v1.Error) {
	logger = logger.Named("Readers")
	var channels []<-chan tags.Tag

	if query.Matches.All {
		// using the global channel
		channels = append(channels, q.globalChannel)
	} else if query.Matches.StrictMatches != nil {
		// need to only find the strict channel
		item, err := q.tagGroups.Find(query.Matches.StrictMatches.Equals.ToStrings(), nil)
		if err != nil {
			logger.Error("failed to find tag group", zap.Error(err))
			return nil, &v1.Equals{Message: "Failed to find the queue", StatusCode: http.StatusInternalServerError}
		}

		if item == nil {
			return nil, nil
		}

		node := item.(*tagNode)
		if node.tagGroup != nil {
			channels = append(channels, node.tagGroup.strictChannel)
		}
	} else {
		// can find any number of channels
		// This needs a general recursive search for keys
	}

	return channels, nil
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
