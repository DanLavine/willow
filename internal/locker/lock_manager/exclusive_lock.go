package lockmanager

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type exclusiveLock struct {
	heartbeat chan struct{}
	release   chan struct{}

	lastHeartbeatLock *sync.RWMutex
	lastHeartbeat     time.Time

	timeout                func() bool
	lockTimeout            time.Duration
	clientsWaitingForClaim *atomic.Uint64
}

func newExclusiveLock(clientsWaitingForClaim *atomic.Uint64, lockTimeout time.Duration, timeout func() bool) *exclusiveLock {
	return &exclusiveLock{
		heartbeat: make(chan struct{}),
		release:   make(chan struct{}),

		lastHeartbeatLock: new(sync.RWMutex),

		timeout:                timeout,
		lockTimeout:            lockTimeout,
		clientsWaitingForClaim: clientsWaitingForClaim,
	}
}

func (exclusiveLock *exclusiveLock) Execute(ctx context.Context) error {
	defer func() {
		close(exclusiveLock.heartbeat)
		close(exclusiveLock.release)
	}()

	ticker := time.NewTicker(exclusiveLock.lockTimeout)
	exclusiveLock.setLastHeartbeatTime()

	for {
		select {
		// service was told to shutdown
		case <-ctx.Done():
			return nil

		// processed a heartbeat
		case exclusiveLock.heartbeat <- struct{}{}:
			exclusiveLock.setLastHeartbeatTime()
			ticker.Reset(exclusiveLock.lockTimeout)

		// processed a release
		case exclusiveLock.release <- struct{}{}:
			return nil

		// timed out
		case <-ticker.C:
			if !exclusiveLock.timeout() {
				// in the case that a `Release()`` request started to process the delete, but was paused
				// and has yet to read from the release channel. So we still need to send on the release channel
				exclusiveLock.release <- struct{}{}
			}
			return nil
		}
	}
}

func (exclusiveLock *exclusiveLock) setLastHeartbeatTime() {
	exclusiveLock.lastHeartbeatLock.Lock()
	defer exclusiveLock.lastHeartbeatLock.Unlock()

	exclusiveLock.lastHeartbeat = time.Now()
}

func (exclusiveLock *exclusiveLock) getLastHeartbeatTime() time.Time {
	exclusiveLock.lastHeartbeatLock.RLock()
	defer exclusiveLock.lastHeartbeatLock.RUnlock()

	return exclusiveLock.lastHeartbeat
}
