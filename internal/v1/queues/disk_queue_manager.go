package queues

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/DanLavine/willow/internal/config"
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

	// configuration for queues
	configQueue config.ConfigQueue

	// all the queues
	queues []*taggedQueue

	// fitleredTags
	filteredTags *filteredTags
}

// TODO: handle duplicate tags
func NewDiskQueueManager(configQueue config.ConfigQueue) *DiskQueueManager {
	return &DiskQueueManager{
		lock:         new(sync.RWMutex),
		configQueue:  configQueue,
		queues:       []*taggedQueue{},
		filteredTags: NewFilteredTags(),
	}
}

// Create a new queue for a specific tag. If the queue and tag already exists
// this returns nil and is a NO-OP
func (dqm *DiskQueueManager) Create(createParams v1.Create) *v1.Error {
	dqm.lock.Lock()
	defer dqm.lock.Unlock()

	// if queue already exist, just return
	queue, _ := dqm.getDiskQueue(createParams.BrokerTags)
	if queue != nil {
		return nil
	}

	readers := []chan *models.Location{make(chan *models.Location)} // strict reader
	readers = append(readers, dqm.filteredTags.createFilteredTags(createParams.BrokerTags)...)

	newDiskQueue, err := disk.NewDiskQueue(dqm.configQueue, createParams, readers)
	if err != nil {
		// TODO: decrement reader count
		return err
	}
	dqm.queues = append(dqm.queues, &taggedQueue{tags: createParams.BrokerTags, reader: readers[0], queue: newDiskQueue})

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
func (dqm *DiskQueueManager) Item(ctx context.Context, matchQuery v1.MatchQuery) (*v1.DequeueMessage, *v1.Error) {
	switch matchQuery.MatchRestriction {
	case v1.STRICT: // only sstrict returns if there is no queue available
		taggedQueue, err := dqm.getDiskQueue(matchQuery.BrokerTags)
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
		dqm.lock.Lock()
		reader := dqm.filteredTags.findOrCreateSubset(matchQuery.BrokerTags)
		dqm.lock.Unlock()

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
		dqm.lock.Lock()
		readers := dqm.filteredTags.findAny(matchQuery.BrokerTags)
		dqm.lock.Unlock()

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
func (dqm *DiskQueueManager) ACK(id uint64, queueTags []string, passed bool) *v1.Error {
	sort.Strings(queueTags)

	taggedQueue, err := dqm.getDiskQueue(queueTags)
	if err != nil {
		return err
	}

	return taggedQueue.queue.ACK(id, passed)
}

func (dqm *DiskQueueManager) Drain() <-chan struct{} {
	return nil
}

func (dqm *DiskQueueManager) Metrics(matchRestriction *v1.MatchQuery) *v1.Metrics {
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
