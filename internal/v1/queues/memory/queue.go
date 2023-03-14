package memory

import (
	"context"
	"sync"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type Queue struct {
	lock     *sync.RWMutex
	done     chan struct{}
	doneOnce *sync.Once

	// queue information and limits
	name       string
	retryLimit uint64

	// thread safe notifier that indicates if something is available. This also is minimal on memory usage
	notifier *gonotify.Notify

	// items that are enqueued and ready to be processed
	tags  datastructures.BTree
	items []uint64
	//itemIDTracker *datastructures.IDTree
	// item Chan that is used to pull an item from the queue. Each reader must call the func to actually get the value
	//itemChan      chan func() (*v1.DequeueItemResponse, *v1.Error) // TODO, chang to be a struct. So we can check the tags[] on it up front, without haveing to process the message
	//processedChan chan struct{}

	// queue metrics information
	itemReadyCount      uint64
	itemProcessingCount uint64
}

func NewQueue(create *v1.Create) *Queue {
	bTree, err := datastructures.NewBTree(2)
	if err != nil {
		panic(err)
	}

	return &Queue{
		lock:     new(sync.RWMutex),
		doneOnce: new(sync.Once),

		name:       create.Name,
		retryLimit: create.ItemRetryAttempts,

		tags:  bTree,
		items: make([]uint64, 0, create.QueueMaxSize),
		//itemIDTracker: datastructures.NewIDTree(),

		//itemReadyCount:      0,
		//itemProcessingCount: 0,
	}
}

func (q *Queue) Init() *v1.Error {
	q.done = make(chan struct{})
	q.notifier = gonotify.New()

	return nil
}

// Handled by GoAsync to constantly read items from the queue and handle shutdown
func (q *Queue) Execute(ctx context.Context) error {
	defer q.Stop()

	//for {
	//	select {
	//	case <-ctx.Done():
	//		// server is shutting down
	//		return nil
	//	case <-q.done:
	//		// received some sort of error and queue manager told this loop to shut down
	//		return nil
	//	case <-q.notifier.Ready():
	//		// item is ready for processing.
	//		select {
	//		case <-ctx.Done():
	//			return nil
	//		case <-q.done:
	//			return nil
	//		case q.itemChan <- q.processing:
	//			// sent an item for processing
	//		}

	//		// wait for the item to be processed
	//		select {
	//		case <-ctx.Done():
	//			return nil
	//		case <-q.done:
	//			return nil
	//		case _ = <-q.processedChan:
	//			// item was processed, can enqueue the next item
	//		}
	//	}
	//}

	return nil
}

func (q *Queue) processing() (*v1.DequeueItemResponse, *v1.Error) {
	//q.lock.Lock()
	//defer func() {
	//	q.lock.Unlock()
	//	q.processedChan <- struct{}{}
	//}()

	//// this should always be safe since we know the notifier took place
	//nextID := q.items[0]
	//q.items = q.items[1:] // drop the first item, don't need it anymore

	////dequeueItem := &v1.DequeueItem{
	////	ID: nextID,
	////	Name: q.name,
	////	Tags: q.tags,
	////	Data: ,
	////}

	//// update metrics
	//q.itemReadyCount--
	//q.itemProcessingCount++

	return nil, nil
}

func (q *Queue) Enqueue(enqueueItem *v1.EnqueueItem) *v1.Error {
	q.lock.Lock()
	defer q.lock.Unlock()

	//	:= q.tags.FindOrCreate(enqueueItem.Tags)
	//
	//	_ = q.notifier.Add()
	//	q.itemReadyCount++

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
		Max:                    uint64(cap(q.items)),
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
