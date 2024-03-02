package lockerclient

import (
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
	Release() error
}

type lock struct {
	// record stop heartbeating
	doneOnce *sync.Once
	done     chan struct{}

	// used to ensure only 1 delete operation proccesses
	heartbeating chan struct{}
	released     *atomic.Bool

	// remote server client configuration
	url         string
	client      clients.HttpClient
	contentType string

	// record error callback if configured. Can be used to monitor any unexpeded errors
	// with the remote service and record them
	releaseLockCallback func() // cleans up the tree in the locker client

	// Lock unique IDs created by the service
	lockID    string
	sessionID string

	// timeout for the configured Lock
	timeout time.Duration
}

func newLock(lockResponse *v1locker.LockCreateResponse, url string, client clients.HttpClient, contentType string, heartbeatErrorCallback func(err error), releaseLockCallback func()) *lock {
	lock := &lock{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		heartbeating: make(chan struct{}),
		released:     new(atomic.Bool),

		client:      client,
		url:         url,
		contentType: contentType,

		releaseLockCallback: releaseLockCallback,

		lockID:    lockResponse.LockID,
		sessionID: lockResponse.SessionID,
		timeout:   lockResponse.LockTimeout,
	}

	go func() {
		defer func() {
			lock.stop()
			lock.released.Store(true)
			close(lock.heartbeating)
		}()

		// set ticker to be ((timeout - 10%) /3). This way we try and heartbeat at least 3 times before a failure occurs
		lastHeartbeat := time.Now()
		adjustedTimeout := lockResponse.LockTimeout - (lockResponse.LockTimeout / 10)
		ticker := time.NewTicker(adjustedTimeout / 3)

		for {
			select {
			case <-lock.done:
				// release was called. can just excape
				return
			case <-ticker.C:
				if time.Since(lastHeartbeat) >= adjustedTimeout {
					if heartbeatErrorCallback != nil {
						heartbeatErrorCallback(fmt.Errorf("could not heartbeat successfuly since the timeout. Releasing the local Lock since remote is unreachable"))
					}
					return
				}

				// need to heartbeat
				resp, err := client.Do(&clients.RequestData{
					Method: "POST",
					Path:   fmt.Sprintf("%s/v1/locks/%s/heartbeat", url, lockResponse.LockID),
					Model: &v1locker.LockClaim{
						SessionID: lockResponse.SessionID,
					},
				})

				if err != nil {
					if heartbeatErrorCallback != nil {
						heartbeatErrorCallback(fmt.Errorf("failed to heartbeat: %w", err))
					}

					continue
				}

				switch resp.StatusCode {
				case http.StatusOK:
					// this is the success case and the heartbeat passed
					lastHeartbeat = time.Now()
				case http.StatusConflict, http.StatusBadRequest:
					// there was an error with the request body or seession id
					apiError := &errors.Error{}
					if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
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

//	RETURNS:
//	- error - from the service when realeasing the lock. If this happens the lock should be treated
//	          as realesed from the client and will eventually time out service side.
//
// Release the currently held lock
func (l *lock) Release() error {
	// release the actual lock
	if l.released.CompareAndSwap(false, true) {
		// stop heartbeating
		l.stop()

		// wait for the heartbeat process to stop
		<-l.heartbeating

		// Delete Lock
		resp, err := l.client.Do(&clients.RequestData{
			Method: "DELETE",
			Path:   fmt.Sprintf("%s/v1/locks/%s", l.url, l.lockID),
			Model: &v1locker.LockClaim{
				SessionID: l.sessionID,
			},
		})

		if err != nil {
			return err
		}

		switch resp.StatusCode {
		case http.StatusNoContent:
			// this is the success case and the Lock was deleted
			return nil
		case http.StatusConflict, http.StatusBadRequest:
			// there was an error with the request body or seession id
			apiError := &errors.Error{}
			if err := api.DecodeAndValidateHttpResponse(resp, apiError); err != nil {
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
