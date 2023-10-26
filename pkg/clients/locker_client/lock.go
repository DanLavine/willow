package lockerclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type lock struct {
	// remote server client configuration
	client *http.Client
	url    string

	// callbacks if a lock is lost or an error
	lockLostCallback func(keyValues datatypes.KeyValues, err error)
	// Optional
	heartbeatErrorCallback func(keyValues datatypes.KeyValues, err error)
	// Optional
	deleteLockErrorCallback func(keyValues datatypes.KeyValues, err error)

	// key values for the lock
	keyValues datatypes.KeyValues

	// channel to signal that we should stop heartbeating
	done        chan struct{}
	releaseChan chan struct{}

	// lock unique session ID created by the service
	sessionID string

	// timeout for the configured lock
	timeout time.Duration
}

func (l *lock) Execute(ctx context.Context) error {
	ticker := time.NewTicker(l.timeout / 3)
	lastTick := time.Now()

	// close done when the lock has been released
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
				// and mimic that we lost the lock
				if time.Since(lastTick) >= l.timeout {
					l.lockLostCallback(l.keyValues, fmt.Errorf("could not heartbeat successfuly since the timeout. Releasing the local lock since remote is unreachable"))
					return nil
				}
			case 2:
				// lock has been lost and processed accordingly
				return nil
			}
		case <-ctx.Done():
			// stopping the client, so release the lock
			l.release()
			l.lockLostCallback(l.keyValues, nil)
			return nil
		case <-l.releaseChan:
			// stop the heartbeat loop fom the client perspective
			l.release()
			l.lockLostCallback(l.keyValues, nil)
			return nil
		}
	}
}

// heartbeat is managed by the goasync loop
//
//	RETURNS:
//	- int - 0 indicattes success, 1 indicates that the heartbeat failed, 2 indicates that the lock was lost and we can stop the async loop
func (l *lock) heartbeat() int {
	// heartbeat lock request body
	heartbeatLocksRequest := v1locker.HeartbeatLocksRequst{
		SessionIDs: []string{l.sessionID},
	}
	body, err := json.Marshal(heartbeatLocksRequest)
	if err != nil {
		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(l.keyValues, err)
		}
		return 1
	}

	// heartbeat lock
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/locks/heartbeat", l.url), bytes.NewBuffer(body))
	if err != nil {
		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(l.keyValues, err)
		}
		return 1
	}
	resp, err := l.client.Do(req)
	if err != nil {
		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(l.keyValues, fmt.Errorf("client closed unexpectedly when heartbeating: %w", err))
		}
		return 1
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// this is the success case and the lock was deleted
		return 0
	case http.StatusBadRequest:
		// there was an error with the request body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if l.heartbeatErrorCallback != nil {
				l.heartbeatErrorCallback(l.keyValues, fmt.Errorf("internal error. client unable to read response body: %w", err))
			}
			return 1
		}

		apiError := &api.Error{}
		if err = json.Unmarshal(respBody, apiError); err != nil {
			if l.heartbeatErrorCallback != nil {
				l.heartbeatErrorCallback(l.keyValues, fmt.Errorf("error paring server response body: %w", err))
			}
			return 1
		}

		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(l.keyValues, apiError)
		}
		return 1
	case http.StatusConflict:
		// there was an error processing one of the sessionIDs
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if l.heartbeatErrorCallback != nil {
				l.heartbeatErrorCallback(l.keyValues, fmt.Errorf("internal error. client unable to read response body: %w", err))
			}
			return 1
		}

		heartbeatErrors := &v1locker.HeartbeatLocksResponse{}
		if err = json.Unmarshal(respBody, heartbeatErrors); err != nil {
			if l.heartbeatErrorCallback != nil {
				l.heartbeatErrorCallback(l.keyValues, fmt.Errorf("error paring server response body: %w", err))
			}
			return 1
		}

		for _, heartbeatError := range heartbeatErrors.HeartbeatErrors {
			select {
			case <-l.releaseChan:
				// in this case, the heartbeat could have been processing when a delete request also came in.
				// so just double check that the error should really be reported back to the user. Otherwise
				// if this is the case, this is a no-op
			default:
				l.lockLostCallback(l.keyValues, fmt.Errorf(heartbeatError.Error))
			}
		}

		return 2
	default:
		if l.heartbeatErrorCallback != nil {
			l.heartbeatErrorCallback(l.keyValues, fmt.Errorf("received an unexpected status code: %d", resp.StatusCode))
		}
		return 1
	}
}

// called by the locker client
func (l *lock) releaseAndStopHeartbeat() {
	close(l.releaseChan)

	// wait for done to be close so we know the lock was deleted properly
	<-l.done
}

func (l *lock) release() {
	// delete lock request body
	deleteLockRequest := v1locker.DeleteLockRequest{
		SessionID: l.sessionID,
	}
	body, err := json.Marshal(deleteLockRequest)
	if err != nil {
		// nothing to do here if there is an error. Could report it, but no action to take
		return
	}

	// Delete lock
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/locks/delete", l.url), bytes.NewBuffer(body))
	if err != nil {
		// nothing to do here if there is an error. Could report it, but no action to take.
		// server will eventually drop the lock if we cannot make the request for soome reason
		return
	}
	resp, err := l.client.Do(req)
	if err != nil {
		// nothing to do here if there is an error. Could report it, but no action to take.
		// server will eventually drop the lock if we cannot make the request for soome reason
		return
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		// this is the success case and the lock was deleted
	case http.StatusBadRequest:
		// there was an error parsing the request body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			// nothing to do here if there is an error. Could report it, but no action to take.
			// server will eventually drop the lock if we cannot make the request for soome reason
			return
		}

		apiError := &api.Error{}
		if err = json.Unmarshal(respBody, apiError); err != nil {
			// nothing to do here if there is an error. Could report it, but no action to take.
			// server will eventually drop the lock if we cannot make the request for soome reason
		}
	default:
		// nothing to do here if there is an error. Could report it, but no action to take.
		// server will eventually drop the lock if we cannot make the request for soome reason
	}
}
