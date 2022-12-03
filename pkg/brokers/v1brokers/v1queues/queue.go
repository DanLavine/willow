package v1queues

import (
	"sync"

	"github.com/DanLavine/gonotify"
)

// This doesn't need locks because the whole Queues wrapper has them?
type Queue struct {
	updateable bool

	notify *gonotify.Notify

	readChan chan []byte

	messageLock *sync.Mutex
	messages    [][]byte // slice of byte arrays
}

func NewQueue(updateable bool) *Queue {
	return &Queue{
		updateable: updateable,

		notify: gonotify.New(),

		messageLock: new(sync.Mutex),
		messages:    [][]byte{},
	}
}

// add a message to the queue.
func (q *Queue) addMessage(message []byte) {
	q.messageLock.Lock()
	defer q.messageLock.Unlock()

	// is this faster than append to do it ourselves than using append?
	newSize := len(q.messages) + 1
	messages := make([][]byte, newSize)
	copy(messages, q.messages)
	messages[newSize-1] = message

	q.messages = messages
	q.notify.Add()
}

func (q *Queue) messageChan() <-chan []byte {
	q.messageLock.Lock()
	defer q.messageLock.Unlock()

	return nil
}
