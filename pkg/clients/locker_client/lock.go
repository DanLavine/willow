package lockerclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
)

type Lock interface {
	// Done can be used to monitor if a lock is released
	Done() <-chan struct{}

	// Release can be used to release the currently held lock
	Release() error
}

type lock struct {
	// used to ensure only 1 delete operation proccesses
	lock     *sync.Mutex
	released bool

	// remote server client configuration
	client *http.Client
	url    string

	// record error callback if configured. Can be used to monitor any unexpeded errors
	// with the remote service and record them
	heartbeatErrorCallback func(err error) // optional
	releaseLockCallback    func()          // cleans up the tree in the locker client

	// channels to signal that we should stop heartbeating
	done chan struct{}
	// release is called when only 1 lock wants to be stopped
	realeasOnce *sync.Once
	releaseChan chan struct{}

	// Lock unique session ID created by the service
	sessionID string

	// timeout for the configured Lock
	timeout time.Duration
}

func newLock(sessionID string, timeout time.Duration, client *http.Client, url string, heartbeatErrorCallback func(err error), releaseLockCallback func()) *lock {
	return &lock{
		lock:     new(sync.Mutex),
		released: false,

		client: client,
		url:    url,

		releaseLockCallback:    releaseLockCallback,
		heartbeatErrorCallback: heartbeatErrorCallback,

		done:        make(chan struct{}),
		realeasOnce: new(sync.Once),
		releaseChan: make(chan struct{}),

		sessionID: sessionID,
		timeout:   timeout,
	}
}

func (l *lock) Execute(ctx context.Context) error {
	ticker := time.NewTicker(l.timeout / 3)
	lastTick := time.Now()

	// close done when the Lock has been released
	defer close(l.done)

	for {
		select {
		case tickTime := <-ticker.C:
			// need to heartbeat
			switch l.heartbeat() {
			case 0:
				// on successful heartbeat, reset the ticker
				ticker.Reset(l.timeout / 3)
				lastTick = tickTime
			case 1:
				// must be some sort of error on service side. So stop the ticker since we don't know what the actual issue is
				// and mimic that we lost the Lock
				if time.Since(lastTick) >= l.timeout {
					if l.heartbeatErrorCallback != nil {
						l.heartbeatErrorCallback(fmt.Errorf("could not heartbeat successfuly since the timeout. Releasing the local Lock since remote is unreachable"))
					}

					l.releaseLockCallback()
					return nil
				}
			case 2:
				// Lock has been lost and processed accordingly
				return nil
			}
		case <-ctx.Done():
			// stopping the client, so release the Lock
			l.releaseLockCallback()
			_ = l.release() // ignore this error, we are shutting down anyways
			return nil
		case <-l.releaseChan:
			// stop the heartbeat loop fom the client perspective
			return nil
		}
	}
}

// heartbeat is managed by the goasync loop
//
//	RETURNS:
//	- int - 0 indicattes success, 1 indicates that the heartbeat failed, 2 indicates that the Lock was lost and we can stop the async loop
func (l *lock) heartbeat() int {
	// heartbeat Lock
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locker/%s/heartbeat", l.url, l.sessionID), nil)
	if err != nil {
		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(err)
		}
		return 1
	}

	resp, err := l.client.Do(req)
	if err != nil {
		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(fmt.Errorf("client closed unexpectedly when heartbeating: %w", err))
		}
		return 1
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// this is the success case and the Lock was deleted
		return 0
	case http.StatusBadRequest, http.StatusConflict:
		// in both cases, try and read a request body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if l.heartbeatErrorCallback != nil {
				l.heartbeatErrorCallback(fmt.Errorf("internal error. client unable to read response body: %w", err))
			}
			return 1
		}

		switch resp.StatusCode {
		case http.StatusBadRequest:
			// there was an error with the request body
			apiError := &api.Error{}
			if err = json.Unmarshal(respBody, apiError); err != nil {
				if l.heartbeatErrorCallback != nil {
					l.heartbeatErrorCallback(fmt.Errorf("error paring server response body: %w", err))
				}
				return 1
			}

			if l.heartbeatErrorCallback != nil {
				l.heartbeatErrorCallback(apiError)
			}
			return 1
		default: // http.StatusConflict
			// there was an error with the sessionID
			heartbeatError := &v1locker.HeartbeatError{}
			if err = json.Unmarshal(respBody, heartbeatError); err != nil {
				if l.heartbeatErrorCallback != nil {
					l.heartbeatErrorCallback(fmt.Errorf("error paring server response body: %w", err))
				}
				return 1
			}

			// record the error and release the lock
			if l.heartbeatErrorCallback != nil {
				l.heartbeatErrorCallback(fmt.Errorf(heartbeatError.Error))
			}

			l.releaseLockCallback()

			return 2
		}
	default:
		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(fmt.Errorf("received an unexpected status code: %d", resp.StatusCode))
		}
		return 1
	}
}

// Release the currently held lock
func (l *lock) Release() error {
	// release the releases chan
	l.closeRelease()

	// remove the lock from the tree
	l.releaseLockCallback()

	// wait for the heartbeat process to stop
	<-l.done

	// release the actual lock
	return l.release()
}

// Done can be used by the client to know when a lock has been released successfully
func (l *lock) Done() <-chan struct{} {
	return l.done
}

// make a call to delete the lock from the remote service
func (l *lock) release() error {
	l.lock.Lock()
	defer func() {
		l.released = true
		l.lock.Unlock()
	}()

	// have already been released. Don't allow for multiple calls to mistakenly report that we cannot dele
	// the lock from the server side
	if l.released {
		return nil
	}

	// delete Lock request body
	deleteLockRequest := v1locker.DeleteLockRequest{
		SessionID: l.sessionID,
	}
	body, err := json.Marshal(deleteLockRequest)
	if err != nil {
		return err
	}

	// Delete Lock
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locker/delete", l.url), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		// this is the success case and the Lock was deleted
		return nil
	case http.StatusBadRequest:
		// there was an error parsing the request body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("internal error. client unable to read response body: %w", err)
		}

		apiError := &api.Error{}
		if err = json.Unmarshal(respBody, apiError); err != nil {
			return fmt.Errorf("error paring server response body: %w", err)
		}

		return apiError
	default:
		return fmt.Errorf("unexpected response code from the remote locker service. Need to wait for the lock to expire remotely")
	}
}

// close the release chan only once
func (l *lock) closeRelease() {
	l.realeasOnce.Do(func() {
		close(l.releaseChan)
	})
}
