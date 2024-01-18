package heartbeater

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Heartbeater interface {
	Execute(ctx context.Context) error

	Stop() bool

	Start() bool

	Heartbeat() bool

	GetLastHeartbeat() time.Time
}

type heartbeater struct {
	// record keeping of when the last heartbeat occurred
	lastHeartbeatLock *sync.RWMutex
	lastHeartbeat     time.Time

	// channel to know if any blocked calles can br released because the heartbeater has stopped
	doneChan chan struct{}
	// channel to know that the heartbeater should start executing
	started *atomic.Bool
	start   chan struct{}
	// channel to stop the heartbeater
	stopChan chan struct{}
	// channel to signal a heartbeat and reset the heartbeat timer
	heartbeaterChan chan struct{}

	// how long until the heartbeat times out if no heartbeats are received
	timeout time.Duration

	// calbaack to run when shutdown is processed
	onShutdown func()
	// callback to run when the heartbeater times out
	onTimeout func()
}

//	PARAMETERS:
//	- timeout - how long until timeout occurs since the last heartbeat
//	- onShutdown - (optional) callback function to run when the shutdown process is received
//	- onTimeout - callbacck function to run when a timeout occurs
//
//	RETURNS:
//	- *heartbeater - private instance of the heartbeater
//	- error - error with any of the parameters
//
// New sets up a heartbeater processes that can be mnaged by goAsync.
func New(timeout time.Duration, onShutdown, onTimeout func()) (*heartbeater, error) {
	if onTimeout == nil {
		return nil, fmt.Errorf("onTimeout cannot be nil")
	}

	return &heartbeater{
		lastHeartbeatLock: new(sync.RWMutex),

		doneChan:        make(chan struct{}),
		started:         new(atomic.Bool),
		start:           make(chan struct{}),
		stopChan:        make(chan struct{}),
		heartbeaterChan: make(chan struct{}),
		timeout:         timeout,
		onShutdown:      onShutdown,
		onTimeout:       onTimeout,
	}, nil
}

// Execute expects a single instance of Start to be called
func (h *heartbeater) Execute(ctx context.Context) error {
	defer close(h.doneChan)

	// wait for start or stop instruction to be called
	select {
	case <-h.start:
	case <-h.stopChan:
		return nil
	}

	timer := time.NewTicker(h.timeout)
	h.setLastHeartbeat()

	for {
		select {
		case <-ctx.Done():
			// in this case, the service is shutting down
			if h.onShutdown != nil {
				h.onShutdown()
			}

			return nil
		case <-h.stopChan:
			// in this case, a stop was triggered by the caller
			return nil
		case <-timer.C:
			// in this case the timer has hit the timeout limit. assumed the resource has failed.
			h.onTimeout()

			return nil
		case <-h.heartbeaterChan:
			// in this case, the client sent a heartbeat so reset the time
			h.setLastHeartbeat()
			timer.Reset(h.timeout)
		}
	}
}

//	PARAMETERS:
//	- onStop - callback to run if the heartbeater was stopped through this function call
//
//	RETURNS:
//	- bool - TRUE iff the heartbeater was stopped through this call
//	- error - returns an error if there is a problem with the parameters
//
// Stop the heartbeter. This will panic if onStop is nil
func (h *heartbeater) Stop() bool {
	select {
	case h.stopChan <- struct{}{}:
		return true
	case <-h.doneChan:
		// either multiple calls to stop happened. Or there was a race wherre timeout and stop were called in parallel
	}

	return false
}

// Start is used to ensure that the heartbater process has started in an expected fashion and must be called for Execute to stop processing
func (h *heartbeater) Start() bool {
	if !h.started.Swap(true) {
		select {
		case h.start <- struct{}{}:
			return true
		case <-h.doneChan:
			// process was stopped
		}
	}

	return false
}

//	RETURNS:
//	- bool - TRUE iff the heartneat was processed
//
// Signal the heartbeter that we recieved a heartbeat and to reset the timer
func (h *heartbeater) Heartbeat() bool {
	select {
	case h.heartbeaterChan <- struct{}{}:
		return true
	case <-h.doneChan:
		// either async calls to stop or a timeout happened if this processes, then we know that this parallel process happened
	}

	return false
}

// GetLastHeartbeat is a thread safe way to get the time for when the last heartbeat occurred
func (h *heartbeater) GetLastHeartbeat() time.Time {
	h.lastHeartbeatLock.Lock()
	defer h.lastHeartbeatLock.Unlock()

	if (h.lastHeartbeat == time.Time{}) {
		return time.Now()
	}

	return h.lastHeartbeat
}

func (h *heartbeater) setLastHeartbeat() {
	h.lastHeartbeatLock.Lock()
	defer h.lastHeartbeatLock.Unlock()

	h.lastHeartbeat = time.Now()
}
