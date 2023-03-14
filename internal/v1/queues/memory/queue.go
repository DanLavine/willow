package memory

import (
	"context"
	"sync"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/v1/tags"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type Queue struct {
	lock     *sync.RWMutex
	done     chan struct{}
	doneOnce *sync.Once

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
	itemReadyCount      uint64
	itemProcessingCount uint64
}

func NewQueue(create *v1.Create) *Queue {
	tree, _ := datastructures.NewBTree(2)

	return &Queue{
		lock:     new(sync.RWMutex),
		doneOnce: new(sync.Once),

		name:       create.Name,
		maxSize:    create.QueueMaxSize,
		retryLimit: create.ItemRetryAttempts,

		tagGroups:  tree,
		tagReaders: tags.NewReaderTree(),

		itemReadyCount:      0,
		itemProcessingCount: 0,
	}
}

func (q *Queue) Init() *v1.Error {
	q.done = make(chan struct{})
	q.notifier = gonotify.New()
	q.taskManager = goasync.NewTaskManager(goasync.RelaxedConfig())

	return nil
}

func (q *Queue) OnFind() {
	// TODO something?
}

func (q *Queue) Execute(ctx context.Context) error {
	_ = q.taskManager.Run(ctx)
	return nil
}

func (q *Queue) Enqueue(enqueueItem *v1.EnqueueItem) *v1.Error {
	//readers := q.tags.CreateTagsGroup(enqueueItem.Tags)
	//q.tagGroups.FindOrCreate(datastructures.NewStringTreeKey(strings.Join(enqueueItem.Tags, ""), newTagGroup(readers)))

	q.itemReadyCount++
	return nil
}

func (q *Queue) GetItem() <-chan func() (*v1.DequeueItemResponse, *v1.Error) {
	//return q.itemChan
	return nil
}

func (q *Queue) Metrics() *v1.QueueMetrics {
	q.lock.Lock()
	defer q.lock.Unlock()

	return &v1.QueueMetrics{
		Name:                   q.name,
		Ready:                  q.itemReadyCount,
		Processing:             q.itemProcessingCount,
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
