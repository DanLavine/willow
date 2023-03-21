package memory

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/v1/tags"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type Queue struct {
	done      chan struct{}
	setupOnce *sync.Once
	doneOnce  *sync.Once

	// queue information and limits
	name       string
	maxSize    uint64
	retryLimit uint64

	// thread safe notifier that indicates if something is available. This also is minimal on memory usage
	notifier *gonotify.Notify

	// items that are enqueued and ready to be processed
	tagGroups  datastructures.BTree
	tagReaders tags.Readers

	taskManager goasync.TaskManager
	//itemIDTracker *datastructures.IDTree
	// item Chan that is used to pull an item from the queue. Each reader must call the func to actually get the value
	//itemChan      chan func() (*v1.DequeueItemResponse, *v1.Error) // TODO, chang to be a struct. So we can check the tags[] on it up front, without haveing to process the message
	//processedChan chan struct{}

	// queue metrics information
	itemReadyCount      *atomic.Uint64
	itemProcessingCount *atomic.Uint64
}

func NewQueue(create *v1.Create) *Queue {
	tree, _ := datastructures.NewBTree(2)

	return &Queue{
		setupOnce: new(sync.Once),
		doneOnce:  new(sync.Once),

		name:       create.Name,
		maxSize:    create.QueueMaxSize,
		retryLimit: create.ItemRetryAttempts,

		tagGroups:  tree,
		tagReaders: tags.NewReaderTree(),

		itemReadyCount:      new(atomic.Uint64),
		itemProcessingCount: new(atomic.Uint64),
	}
}

func (q *Queue) OnFind() {
	q.setupOnce.Do(func() {
		q.done = make(chan struct{})
		q.notifier = gonotify.New()
		q.taskManager = goasync.NewTaskManager(goasync.RelaxedConfig())
	})
}

func (q *Queue) Execute(ctx context.Context) error {
	_ = q.taskManager.Run(ctx)
	return nil
}

func (q *Queue) Enqueue(enqueueItem *v1.EnqueueItem) *v1.Error {
	// TODO. Need something here to lookup my "tags" in a safe way. Combining them isn't safe since
	// [a, b] == [ab]
	//readers := q.tags.CreateTagsGroup(enqueueItem.Tags)
	//q.tagGroups.FindOrCreate(datastructures.NewStringTreeKey(strings.Join(enqueueItem.Tags, ""), newTagGroup(readers)))

	q.itemProcessingCount.Add(1)
	// q.itemProcessingCount.Add()
	return nil
}

func (q *Queue) GetItem() <-chan func() (*v1.DequeueItemResponse, *v1.Error) {
	//return q.itemChan
	return nil
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
		//close(q.itemChan)

		q.notifier.ForceStop()
	})
}
