package v1queues

import (
	"sync"
	"sync/atomic"

	"github.com/DanLavine/gonotify"
)

// Queues are shared between any number of producer and consumer clients.
//
// TODO. What does a valid "close" or "cleanup" for a queue look like?
type Queue struct {
	// updateable indicates if messages should overwrite a message thats currently waiting to be
	// processed from a consumer. RE-QUEUED failures will not be overwritten.
	// TODO, feature not yet implemented
	updateable bool

	// helper package to know if there is a new message to process.
	notify *gonotify.Notify

	readChan chan []byte

	messageLock *sync.Mutex
	messages    [][]byte // slice of byte arrays

	// keep track of number of producers, consumers and messages
	messageCount  *atomic.Int32
	producerCount *atomic.Int32
	consumerCount *atomic.Int32
}

func NewQueue(updateable bool) *Queue {
	queue := &Queue{
		updateable: updateable,

		notify: gonotify.New(),

		readChan: make(chan []byte),

		messageLock: new(sync.Mutex),
		messages:    [][]byte{},

		messageCount:  new(atomic.Int32),
		producerCount: new(atomic.Int32),
		consumerCount: new(atomic.Int32),
	}

	go func() {
		for {
			select {
			// TODO shutdown/cleanup case
			//case <-queue.done:
			case ready := <-queue.notify.Ready():
				if ready != nil {
					// always grab the first message on the queue and pop the first item
					queue.messageLock.Lock()
					message := queue.messages[0]
					queue.messages = queue.messages[1:]
					queue.messageLock.Unlock()

					queue.messageCount.Add(-1)

					// wait for a client to process the message
					queue.readChan <- message
				}
			}
		}
	}()

	return queue
}

// add a message to the queue.
func (q *Queue) AddMessage(message []byte) {
	q.messageLock.Lock()
	defer q.messageLock.Unlock()

	// is this faster than append to do it ourselves than using append?
	newSize := len(q.messages) + 1
	messages := make([][]byte, newSize)
	copy(messages, q.messages)
	messages[newSize-1] = message

	q.messageCount.Add(1)

	q.messages = messages
	q.notify.Add()
}

func (q *Queue) messageChan() <-chan []byte {
	return q.readChan
}

// metic counters
func (q *Queue) AddProducer() {
	q.producerCount.Add(1)
}
func (q *Queue) DecrementProducer() {
	q.producerCount.Add(-1)
}
func (q *Queue) GetProducerCount() int32 {
	return q.producerCount.Load()
}

func (q *Queue) AddConsumer() {
	q.consumerCount.Add(1)
}
func (q *Queue) DecrementConsumer() {
	q.consumerCount.Add(-1)
}
func (q *Queue) GetConsumerCount() int32 {
	return q.consumerCount.Load()
}

func (q *Queue) GetMessageCount() int32 {
	return q.messageCount.Load()
}

// nothing to do on an ack?
func (q *Queue) Ack() {
	//TODO
}

// how do we ensure that a message is process properly?
// if there are multiple failure. Then their order should be preserved
func (q *Queue) Fail() {
	//TODO
}
