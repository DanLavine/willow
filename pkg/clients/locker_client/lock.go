package lockerclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
)

// Lock interface defines the methods for an obtained lock from the Locker Service.
// This is kept as an interface to ensure that any code for the LockClient can be tested and do not need
// to setup the LockerSerivce so Locks work properly
//
//go:generate mockgen -imports v1locker="github.com/DanLavine/willow/pkg/models/api/locker/v1" -destination=lockerclientfakes/lock_mock.go -package=lockerclientfakes github.com/DanLavine/willow/pkg/clients/locker_client Lock
type Lock interface {
	//	RETURNS:
	//	- <-chan struct{} - can be monitored to know if a lock has been lost or was released
	//
	// Done can be used to monitor if a lock is released because of heartbeat failures from the client.
	Done() <-chan struct{}

	//	RETURNS:
	//	- error - from the service when realeasing the lock. If this happens the lock should be treated
	//	          as realesed from the client and will eventually time out service side.
	//
	// Release the currently held lock
	Release(ctx context.Context) error
}

type lock struct {
	// record stop heartbeating
	doneOnce *sync.Once
	done     chan struct{}

	// used to ensure only 1 delete operation proccesses
	heartbeating chan struct{}
	released     *atomic.Bool

	// remote server client configuration
	url    string
	client *http.Client

	// record error callback if configured. Can be used to monitor any unexpeded errors
	// with the remote service and record them
	releaseLockCallback func() // cleans up the tree in the locker client

	// Lock unique IDs created by the service
	lockID    string
	sessionID string

	// timeout for the configured Lock
	timeout time.Duration
}

func newLock(lockResponse *v1locker.Lock, url string, client *http.Client, heartbeatErrorCallback func(err error), releaseLockCallback func()) *lock {
	lock := &lock{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		heartbeating: make(chan struct{}),
		released:     new(atomic.Bool),

		client: client,
		url:    url,

		releaseLockCallback: releaseLockCallback,

		lockID:    lockResponse.State.LockID,
		sessionID: lockResponse.State.SessionID,
		timeout:   *lockResponse.Spec.Timeout,
	}

	go func() {
		defer func() {
			lock.stop()
			lock.released.Store(true)
			close(lock.heartbeating)
		}()

		// set ticker to be ((timeout - 10%) /3). This way we try and heartbeat at least 3 times before a failure occurs
		adjustedTimeout := *lockResponse.Spec.Timeout - (*lockResponse.Spec.Timeout / 10)
		ticker := time.NewTicker(adjustedTimeout / 3)
		timeoutTicker := time.NewTicker(lock.timeout)

		for {
			select {
			case <-lock.done:
				// release was called. can just escape
				return
			case <-timeoutTicker.C:
				// the timeout was reached. can just escape
				return
			case <-ticker.C:
				lockClaimData, err := api.ModelEncodeRequest(&v1locker.LockClaim{SessionID: lockResponse.State.SessionID})
				if err != nil {
					heartbeatErrorCallback(fmt.Errorf("failed to encode heartbeat request: %w", err))
					return
				}

				req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locks/%s/heartbeat", url, lockResponse.State.LockID), bytes.NewBuffer(lockClaimData))
				if err != nil {
					if heartbeatErrorCallback != nil {
						heartbeatErrorCallback(fmt.Errorf("failed to setup heartbeat request: %w", err))
					}
					return
				}
				req.Header.Add("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					if heartbeatErrorCallback != nil {
						heartbeatErrorCallback(fmt.Errorf("failed to heartbeat: %w", err))
					}

					continue
				}

				switch resp.StatusCode {
				case http.StatusOK:
					// this is the success case and the heartbeat passed
					timeoutTicker.Reset(lock.timeout)
				case http.StatusConflict, http.StatusBadRequest:
					// there was an error with the request body or session id
					apiError := &errors.Error{}
					if err := api.ModelDecodeResponse(resp, apiError); err != nil {
						if heartbeatErrorCallback != nil {
							heartbeatErrorCallback(err)
						}
					} else {
						if heartbeatErrorCallback != nil {
							heartbeatErrorCallback(apiError)
						}
					}

					// release the lock
					return
				default:
					if heartbeatErrorCallback != nil {
						heartbeatErrorCallback(fmt.Errorf("received an unexpected status code: %d", resp.StatusCode))
					}
				}
			}
		}
	}()

	return lock
}

//	PARAMTERS
//	- headers - optional http headers to apply to the request
//
//	RETURNS:
//	- error - from the service when realeasing the lock. If this happens the lock should be treated
//	          as realesed from the client and will eventually time out service side.
//
// Release the currently held lock
func (l *lock) Release(ctx context.Context) error {
	// release the actual lock
	if l.released.CompareAndSwap(false, true) {
		// stop heartbeating
		l.stop()

		// wait for the heartbeat process to stop
		<-l.heartbeating

		// Delete Lock request
		lockClaimData, err := api.ModelEncodeRequest(&v1locker.LockClaim{SessionID: l.sessionID})
		if err != nil {
			return fmt.Errorf("failed to encode release request. Server will need to timeout: %w", err)
		}

		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locks/%s", l.url, l.lockID), bytes.NewBuffer(lockClaimData))
		if err != nil {
			return err
		}
		clients.AddHeadersFromContext(req, ctx)
		req.Header.Add("Content-Type", "application/json")

		resp, err := l.client.Do(req)
		if err != nil {
			return err
		}

		// parse the release response
		switch resp.StatusCode {
		case http.StatusNoContent:
			// this is the success case and the Lock was deleted
			return nil
		case http.StatusConflict, http.StatusBadRequest:
			// there was an error with the request body or seession id
			apiError := &errors.Error{}
			if err := api.ModelDecodeResponse(resp, apiError); err != nil {
				return err
			}

			// release the lock
			return apiError
		default:
			return fmt.Errorf("unexpected response code from the remote locker service '%d'. Need to wait for the lock to expire remotely", resp.StatusCode)
		}
	} else {
		return fmt.Errorf("already released the lock")
	}
}

// Done can be used by the client to know when a lock has been released successfully
func (l *lock) Done() <-chan struct{} {
	return l.done
}

// close the release chan only once
func (l *lock) stop() {
	l.doneOnce.Do(func() {
		// remove the item from the client's BTree
		l.releaseLockCallback()

		// close done
		close(l.done)
	})
}

// What I think could be useful for a reload  on a process that has stopped for an update and restarted. I.E K8S node's would still
// have docker images for running JOBS that could be picked up and restart heartbeating that they are still processing. In this case
// the joibs would have a long time to run, but that is to be expected in those use cases

// func (l *lock) ReleaseWithoutAck() { }}

// func (l *lock) Save(diskDir string) {}

// func LoadItem(diskDir string) *Item { }
