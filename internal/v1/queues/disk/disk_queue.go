package disk

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/DanLavine/gonotify"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/models"
	"github.com/DanLavine/willow/internal/v1/queues/disk/encoder"
	"github.com/DanLavine/willow/internal/v1/queues/disk/itemtracker"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type DiskQueue struct {
	drainOnce *sync.Once
	drainWG   *sync.WaitGroup
	drain     chan struct{}
	drainRead chan struct{}

	doneOnce *sync.Once
	done     chan struct{}

	readLoop chan struct{}

	lock *sync.Mutex

	// queue info
	queueTags []string

	// a tree that manages its own memory for each add and delete of an item
	itemTracker *itemtracker.IDTree

	// thread safe notifier that indicates if something is available. This also is minimal on memory usage
	notifier *gonotify.Notify

	// encoder to safely record any item to disk
	encoder *encoder.EncoderQueue

	// dead letter queue
	deadLetterQueue *DiskDeadLetterQueue

	// items that a client can process
	// these are the IDs in the itemTracker
	items          []uint64
	itemRetryCount uint64
	maxItems       uint64

	// queue stats
	readyCount      uint64
	processingCount uint64
}

// TODO record queue configuration as well for a reload:
// 1. Timeout length
// 2. Retry Count

func NewDiskQueue(configQueue config.ConfigQueue, createParams v1.Create, readers []chan *models.Location) (*DiskQueue, *v1.Error) {
	// reader validation
	if len(readers) == 0 {
		return nil, errors.NoReaders
	}
	for _, reader := range readers {
		if reader == nil {
			return nil, errors.NilReader
		}
	}

	// encoder for disk queue
	queueEncoder, err := encoder.NewEncoderQueue(configQueue.ConfigDisk.StorageDir, createParams.BrokerTags)
	if err != nil {
		return nil, err
	}

	// optional dead letter queue
	var deadLetterQueue *DiskDeadLetterQueue
	if createParams.DeadLetterQueueParams != nil {
		deadLetterQueue, err = NewDiskDeadLetterQueue(configQueue, createParams)
		if err != nil {
			queueEncoder.Close()
			return nil, err
		}
	}

	dq := &DiskQueue{
		drainOnce: new(sync.Once),
		drainWG:   new(sync.WaitGroup),
		drain:     make(chan struct{}),
		drainRead: make(chan struct{}),

		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		readLoop: make(chan struct{}),

		lock: new(sync.Mutex),

		queueTags: createParams.BrokerTags,

		itemTracker: itemtracker.NewIDTree(),

		notifier: gonotify.New(),

		encoder: queueEncoder,

		deadLetterQueue: deadLetterQueue,

		items:          []uint64{},
		itemRetryCount: createParams.QueueParams.RetryCount,
		maxItems:       createParams.QueueParams.MaxSize,

		readyCount:      uint64(0),
		processingCount: uint64(0),
	}

	go dq.nextItem(readers)

	return dq, nil
}

// TODO
// Possible changes. Because each item in the queue needs to be "processed", this could block untill that happens.
// Then we can have each queue with a nex() function that just returns the next item in the queue. Once that processes
// by a client we don't have a need to pass in the list of readable channels
func (dq *DiskQueue) nextItem(readers []chan *models.Location) {
	defer func() {
		close(dq.readLoop)
	}()

	notifierReady := dq.notifier.Ready()

	for {
		select {
		case _, ok := <-notifierReady:
			// closed
			if !ok {
				return
			}

			// add one to drain at the start. A Drain or Close operation will properly decrement this
			dq.drainWG.Add(1)

			dq.lock.Lock()
			id := dq.items[0]       // pull next item to be processed
			dq.items = dq.items[1:] // update items to pull next
			item := dq.itemTracker.Get(id)
			dq.lock.Unlock()

			if item == nil {
				panic(fmt.Errorf("queue should not be trying to get a nil item: %v", dq.queueTags))
			}

			location := item.(*models.Location)

			// setup select cases
			readersLen := len(readers)
			selectCases := make([]reflect.SelectCase, readersLen+2)
			for i, reader := range readers {
				selectCases[i] = reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(reader), Send: reflect.ValueOf(location)}
			}
			// setup the 2 cancel cases when shutting down
			selectCases[readersLen] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(dq.drainRead)}
			selectCases[readersLen+1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(dq.done)}

			_, _, ok = reflect.Select(selectCases)
			if ok {
				// recieved a close operaton
				dq.drainWG.Done()
				return
			}
		}
	}
}

func (dq *DiskQueue) processing(id uint64, startIndex, size int) (*v1.DequeueMessage, *v1.Error) {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	data, err := dq.encoder.Read(startIndex, size)
	if err != nil {
		return nil, err
	}

	err = dq.encoder.Processing(id)
	if err != nil {
		return nil, err
	}

	// update processing counts
	dq.readyCount--
	dq.processingCount++

	return &v1.DequeueMessage{
		ID:         id,
		BrokerTags: dq.queueTags,
		Data:       data,
	}, nil
}

func (dq *DiskQueue) Enqueue(item []byte) *v1.Error {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	// create the inital location to generate an ID
	location := models.NewLocation(dq.processing)
	location.ID = dq.itemTracker.Add(location)

	// write to file the data to encode
	var err *v1.Error
	location.StartIndex, location.Size, err = dq.encoder.Write(location.ID, item)
	if err != nil {
		// on a failure, the entire queue is going to close since something is corrupted. So remove the item from tracking anyways
		dq.itemTracker.Remove(location.ID)
		return err
	}

	dq.items = append(dq.items, location.ID)

	// ignore the error. if we are shutting down, then on a restart the load()
	// will just pick up the item
	_ = dq.notifier.Add()

	dq.readyCount++

	return nil
}

func (dq *DiskQueue) ACK(id uint64, passed bool) *v1.Error {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	item := dq.itemTracker.Get(id)
	if item == nil {
		return errors.ItemNotfound
	}

	location := item.(*models.Location)

	if !location.Processing() {
		return errors.ItemNotProcessing
	}

	if passed {
		if err := dq.encoder.Delete(location.ID); err != nil {
			return err
		}

		dq.processingCount--

		// don't care about the return id
		_ = dq.itemTracker.Remove(id)
	} else {
		if location.RetryCount >= dq.itemRetryCount {
			// delete and send to dead letter queue
			if err := dq.encoder.SentToDeadLetter(location.ID); err != nil {
				return err
			}

			data, err := dq.encoder.ReadRaw(location.StartIndex, location.Size)
			if err != nil {
				return err
			}

			dq.deadLetterQueue.Enqueue(data)
		} else {
			// reprocess
			if err := dq.encoder.Retry(location.ID); err != nil {
				return err
			}

			dq.items = append(dq.items, id)
			location.RetryCount++

			dq.readyCount++

			_ = dq.notifier.Add() // don't care about these errors
		}

		dq.processingCount--
	}

	return nil
}

// dead letter queue functions
// Retrieve an item from the DeadLetterQueue. Called from ADMIN API to inspect failures
func (dq *DiskQueue) DeadLetterQueueGet(index uint64) (*v1.DequeueMessage, *v1.Error) {
	if dq.deadLetterQueue == nil {
		return nil, errors.DeadLetterQueueNotConfigured
	}

	return dq.deadLetterQueue.Get(index, dq.queueTags)
}

// Retrieve total number of items enqueued in the dead letter queue
func (dq *DiskQueue) DeadLetterQueueCount() (uint64, *v1.Error) {
	if dq.deadLetterQueue == nil {
		return 0, errors.DeadLetterQueueNotConfigured
	}

	return dq.deadLetterQueue.Count(), nil
}

// Clear out all items in the dead letter queue
func (dq *DiskQueue) DeadLetterQueueClear() *v1.Error {
	if dq.deadLetterQueue == nil {
		return errors.DeadLetterQueueNotConfigured
	}

	return dq.deadLetterQueue.Clear()
}

// Metrics functions
func (dq *DiskQueue) Metrics() v1.QueueMetrics {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	var deadLetterQueueMetrics *v1.DeadLetterQueueMetrics
	if dq.deadLetterQueue != nil {
		deadLetterQueueMetrics = &v1.DeadLetterQueueMetrics{Count: dq.deadLetterQueue.Count()}
	}

	return v1.QueueMetrics{
		Tags:                   dq.queueTags,
		Ready:                  dq.readyCount,
		Processing:             dq.processingCount,
		DeadLetterQueueMetrics: deadLetterQueueMetrics, // should be set to nil if there is no dead letter queue configured
	}
}

// when shutting down, encoder might want to clean up some resorces
func (dq *DiskQueue) Drain() <-chan struct{} {
	dq.drainOnce.Do(func() {
		dq.notifier.Stop()
		close(dq.drainRead)

		go func() {
			dq.drainWG.Wait()
			close(dq.drain)

			dq.lock.Lock()
			defer dq.lock.Unlock()
			dq.encoder.Close()
		}()
	})

	return dq.drain
}

// called on an immediate shutdown, or some error occured as part of the read loop
func (dq *DiskQueue) Close() {
	dq.doneOnce.Do(func() {
		dq.notifier.ForceStop()
		close(dq.done)

		dq.lock.Lock()
		defer dq.lock.Unlock()
		dq.encoder.Close()
	})

	// this will close as part of the read loop
	// should wait to know that that was properly shutdown
	<-dq.readLoop
}
