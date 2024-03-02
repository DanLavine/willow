package lockmanager

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DanLavine/willow/internal/idgenerator"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
)

type channelAction struct {
	// used to indicate when an item can be processed.
	okToProcess bool

	// indicates how an item to be pprocessed should be treated
	ok bool
}

type exclusiveLock struct {
	// done is used to indicate that the lock is no longer processing heartbeats, claims, or releases
	done chan struct{}

	// id of the current session that hold the claim to this lock
	sessionID      string
	sessionChannel chan string
	sessionIDLock  *sync.RWMutex

	// channels to manage lock operations
	claim     chan time.Duration
	heartbeat chan channelAction
	release   chan channelAction

	// timers for clients to know how long a lock is still valid for
	lastHeartbeatLock *sync.RWMutex
	lastHeartbeat     time.Time

	// timeout operation handlers
	timeout         func() channelAction
	timedOut        chan bool
	lockTimeout     time.Duration
	lockTimeoutLock *sync.RWMutex

	// stats around locks
	clientsWaitingForClaim *atomic.Uint64
}

func newExclusiveLock(timeout func() channelAction) *exclusiveLock {
	clientsWaitingForClaim := new(atomic.Uint64)

	return &exclusiveLock{
		done: make(chan struct{}),

		sessionChannel: make(chan string),
		sessionIDLock:  new(sync.RWMutex),

		claim:     make(chan time.Duration),
		heartbeat: make(chan channelAction),
		release:   make(chan channelAction),

		lastHeartbeatLock: new(sync.RWMutex),

		timeout:         timeout,
		timedOut:        make(chan bool),
		lockTimeoutLock: new(sync.RWMutex),

		clientsWaitingForClaim: clientsWaitingForClaim,
	}
}

func (exclusiveLock *exclusiveLock) Execute(ctx context.Context) error {
	defer close(exclusiveLock.done)

	for {
		select {
		// service was told to shutdown
		case <-ctx.Done():
			return nil

		// api called heartbeat but nothing is claimed. This case can hit when a client accidently sends multiple
		// release requests and they process before a client waitig actually processes a new claim
		case exclusiveLock.heartbeat <- channelAction{okToProcess: false}:
			// nothing to do

		// api called release but nothing is claimed. This case can hit when a client accidently sends heartbeats
		// after a Release requests and they process before a client waitig actually processes a new claim
		//
		// OR
		//
		// this can be called when a client is closed and has not processed a request
		case exclusiveLock.release <- channelAction{okToProcess: false}:
			if release := <-exclusiveLock.release; release.okToProcess {
				if release.ok {
					return nil
				}
			}

		// processed a claim
		case lockTimeout := <-exclusiveLock.claim:
			sessionID := idgenerator.UUID().ID()

			exclusiveLock.sessionIDLock.Lock()
			exclusiveLock.sessionID = sessionID
			exclusiveLock.sessionIDLock.Unlock()

			exclusiveLock.lockTimeoutLock.Lock()
			exclusiveLock.lockTimeout = lockTimeout
			exclusiveLock.lockTimeoutLock.Unlock()

			exclusiveLock.sessionChannel <- sessionID

			fmt.Println("tsimeout for ticker", lockTimeout)
			exclusiveLock.setLastHeartbeatTime()
			ticker := time.NewTicker(lockTimeout)

		HEARTBEAT_LOOP:
			select {
			// service was told to shutdown
			case <-ctx.Done():
				ticker.Stop()
				return nil

			// processed a heartbeat
			case exclusiveLock.heartbeat <- channelAction{okToProcess: true}:
				// wait for the heartbeat operation to finish
				if heartbeat := <-exclusiveLock.heartbeat; heartbeat.okToProcess {
					if heartbeat.ok {
						// rest the ticker for the next timeout
						exclusiveLock.setLastHeartbeatTime()
						ticker.Reset(exclusiveLock.getLockTimeout())
					}
				}
				goto HEARTBEAT_LOOP

			// processed a release
			case exclusiveLock.release <- channelAction{okToProcess: true}:
				if release := <-exclusiveLock.release; release.okToProcess {
					// release was processed and there are no more clients waiting
					if release.ok {
						return nil
					}

					// release was processed, but there are still clients waiting
					ticker.Stop()
				} else {
					// this case we recieved a release, but the claim was invalid
					goto HEARTBEAT_LOOP
				}

			// timed out
			case <-ticker.C:
				ticker.Stop()

				// in this case the timeout operation grabbed the lock to the tree and processed everyhing properly
				if action := exclusiveLock.timeout(); action.okToProcess {
					// in this case the timeout operation grabbed the lock to the tree and processed everyhing properly
					if action.ok {
						return nil
					}
				} else {
					// in the case that a `Release()` request started to process the delete, but the thread was then paused.
					// It can still claim the lock to the tree. In that case, we need to ensure that release processes as normal
					exclusiveLock.release <- channelAction{okToProcess: true, ok: true}
					// release was processed and there are no more clients waiting
					if release := <-exclusiveLock.release; release.okToProcess {
						// release was processed and there are no more clients waiting
						if release.ok {
							return nil
						}
					}
				}
			}

		}
	}
}

//	PARAMETERS:
//	- timeout - how long a lock should last until it times out
//
//	RETURNS
//	- string - the sessionID for the claim
//
// Claim a lock. This will set the unique sessionID if the claim has been captured
func (exclusiveLock *exclusiveLock) ProcessClaim(timeout time.Duration) string {
	return <-exclusiveLock.sessionChannel
}

//	RETURNS
//	- <-chan struct{} - recieves a value when a lock is claimed. This channel does not close
//
// GetClaimChannel returns a channel any claims can be processed on. This also tracks the total number of clients waiting for
// a claim
func (exclusiveLock *exclusiveLock) GetClaimChannel() chan<- time.Duration {
	exclusiveLock.clientsWaitingForClaim.Add(1)
	return exclusiveLock.claim
}

//	PARAMETERS:
//	- claim - the unique claim generated when the lock was created
//
//	RETURNS
//	- *errors.ServerError - error encountered when heartbeating the lock
//
// Heartbeat a currently held lock
func (exclusiveLock *exclusiveLock) Heartbeat(claim *v1locker.LockClaim) *errors.ServerError {
	select {
	case action := <-exclusiveLock.heartbeat:
		// this case we are in a proper heartbeat loop
		if action.okToProcess {
			if exclusiveLock.sessionID == claim.SessionID {
				exclusiveLock.heartbeat <- channelAction{okToProcess: true, ok: true}
				return nil
			}

			// wrong session id
			exclusiveLock.heartbeat <- channelAction{okToProcess: false, ok: false}
			return &errors.ServerError{Message: "SessionID for the claim is invalid", StatusCode: http.StatusConflict}
		}

		// this is the case that heartbeat request processed after a release or timeout
		return &errors.ServerError{Message: "Lock is not currently claimed", StatusCode: http.StatusConflict}
	case <-exclusiveLock.done:
		return errors.ServerShutdown
	}
}

//	PARAMETERS:
//	- claim - the unique claim generated when the lock was created
//
//	RETURNS
//	- bool - true iff there are no more clients waiting and can be released from the locker
//	- *errors.ServerError - error encountered when heartbeating the lock
//
// Release a currently held lock and allow for another client to process the request
func (exclusiveLock *exclusiveLock) ReleaseLock(claim *v1locker.LockClaim) (bool, *errors.ServerError) {
	select {
	case action := <-exclusiveLock.release:
		// this is the case that we have a proper claim
		if action.okToProcess {
			// client requested the release with the proper session, or the `action.ok` is true because we need to proceess a timeout
			if exclusiveLock.sessionID == claim.SessionID || action.ok {
				// all cases, can remove the sessionID to be claimed again
				exclusiveLock.sessionIDLock.Lock()
				exclusiveLock.sessionID = ""
				exclusiveLock.sessionIDLock.Unlock()

				// on releasing, there are no more waiting clients so can delete the lock
				if exclusiveLock.clientsWaitingForClaim.Add(^uint64(0)) == 0 {
					exclusiveLock.release <- channelAction{okToProcess: true, ok: true}
					return true, nil
				}

				// there are still clients waiting so cannot release the lock
				exclusiveLock.release <- channelAction{okToProcess: true, ok: false}
				return false, nil
			}

			// this case there was multiple deletes processed for a single claim
			exclusiveLock.release <- channelAction{okToProcess: false, ok: false}
			return false, &errors.ServerError{Message: "SessionID for the claim is invalid", StatusCode: http.StatusConflict}
		}

		// this is the case where a claim is empty
		exclusiveLock.release <- channelAction{okToProcess: false, ok: false}

		return false, &errors.ServerError{Message: "Lock is not currently claimed", StatusCode: http.StatusConflict}
	case <-exclusiveLock.done:
		return false, errors.ServerShutdown
	}
}

//	PARAMETERS:
//	- claim - the unique claim generated when the lock was created
//
//	RETURNS
//	- bool - true iff there are no more clients waiting and can be released from the locker
//	- *errors.ServerError - error encountered when heartbeating the lock
//
// ReleaseLostClient is called from the manager when a client disconnects
func (exclusiveLock *exclusiveLock) ReleaseLostClient() bool {
	select {
	case <-exclusiveLock.release:
		// on releasing, there are no more waiting clients so can delete the lock
		if exclusiveLock.clientsWaitingForClaim.Add(^uint64(0)) == 0 {
			exclusiveLock.release <- channelAction{okToProcess: true, ok: true}
			return true
		}

		// there are still clients waiting so cannot release the lock
		exclusiveLock.release <- channelAction{okToProcess: true, ok: true}
		return false

	case <-exclusiveLock.done:
		return false
	}
}

//	RETURNS
//	- bool - true iff there are no more clients waiting and can be released from the locker
//
// TimeOut a currently held lock and allow for another client to process the request
func (exclusiveLock *exclusiveLock) TimeOut() channelAction {
	// all cases, can remove the sessionID to be claimed again
	exclusiveLock.sessionIDLock.Lock()
	exclusiveLock.sessionID = ""
	exclusiveLock.sessionIDLock.Unlock()

	// on releasing, there are no more waiting clients so can delete the lock
	if exclusiveLock.clientsWaitingForClaim.Add(^uint64(0)) == 0 {
		return channelAction{okToProcess: true, ok: true}
	}

	// there are still clients waiting so cannot release the lock
	return channelAction{okToProcess: true, ok: false}
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

func (exclusiveLock *exclusiveLock) getLockTimeout() time.Duration {
	exclusiveLock.lockTimeoutLock.RLock()
	defer exclusiveLock.lockTimeoutLock.RUnlock()

	return exclusiveLock.lockTimeout
}

func (exclusiveLock *exclusiveLock) getSessionID() string {
	exclusiveLock.sessionIDLock.RLock()
	defer exclusiveLock.sessionIDLock.RUnlock()

	return exclusiveLock.sessionID
}
