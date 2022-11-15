package brokers

import (
	"fmt"
	"sync"

	"github.com/DanLavine/gomultiplex"
)

// This doesn't need locks because the whole Queues wrapper has them?
type queue struct {
	updateable bool

	messageLock *sync.Mutex
	messages    [][]byte // slice of byte arrays

	clientLocks *sync.Mutex
	clients     []*gomultiplex.Pipe
}

func newQueue(updateable bool) *queue {
	return &queue{
		updateable: updateable,

		messageLock: new(sync.Mutex),
		messages:    [][]byte{},

		clientLocks: new(sync.Mutex),
		clients:     []*gomultiplex.Pipe{},
	}
}

func (q *queue) addClient(pipe *gomultiplex.Pipe) error {
	//q.clientLocks.Lock()
	//defer q.clientLocks.Unlock()

	for _, currentPipe := range q.clients {
		if pipe == currentPipe {
			return fmt.Errorf("Pipe already exists")
		}
	}

	q.clients = append(q.clients)
	return nil
}

func (q *queue) addMessage(message []byte) {
	//q.messageLock.Lock()
	//defer q.messageLock.Unlock()

	q.messages = append(q.messages, message)
}
