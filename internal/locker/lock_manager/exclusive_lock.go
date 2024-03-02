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

type releaseResponseCode int

const (
	failedRelease       releaseResponseCode = 0
	processedRelease                        = 1
	processedAndDestroy                     = 2
)

type exclusiveLock struct {
	// done is used to indicate that the lock is no longer processing heartbeats, claims, or releases
	done      chan struct{}
	blockDone chan struct{}

	// id of the current session that hold the claim to this lock
	sessionID      string
	sessionChannel chan string
	sessionIDLock  *sync.RWMutex

	// channels to manage heartbeat operations
	heartbeat         chan *v1locker.LockClaim
	heartbeatResponse chan bool

	// channels to manager releasing a client or lock
	release         chan *v1locker.LockClaim
	releaseResponse chan releaseResponseCode

	// channels to manage lock operations
	claim         chan func(time.Duration) string
	claimResponse chan time.Duration

	// timers for clients to know how long a lock is still valid for
	lastHeartbeatLock *sync.RWMutex
	lastHeartbeat     time.Time

	// timeout operation handlers
	timeout  func()
	timedOut chan bool
	// timing info for the timeout
	lockTimeout     time.Duration
	lockTimeoutLock *sync.RWMutex

	// stats around locks
	clientsWaitingForClaim *atomic.Uint64
}

func newExclusiveLock(timeout func()) *exclusiveLock {
	clientsWaitingForClaim := new(atomic.Uint64)

	return &exclusiveLock{
		done:      make(chan struct{}),
		blockDone: make(chan struct{}, 1),

		sessionChannel: make(chan string),
		sessionIDLock:  new(sync.RWMutex),

		heartbeat:         make(chan *v1locker.LockClaim),
		heartbeatResponse: make(chan bool),

		release:         make(chan *v1locker.LockClaim),
		releaseResponse: make(chan releaseResponseCode),

		claim:         make(chan func(time.Duration) string),
		claimResponse: make(chan time.Duration),

		lastHeartbeatLock: new(sync.RWMutex),

		timeout:  timeout,
		timedOut: make(chan bool),

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

		// processed a heartbeat, but there are no claims yet
		case <-exclusiveLock.heartbeat:
			exclusiveLock.heartbeatResponse <- false

		// processed a relase, but there are no claims yet
		case claim := <-exclusiveLock.release:
			// this is the case that a client has disconnected
			if claim == nil {
				if exclusiveLock.clientsWaitingForClaim.Add(^uint64(0)) == 0 {
					exclusiveLock.releaseResponse <- processedAndDestroy
					return nil
				}

				exclusiveLock.releaseResponse <- processedRelease
				continue
			}

			// this is the case that a release occured, but there is no claim
			exclusiveLock.releaseResponse <- failedRelease

		// client was able to claim the lock
		case exclusiveLock.claim <- exclusiveLock.processClaim:
			// respond with the new session id
			lockTimeout := <-exclusiveLock.claimResponse

			// start the new heartbeat loop operation
			ticker := time.NewTicker(lockTimeout)

		HEARTBEAT_LOOP:
			for {
				select {
				// server shutdown
				case <-ctx.Done():
					select {
					// if a timeout is processing, this should wait untill that is finished
					case <-exclusiveLock.blockDone:
						// wait and process the timeout
						<-exclusiveLock.timedOut

						if exclusiveLock.clientsWaitingForClaim.Add(^uint64(0)) == 0 {
							exclusiveLock.timedOut <- true
							return nil
						} else {
							exclusiveLock.timedOut <- false
						}
					default:
					}

					return nil

				// heartbeat
				case claim := <-exclusiveLock.heartbeat:
					fmt.Println("exclusive lock session id", exclusiveLock.sessionID)
					fmt.Println("claim session id", claim.SessionID)
					fmt.Println("equal", exclusiveLock.sessionID == claim.SessionID)

					if claim.SessionID == exclusiveLock.sessionID {
						// setup the last heartbeat record
						exclusiveLock.lastHeartbeatLock.Lock()
						exclusiveLock.lastHeartbeat = time.Now()
						exclusiveLock.lastHeartbeatLock.Unlock()

						// update the ticker
						ticker.Reset(exclusiveLock.lockTimeout)

						exclusiveLock.heartbeatResponse <- true
					} else {
						exclusiveLock.heartbeatResponse <- false
					}

				// timed out
				case <-ticker.C:
					ticker.Stop()

					fmt.Println("timed out")

					// clear the session id
					exclusiveLock.sessionIDLock.Lock()
					exclusiveLock.sessionID = ""
					exclusiveLock.sessionIDLock.Unlock()

					// set the current timeout
					exclusiveLock.lockTimeoutLock.Lock()
					exclusiveLock.lockTimeout = 0
					exclusiveLock.lockTimeoutLock.Unlock()

					// clear the last heartbeat time
					exclusiveLock.lastHeartbeatLock.Lock()
					exclusiveLock.lastHeartbeat = time.Time{}
					exclusiveLock.lastHeartbeatLock.Unlock()

					// need to run the cleanup in the background thread so the tree can grab the delete lock to check
					exclusiveLock.blockDone <- struct{}{}
					go func() {
						exclusiveLock.timeout()
					}()

				// will trigger on a timeout
				case <-exclusiveLock.timedOut:
					// always read from the block of done to clear it out
					<-exclusiveLock.blockDone

					if exclusiveLock.clientsWaitingForClaim.Add(^uint64(0)) == 0 {
						exclusiveLock.timedOut <- true
						return nil
					} else {
						exclusiveLock.timedOut <- false
					}

					break HEARTBEAT_LOOP

				// releasing a claim or a client was disconnected
				case claim := <-exclusiveLock.release:
					// this is the case that a client has disconnected
					if claim == nil {
						// should never have more than 0 clients at this point
						_ = exclusiveLock.clientsWaitingForClaim.Add(^uint64(0))
						exclusiveLock.releaseResponse <- processedRelease
						break HEARTBEAT_LOOP
					}

					// this is the case that a release occured with proper session id
					if claim.SessionID == exclusiveLock.sessionID {
						// reset the session id
						exclusiveLock.sessionIDLock.Lock()
						exclusiveLock.sessionID = ""
						exclusiveLock.sessionIDLock.Unlock()

						// set the current timeout
						exclusiveLock.lockTimeoutLock.Lock()
						exclusiveLock.lockTimeout = 0
						exclusiveLock.lockTimeoutLock.Unlock()

						// clear the heartbeat timer
						exclusiveLock.lastHeartbeatLock.Lock()
						exclusiveLock.lastHeartbeat = time.Time{}
						exclusiveLock.lastHeartbeatLock.Unlock()

						// stop the timeout ticker
						ticker.Stop()

						if exclusiveLock.clientsWaitingForClaim.Add(^uint64(0)) == 0 {
							exclusiveLock.releaseResponse <- processedAndDestroy
							return nil
						} else {
							exclusiveLock.releaseResponse <- processedRelease
							break HEARTBEAT_LOOP
						}
					} else {
						exclusiveLock.releaseResponse <- failedRelease
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
func (exclusiveLock *exclusiveLock) processClaim(lockTimeout time.Duration) string {
	// setup the new session id
	exclusiveLock.sessionIDLock.Lock()
	exclusiveLock.sessionID = idgenerator.UUID().ID()
	exclusiveLock.sessionIDLock.Unlock()

	// set the current timeout
	exclusiveLock.lockTimeoutLock.Lock()
	exclusiveLock.lockTimeout = lockTimeout
	exclusiveLock.lockTimeoutLock.Unlock()

	// setup the last heartbeat record
	exclusiveLock.lastHeartbeatLock.Lock()
	exclusiveLock.lastHeartbeat = time.Now()
	exclusiveLock.lastHeartbeatLock.Unlock()

	// inform the async process to continue
	exclusiveLock.claimResponse <- lockTimeout

	return exclusiveLock.sessionID
}

//	RETURNS
//	- <-chan struct{} - recieves a value when a lock is claimed. This channel does not close
//
// GetClaimChannel returns a channel any claims can be processed on. This also tracks the total number of clients waiting for
// a claim
func (exclusiveLock *exclusiveLock) GetClaimChannel() <-chan func(lockTimeout time.Duration) string {
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
	case exclusiveLock.heartbeat <- claim:
		if passed := <-exclusiveLock.heartbeatResponse; passed {
			fmt.Println("passed?", passed)
			return nil
		}

		fmt.Println("failed to pass")
		return &errors.ServerError{Message: "SessionID for the claim is invalid", StatusCode: http.StatusConflict}
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
	case exclusiveLock.release <- claim:
		response := <-exclusiveLock.releaseResponse

		switch response {
		case failedRelease:
			return false, &errors.ServerError{Message: "SessionID for the claim is invalid", StatusCode: http.StatusConflict}
		case processedRelease:
			return false, nil
		default: // processedAndDestroy:
			return true, nil
		}
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
	case exclusiveLock.release <- nil:
		response := <-exclusiveLock.releaseResponse

		switch response {
		case processedRelease:
			return false
		default: // processedAndDestroy:
			return true
		}
	case <-exclusiveLock.done:
		return false
	}
}

//	RETURNS
//	- bool - true iff there are no more clients waiting and can be released from the locker
//
// TimeOut a currently held lock and allow for another client to process the request
func (exclusiveLock *exclusiveLock) TimeOut() bool {
	exclusiveLock.timedOut <- true
	return <-exclusiveLock.timedOut
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
