package v1queues

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/DanLavine/gonotify"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
)

// Queues are shared between any number of producer and consumer clients.
//
// TODO. What does a valid "close" or "cleanup" for a queue look like?
type Queue struct {
	done     chan struct{}
	doneOnce *sync.Once

	// helper package to know if there is a new message to process.
	notify *gonotify.Notify

	// ring queue
	messageLock *sync.Mutex
	messages    []*v1.Message
	messageChan chan *v1.Message

	// keep track of number of producers, consumers and messages
	messageReadyCount      *atomic.Int32
	messageProcessingCount *atomic.Int32
}

func NewQueue() *Queue {
	queue := &Queue{
		done:     make(chan struct{}),
		doneOnce: new(sync.Once),

		notify: gonotify.New(),

		messageLock: new(sync.Mutex),
		messages:    []*v1.Message{},
		messageChan: make(chan *v1.Message),

		messageReadyCount:      new(atomic.Int32),
		messageProcessingCount: new(atomic.Int32),
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

					queue.messageReadyCount.Add(-1)

					// wait for a client to process the message
					select {
					case queue.messageChan <- message:
					case <-queue.done:
					}
				} else {
					close(queue.messageChan)
					return
				}
			}
		}
	}()

	return queue
}

func (q *Queue) Cancel() {
	q.doneOnce.Do(func() {
		q.notify.Stop()
		close(q.done)
	})
}

// add a message to the queue.
// TODO updateable not yet implemented
func (q *Queue) AddMessage(message *v1.Message) {
	q.messageLock.Lock()
	defer q.messageLock.Unlock()

	insertIndex := len(q.messages)
	message.ID = uint64(insertIndex)

	// is this faster than append to do it ourselves than using append?
	newSize := len(q.messages) + 1
	messages := make([]*v1.Message, newSize)
	copy(messages, q.messages)
	messages[insertIndex] = message

	q.messages = messages

	q.messageReadyCount.Add(1)
	q.notify.Add()
}

func (q *Queue) RetrieveMessage() *v1.Message {
	select {
	case message, ok := <-q.messageChan:
		if ok {
			q.messageProcessingCount.Add(1)

			return message
		}

		return nil
	}
}

func (q *Queue) ACK(id uint64, passed bool) error {
	q.messageProcessingCount.Add(-1)

	switch passed {
	case true:
		q.messageLock.Lock()
		defer q.messageLock.Unlock()

		if uint64(len(q.messages)) >= id {
			// passed, so remove the item from the queue
			q.messages = append(q.messages[:id], q.messages[id+1:]...)
			q.messageReadyCount.Add(-1)
			return nil
		} else {
			// invalid index received
			return fmt.Errorf("message id is large then total number of messages")
		}
	case false:
	}

	return nil
}

// metic counters
func (q *Queue) GetMessageReadyCount() int32 {
	return q.messageReadyCount.Load()
}
func (q *Queue) GetMessageProcessingCount() int32 {
	return q.messageProcessingCount.Load()
}
