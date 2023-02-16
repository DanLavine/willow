package disk

import (
	"container/list"
	"context"
	"sync"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/datastructures"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type Queue struct {
	doneOnce *sync.Once
	done     chan struct{}

	lock *sync.Mutex

	// default queue info
	name       string
	maxItems   uint64
	retryCount uint64

	// queue stats
	readyCount      uint64
	processingCount uint64

	// thread safe notifier that indicates if something is available. This also is minimal on memory usage
	notifier *gonotify.Notify
	// list of items in the queue that are ready to be processed
	items          *list.List
	itemIDTraceker *datastructures.IDTree
	// item Chan that is used to pull an item from the queue. Each reader must call the func to actually get the value
	itemChan chan func() (*v1.DequeueItem, *v1.Error)
}

func NewQueue(baseDir string, create *v1.Create) (*Queue, *v1.Error) {
	if err := recordQueueInfo(baseDir, create); err != nil {
		return nil, err
	}

	return createQueue(create), nil
}

func LoadQueue(baseDir, name string) (*Queue, *v1.Error) {
	create, err := loadQueueInfo(baseDir, name)
	if err != nil {
		return nil, err
	}

	return createQueue(create), nil
}

func createQueue(create *v1.Create) *Queue {
	return &Queue{
		doneOnce:        new(sync.Once),
		done:            make(chan struct{}),
		lock:            new(sync.Mutex),
		name:            create.Name,
		maxItems:        create.QueueMaxSize,
		retryCount:      create.RetryCount,
		readyCount:      0,
		processingCount: 0,
		notifier:        gonotify.New(),
		items:           list.New(),
		itemChan:        make(chan func() (*v1.DequeueItem, *v1.Error)),
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
				// server is shutting down
				return nil
			case <-q.done:
				// received some sort of error and queue manager told this loop to shut down
				return nil
			case q.itemChan <- q.processing:
				// sent an item for processing
			}
		}
	}
}

func (q *Queue) processing() (*v1.DequeueItem, *v1.Error) {
	return nil, nil
}

func (q *Queue) Enqueue(enqueueItem *v1.EnqueItem) *v1.Error {
	q.lock.Lock
	defer q.lock.Unlock()

	q.readyCount++

	return nil
}

func (q *Queue) GetItem() <-chan func() (*v1.DequeueItem, *v1.Error) {
	return q.itemChan
}

func (q *Queue) Metrics() *v1.QueueMetrics {
	return &v1.QueueMetrics{
		Name:                   q.name,
		Ready:                  q.readyCount,
		Processing:             q.processingCount,
		Max:                    q.maxItems,
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
