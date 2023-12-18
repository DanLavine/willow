package lockmanager

import (
	"context"
	"sync"
	"time"
)

type generalLock struct {
	lockChan chan struct{}
	hertbeat chan struct{}

	timeout time.Duration
	cleanup func() bool

	counterLock *sync.RWMutex
	counter     int
}

func newGeneralLock(timeout time.Duration, cleanup func() bool) *generalLock {
	return &generalLock{
		lockChan:    make(chan struct{}),
		hertbeat:    make(chan struct{}),
		timeout:     timeout,
		cleanup:     cleanup,
		counterLock: new(sync.RWMutex),
		counter:     1,
	}
}

func (generalLock *generalLock) Execute(ctx context.Context) error {
	timer := time.NewTicker(generalLock.timeout)

	for {
		select {
		case <-generalLock.hertbeat:
			// in this case we recieved a heartbeat, so reset the timer
			timer.Reset(generalLock.timeout)
		case <-ctx.Done():
			// in this case, we received a shutdown signal from the server, so just cancel this threadand don't clean
			// anything up since we eventually want locks to persist through a server restart
			return nil
		case <-timer.C:
			// in this case, the timer hit the 15 second timeout, so we received no heartbeats and need to cleanup
			if generalLock.cleanup() {
				return nil
			}

			// there must be another process waiting for the key so start processing again
			timer.Reset(generalLock.timeout)
		}
	}
}
