package memory

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/tasks"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/tags"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type Queue struct {
	doneOnce *sync.Once
	done     chan struct{}

	// queue information and limits
	name       string
	maxSize    uint64
	retryLimit uint64

	// items that are enqueued and ready to be processed
	// each item in the tree is of type *tagGroup
	tagGroups  datastructures.DisjointTree
	tagReaders tags.TagReaders // all readers for the various tag groups

	// manage all tag groups and their associated readers
	shutdown             context.CancelFunc
	taskManager          goasync.TaskManager
	taskManagerCompleted chan struct{}

	// queue metrics information
	itemReadyCount      *atomic.Uint64
	itemProcessingCount *atomic.Uint64
}

func NewQueue(create *v1.Create) *Queue {
	started := make(chan struct{})
	taskManagerCompleted := make(chan struct{})
	taskManager := goasync.NewTaskManager(goasync.RelaxedConfig())
	taskManager.AddTask("start", tasks.Running(started))

	// Want to wait for task manager to be running
	// TODO. Better way to do this? Don't like the go func here to know that things are running so Enqueue won't fail
	context, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(taskManagerCompleted)
		_ = taskManager.Run(context)
	}()

	<-started

	return &Queue{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		name:       create.Name,
		maxSize:    create.QueueMaxSize,
		retryLimit: create.ItemRetryAttempts,

		tagGroups:  datastructures.NewDisjointTree(),
		tagReaders: tags.NewTagReaderTree(),

		shutdown:             cancel,
		taskManager:          taskManager,
		taskManagerCompleted: taskManagerCompleted,

		itemReadyCount:      new(atomic.Uint64),
		itemProcessingCount: new(atomic.Uint64),
	}
}

// Execute is a managment function used by the queue manager to shutdown and cleanup any managed goroutines
func (q *Queue) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
	case <-q.done:
	}

	// stop running the taskmanager
	q.shutdown()

	// wait for the task manager to finish running
	<-q.taskManagerCompleted

	return nil
}

// Enqueue an item onto the message queue
func (q *Queue) Enqueue(enqueueItem *v1.EnqueueItem) *v1.Error {
	if q.itemProcessingCount.Load()+q.itemReadyCount.Load() >= q.maxSize {
		return errors.MaxEnqueuedItems
	}

	// generate all the keys
	var groupTags []datastructures.TreeKey
	for _, tag := range enqueueItem.Tags {
		groupTags = append(groupTags, datastructures.NewStringTreeKey(tag))
	}

	// create the new tags group if it does not currently exist
	// TODO: need something for OnFind to know that we shouldn't delete. Not sure what that is yet
	tagsGroup, err := q.tagGroups.FindOrCreate(groupTags, "", q.setupTagsGroup(enqueueItem.Tags))
	if err != nil {
		return errors.InternalServerError.With("", err.Error())
	}

	if err := tagsGroup.(*tagGroup).Enqueue(enqueueItem, q.itemReadyCount); err != nil {
		// TODO, need to try and remove tagGtoup/Readers?
		return err
	}

	return nil
}

// callbaack for setup Enqueue to use when setting up a new tags group
func (q *Queue) setupTagsGroup(tags []string) func() (any, error) {
	return func() (any, error) {
		allPossibleReaders := q.tagReaders.CreateGroup(tags)
		tagGroup := newTagGroup(allPossibleReaders)

		_ = q.taskManager.AddRunningTask("", tagGroup)
		return tagGroup, nil
	}
}

// Readers is used by any clients to obtain possible readers for tag groups
func (q *Queue) Readers(matchQuery *v1.MatchQuery) []<-chan tags.Tag {
	var channels []<-chan tags.Tag

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

func (q *Queue) Metrics() *v1.QueueMetrics {
	return &v1.QueueMetrics{
		Name:                   q.name,
		Ready:                  q.itemReadyCount.Load(),
		Processing:             q.itemProcessingCount.Load(),
		Max:                    q.maxSize,
		DeadLetterQueueMetrics: nil,
	}
}

func (q *Queue) Stop() {
	q.doneOnce.Do(func() {
		close(q.done)
	})
}
