package queues

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/models"
	"github.com/DanLavine/willow/internal/v1/queues/disk"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type taggedQueue struct {
	tags []string

	reader chan *models.Location

	queue *disk.DiskQueue
}

type DiskQueueManager struct {
	lock *sync.RWMutex

	// dir where all queues will be constructed at
	baseDir string

	// all the queues
	queues []*taggedQueue

	// fitleredTags
	filteredTags *filteredTags
}

// TODO: handle duplicate tags
func NewDiskQueueManager(baseDir string) *DiskQueueManager {
	return &DiskQueueManager{
		lock:         new(sync.RWMutex),
		baseDir:      baseDir,
		queues:       []*taggedQueue{},
		filteredTags: NewFilteredTags(),
	}
}

// Create a new queue for a specific tag. If the queue and tag already exists
// this returns nil and is a NO-OP
func (dqm *DiskQueueManager) Create(queueTags []string) *v1.Error {
	sort.Strings(queueTags)

	// if queue already exist, just return
	queue, _ := dqm.getDiskQueue(queueTags)
	if queue != nil {
		return nil
	}

	// setup a tag reader, or increment the count for the readers already created
	dqm.lock.Lock()
	defer dqm.lock.Unlock()

	readers := []chan *models.Location{make(chan *models.Location)} // strict reader
	readers = append(readers, dqm.filteredTags.createFilteredTags(queueTags)...)

	newDiskQueue, err := disk.NewDiskQueue(dqm.baseDir, queueTags, readers)
	if err != nil {
		// TODO: decrement reader count
		return err
	}
	dqm.queues = append(dqm.queues, &taggedQueue{tags: queueTags, reader: readers[0], queue: newDiskQueue})

	return nil
}

// Enqueue a new massage to a queue
// TODO: actually use updateable
func (dqm *DiskQueueManager) Enqueue(data []byte, updateable bool, queueTags []string) *v1.Error {
	taggedQueue, err := dqm.getDiskQueue(queueTags)
	if err != nil {
		return err
	}

	return taggedQueue.queue.Enqueue(data)
}

// Blocking operation to retrieve a message for any given queue and the tags
// Empty tags means to read from all avilable tags
func (dqm *DiskQueueManager) Message(ctx context.Context, matchRestriction v1.MatchRestriction, queueTags []string) (*v1.DequeueMessage, *v1.Error) {
	switch matchRestriction {
	case v1.STRICT: // only sstrict returns if there is no queue available
		taggedQueue, err := dqm.getDiskQueue(queueTags)
		if err != nil {
			return nil, err
		}

		select {
		case <-ctx.Done(): // client has dissconnected
			return nil, nil
		case location, ok := <-taggedQueue.reader:
			if ok {
				return location.Process()
			}

			// happens on close or drain
			return nil, errors.ServerShutdown
		}
	case v1.SUBSET:
		dqm.lock.RLock()
		reader := dqm.filteredTags.findOrCreateSubset(queueTags)
		dqm.lock.RUnlock()

		select {
		case <-ctx.Done(): // client has dissconnected
			return nil, nil
		case location, ok := <-reader:
			if ok {
				return location.Process()
			}

			// happens on close or drain
			return nil, errors.ServerShutdown
		}
	case v1.ANY:
		dqm.lock.RLock()
		readers := dqm.filteredTags.findAny(queueTags)
		dqm.lock.RUnlock()

		if len(readers) == 0 {
			return nil, errors.QueueNotFound.With("any queues to already be created with provided tags", "")
		}

		selectCases := []reflect.SelectCase{{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}}
		for _, reader := range readers {
			selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(reader)})
		}

		index, val, _ := reflect.Select(selectCases)
		if index == 0 {
			return nil, errors.ServerShutdown
		}

		return val.Interface().(*models.Location).Process()
	case v1.ALL:
		dqm.lock.RLock()
		globalReader := dqm.filteredTags.allTags[0].reader
		dqm.lock.RUnlock()

		select {
		case <-ctx.Done(): // client has dissconnected
			return nil, nil
		case location, ok := <-globalReader:
			if ok {
				return location.Process()
			}

			// happens on close or drain
			return nil, errors.ServerShutdown
		}
	}

	return nil, errors.MessageTypeNotSupported
}

// ACK a message for a particular id and tag
// TODO actually use the passed bool
func (dqm *DiskQueueManager) ACK(id uint64, passed bool, queueTags []string) *v1.Error {
	sort.Strings(queueTags)

	taggedQueue, err := dqm.getDiskQueue(queueTags)
	if err != nil {
		return err
	}

	return taggedQueue.queue.ACK(id)
}

func (dqm *DiskQueueManager) Drain() <-chan struct{} {
	return nil
}

func (dqm *DiskQueueManager) Metrics() *v1.Metrics {
	dqm.lock.Lock()
	defer dqm.lock.Unlock()

	metrics := &v1.Metrics{
		Queues: []v1.QueueMetrics{},
	}

	for _, taggedQueues := range dqm.queues {
		metrics.Queues = append(metrics.Queues, taggedQueues.queue.Metrics())
	}

	return metrics
}

func (dqm *DiskQueueManager) getDiskQueue(queueTags []string) (*taggedQueue, *v1.Error) {
	dqm.lock.RLock()
	defer dqm.lock.RUnlock()

	for _, taggedQueue := range dqm.queues {
		if tagsEqual(taggedQueue.tags, queueTags) {
			return taggedQueue, nil
		}
	}

	tagsString := "["
	for index, tag := range queueTags {
		if index != 0 {
			tagsString = fmt.Sprintf("%s, ", tagsString)
		}

		tagsString = fmt.Sprintf("%s'%s'", tagsString, tag)
	}
	tagsString = fmt.Sprintf("%s]", tagsString)

	return nil, errors.QueueNotFound.With(fmt.Sprintf("Tags %s to have a queue associated", tagsString), "")
}
