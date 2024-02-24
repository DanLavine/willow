package lockmanager

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type generalTimer struct {
	stop    chan struct{} // stop the entire processing chain
	release chan struct{} // release the currently heald lock

	// channel that clients waiting for a lock read from
	claimTrigger chan time.Duration

	// channe to inform a heartbeat has been recieved
	heartbeatChanLock *sync.RWMutex
	heartbeat         chan struct{}
	// callback for when a heartbeat fails
	heartbeatFailure func() bool

	// timing data about the last heartbeat
	heartbeatLock *sync.RWMutex
	lastHeartbeat time.Time

	timeout *atomic.Int64

	sessionID      string
	clientsWaiting *atomic.Int64
}

func newGeneralLock(heartbeatFailure func() bool) *generalTimer {
	counter := new(atomic.Int64)
	counter.Add(1)

	return &generalTimer{
		stop:    make(chan struct{}),
		release: make(chan struct{}),

		claimTrigger: make(chan time.Duration),

		heartbeatChanLock: new(sync.RWMutex),
		heartbeat:         make(chan struct{}),
		heartbeatFailure:  heartbeatFailure,

		heartbeatLock: new(sync.RWMutex),
		lastHeartbeat: time.Now(),

		timeout: new(atomic.Int64),

		clientsWaiting: counter,
	}
}

func (generalTimer *generalTimer) Execute(ctx context.Context) error {
	var ticker *time.Ticker

	defer func() {
		close(generalTimer.heartbeat)
		close(generalTimer.release)
	}()

	for {
		select {
		// shutdown signal
		case <-ctx.Done():
			// in this case, we received a shutdown signal from the server, so just cancel this threadand don't clean
			// anything up since we eventually want locks to persist through a server restart
			return nil

		// all clients waiting have been closed
		case <-generalTimer.stop:
			return nil

		// new client grabbed the timer, need to set the new time to what the client provided
		case newLockTimeout := <-generalTimer.claimTrigger:
			generalTimer.timeout.Store(int64(newLockTimeout))
			if ticker == nil {
				ticker = time.NewTicker(newLockTimeout)
			} else {
				ticker.Reset(newLockTimeout)
			}

		HEARTBEAT_LOOP:
			for {
				select {
				// in this case the server is shutting down
				case <-ctx.Done():
					return nil

				// in this case we processed a heartbeat, so reset the timer
				case generalTimer.heartbeat <- struct{}{}:
					generalTimer.setLastHeartbeat()
					ticker.Reset(time.Duration(generalTimer.timeout.Load()))
					continue HEARTBEAT_LOOP

				// in this case we processed a request to release the lock
				case generalTimer.release <- struct{}{}:

				// in this case, the timer hit the timeout so we need to release the lock
				case <-ticker.C:
					// if this did not process, then there was another operation calling release at the same time
					if !generalTimer.heartbeatFailure() {
						generalTimer.release <- struct{}{}
					}
				}

				// reset the heartbeat chan so any cliients currently wating to heartbeat can be closed.
				// this is the race where heartbeat request came in as a timeout or release also triggered.
				generalTimer.resetHeartbeat()

				// release the current heartbeat loop
				break HEARTBEAT_LOOP
			}
		}
	}
}

// operations to manage the clients waiting
func (generalTimer *generalTimer) AddWaitingClient() {
	generalTimer.clientsWaiting.Add(1)
}

func (generalTimer *generalTimer) RemoveWaitingClient() int64 {
	clientsWaiting := generalTimer.clientsWaiting.Add(-1)

	if clientsWaiting <= 0 {
		close(generalTimer.stop)
	}

	return clientsWaiting
}

// operations to manage the heartbeat
func (generalTimer *generalTimer) GetHeartbeater() chan struct{} {
	generalTimer.heartbeatChanLock.RLock()
	defer generalTimer.heartbeatChanLock.RUnlock()

	return generalTimer.heartbeat
}

func (generalTimer *generalTimer) resetHeartbeat() {
	// close the current channels
	close(generalTimer.heartbeat)

	// the the channel and reset it
	generalTimer.heartbeatChanLock.Lock()
	defer generalTimer.heartbeatChanLock.Unlock()

	generalTimer.heartbeat = make(chan struct{})
}

// operations for getting the time of the last heartbeat
func (generalTimer *generalTimer) setLastHeartbeat() {
	generalTimer.heartbeatLock.Lock()
	defer generalTimer.heartbeatLock.Unlock()

	generalTimer.lastHeartbeat = time.Now()
}

func (generalTimer *generalTimer) GetLastHeartbeat() time.Time {
	generalTimer.heartbeatLock.RLock()
	defer generalTimer.heartbeatLock.RUnlock()

	return generalTimer.lastHeartbeat
}
