package memory

import (
	"context"
	"sync"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"github.com/google/btree"
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
	tags  *datastructures.TagTree
	items []uint64
	//itemIDTracker *datastructures.IDTree
	// item Chan that is used to pull an item from the queue. Each reader must call the func to actually get the value
	itemChan      chan func() (*v1.DequeueItem, *v1.Error) // TODO, chang to be a struct. So we can check the tags[] on it up front, without haveing to process the message
	processedChan chan struct{}

	// queue metrics information
	itemReadyCount      uint64
	itemProcessingCount uint64
}

func NewQueue(create *v1.Create) *Queue {
	return &Queue{
		lock:     new(sync.RWMutex),
		done:     make(chan struct{}),
		doneOnce: new(sync.Once),

		name:       create.Name,
		maxSize:    create.QueueMaxSize,
		retryLimit: create.ItemRetryAttempts,

		notifier: gonotify.New(),

		tags:          btree.New(2),
		items:         []uint64{},
		itemIDTracker: datastructures.NewIDTree(),
		itemChan:      make(chan func() (*v1.DequeueItem, *v1.Error)),
		processedChan: make(chan struct{}),

		itemReadyCount:      0,
		itemProcessingCount: 0,
	}
}

// Handled by GoAsync to constantly read items from the queue and handle shutdown
func (q *Queue) Execute(ctx context.Context) error {
	defer q.Stop()

	for {
		select {
		case <-ctx.Done():
			// server is shutting down
			return nil
		case <-q.done:
			// received some sort of error and queue manager told this loop to shut down
			return nil
		case <-q.notifier.Ready():
			// item is ready for processing.
			select {
			case <-ctx.Done():
				return nil
			case <-q.done:
				return nil
			case q.itemChan <- q.processing:
				// sent an item for processing
			}

			// wait for the item to be processed
			select {
			case <-ctx.Done():
				return nil
			case <-q.done:
				return nil
			case _ = <-q.processedChan:
				// item was processed, can enqueue the next item
			}
		}
	}
}

func (q *Queue) processing() (*v1.DequeueItem, *v1.Error) {
	q.lock.Lock()
	defer func() {
		q.lock.Unlock()
		q.processedChan <- struct{}{}
	}()

	// this should always be safe since we know the notifier took place
	nextID := q.items[0]
	q.items = q.items[1:] // drop the first item, don't need it anymore

	//dequeueItem := &v1.DequeueItem{
	//	ID: nextID,
	//	Name: q.name,
	//	Tags: q.tags,
	//	Data: ,
	//}

	// update metrics
	q.itemReadyCount--
	q.itemProcessingCount++

	return nil, nil
}

func (q *Queue) Enqueue(enqueueItem *v1.EnqueueItem) *v1.Error {
	q.lock.Lock()
	defer q.lock.Unlock()

	tags := q.tags.FindOrCreate(enqueueItem.Tags)

	_ = q.notifier.Add()
	q.itemReadyCount++

	return nil
}

func (q *Queue) GetItem() <-chan func() (*v1.DequeueItem, *v1.Error) {
	return q.itemChan
}

func (q *Queue) Metrics() *v1.QueueMetrics {
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
		close(q.itemChan)

		q.notifier.ForceStop()
	})
}
