package disk

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/DanLavine/gonotify"
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

	// queue stats
	readyCount      uint64
	processingCount uint64

	// encoder to safely record any item to disk
	diskEncoder *encoder.DiskEncoder

	// a tree that manages its own memory for each add and delete of an item
	itemTracker *itemtracker.IDTree

	// items that a client can process
	// these are the IDs in the itemTracker
	items []uint64

	// thread safe notifier that indicates if something is available. This also is minimal on memory usage
	notifier *gonotify.Notify
}

// TODO record queue configuration as well for a reload:
// 1. Timeout length
// 2. Retry Count

func NewDiskQueue(baseDir string, queueTags []string, readers []chan *models.Location) (*DiskQueue, error) {
	if len(readers) == 0 {
		return nil, fmt.Errorf("received empty readers")
	}

	for _, reader := range readers {
		if reader == nil {
			return nil, fmt.Errorf("received an empty reader")
		}
	}

	diskEncoder, err := encoder.NewDiskEncoder(baseDir, queueTags)
	if err != nil {
		return nil, err
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

		queueTags: queueTags,

		readyCount:      uint64(0),
		processingCount: uint64(0),

		diskEncoder: diskEncoder,

		itemTracker: itemtracker.NewIDTree(),

		items: []uint64{},

		notifier: gonotify.New(),
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
			id := dq.items[0]
			item := dq.itemTracker.Get(id)
			dq.items = dq.items[1:]
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

func (dq *DiskQueue) processing(id uint64, startIndex, size int) (*v1.DequeueMessage, error) {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	// update ready count
	dq.readyCount--

	data, err := dq.diskEncoder.Read(startIndex, size)
	if err != nil {
		return nil, err
	}

	// update processing count
	dq.processingCount++

	return &v1.DequeueMessage{
		ID:         id,
		BrokerTags: dq.queueTags,
		Data:       data,
	}, nil
}

func (dq *DiskQueue) Enqueue(item []byte) error {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	// create the inital location to generate an ID
	location := models.NewLocation(dq.processing)
	location.ID = dq.itemTracker.Add(location)

	// write to file the data to encode
	var err error
	location.StartIndex, location.Size, err = dq.diskEncoder.Write(location.ID, item)
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

func (dq *DiskQueue) ACK(id uint64) error {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	item := dq.itemTracker.Get(id)
	if item == nil {
		return fmt.Errorf("Failed to find item to ACK")
	}

	location := item.(*models.Location)

	if err := dq.diskEncoder.Remove(location.ID); err != nil {
		return err
	}

	// don't care about the return yet
	_ = dq.itemTracker.Remove(id)

	return nil
}

func (dq *DiskQueue) Metrics() v1.QueueMetrics {
	dq.lock.Lock()
	defer dq.lock.Unlock()

	return v1.QueueMetrics{
		Tags: dq.queueTags,

		Ready:      dq.readyCount,
		Processing: dq.processingCount,
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
			dq.diskEncoder.Close()
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
		dq.diskEncoder.Close()
	})

	// this will close as part of the read loop
	// should wait to know that that was properly shutdown
	<-dq.readLoop
}
