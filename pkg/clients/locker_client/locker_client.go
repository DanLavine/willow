package lockerclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DanLavine/goasync"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"golang.org/x/net/http2"
)

type lockerclient struct {
	// mange releasing all locks
	done   chan struct{}
	cancel context.CancelFunc

	// client to connect with the remote Locker service
	url    string
	client *http.Client

	// callback to client if a lock was somehow released on the server beccause heartbeats are failing
	lockLostCallback       func(keyValues datatypes.KeyValues, err error)
	heartbeatErrorCallback func(keyValues datatypes.KeyValues, err error)

	// each item in the locks tree's value is a lock
	locks btreeassociated.BTreeAssociated

	// async manager to handle heartbeats
	asyncManager goasync.AsyncTaskManager
}

func NewLockerClient(cfg *Config, lockLostCallback, heartbeatErrorCallback func(keyValues datatypes.KeyValues, err error)) (*lockerclient, error) {
	locks := btreeassociated.NewThreadSafe()

	if err := cfg.Vaidate(); err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	httpClient.Transport = &http2.Transport{}

	if cfg.LockerCAFile != "" {
		httpClient.Transport = &http2.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cfg.cert},
				RootCAs:      cfg.rootCAs,
			},
		}
	}

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())
	go func() {
		_ = asyncManager.Run(ctx)
		close(done)
	}()

	return &lockerclient{
		done:                   done,
		cancel:                 cancel,
		url:                    cfg.URL,
		client:                 httpClient,
		lockLostCallback:       lockLostCallback,
		heartbeatErrorCallback: heartbeatErrorCallback,
		locks:                  locks,
		asyncManager:           goasync.NewTaskManager(goasync.RelaxedConfig()),
	}, nil
}

// Obtain a lock for a particular set of KeyValues. This blocks until the desired lock is obtained, or the context is canceled
//
// PARAMS:
//   - ctx - Context that can be used to cancel the request
func (lc *lockerclient) ObtainLock(ctx context.Context, keyValues datatypes.KeyValues, timeout time.Duration) error {
	// create lock request body
	createLockRequest := v1locker.CreateLockRequest{
		KeyValues: keyValues,
		Timeout:   timeout,
	}
	body, err := json.Marshal(createLockRequest)
	if err != nil {
		// should never actually hit this
		return err
	}

	var lockErr error

	// make the request

	// should check to make sure we don't already have the lock
	onCreate := func() any {
		for {
			req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/locks/create", lc.url), bytes.NewBuffer(body))
			if err != nil {
				// this should never actually hit
				return err
			}
			resp, err := lc.client.Do(req)
			if err != nil {
				// we didn't make the request or was canceled
				select {
				case <-ctx.Done():
					// client was canceled so don't return an error
					return nil
				default:
					lockErr = err
					return nil
				}
			}

			switch resp.StatusCode {
			case http.StatusCreated:
				// created the lock, need to record the session id and start hearbeating
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					// shouldn't actuall hit this
					lockErr = err
					return nil
				}

				createLockResponse := &v1locker.CreateLockResponse{}
				if err = json.Unmarshal(respBody, createLockResponse); err != nil {
					// shouldn't actuall hit this
					lockErr = err
					return nil
				}

				// wrapper is needed so we ensure that the lock is removed from the tree on removal
				lostLockWrapper := func(keyValues datatypes.KeyValues, err error) {
					canDelete := func(_ any) bool {
						return true
					}
					lc.locks.Delete(keyValues, canDelete)

					// only call when there was an error
					if err != nil {
						lc.lockLostCallback(keyValues, err)
					}
				}

				newLock := &lock{
					client:                 lc.client,
					url:                    lc.url,
					lockLostCallback:       lostLockWrapper,
					heartbeatErrorCallback: lc.heartbeatErrorCallback,
					keyValues:              keyValues,
					releaseChan:            make(chan struct{}),
					sessionID:              createLockResponse.SessionID,
					timeout:                createLockRequest.Timeout,
				}

				// setup the heartbeat operation
				if err = lc.asyncManager.AddExecuteTask(createLockResponse.SessionID, newLock); err != nil {
					// could not add the task to execute, must be shutting down so release the lock
					lockErr = fmt.Errorf("client has been closed, will not obtain lock")
					newLock.releaseAndStopHeartbeat()

					return nil
				}

				return newLock
			case http.StatusBadRequest:
				// there was an error with the request. possibly a mismatch on client server versions
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					// shouldn't actuall hit this
					lockErr = err
					return nil
				}

				apiError := &api.Error{}
				if err = json.Unmarshal(respBody, apiError); err != nil {
					// shouldn't actuall hit this
					lockErr = err
					return nil
				}

				lockErr = apiError
				return nil
			case http.StatusBadGateway:
				// server is restarting so retry the request
				time.Sleep(time.Second)
			default:
				lockErr = fmt.Errorf("received an unexpected status code: %d", resp.StatusCode)
				return nil
			}
		}
	}

	onFind := func(_ any) {
		// nothing to do here
	}

	// create or find the lock if we already have it
	lc.locks.CreateOrFind(keyValues, onCreate, onFind)

	return lockErr
}

func (lc *lockerclient) ReleaseLock(keyValues datatypes.KeyValues) error {
	var lockErr error
	canDelete := func(item any) bool {
		lock := item.(*btreeassociated.AssociatedKeyValues).Value().(*lock)
		lock.releaseAndStopHeartbeat()

		return true
	}

	lc.locks.Delete(keyValues, canDelete)
	return lockErr
}

// Release all locks currently held
func (lc *lockerclient) ReleaseLocks() {

}

// Release locks can be called to release all held locks as part of a shutdown process
func (lc *lockerclient) ReleaseLocksAndCloseClient() {
	// cancel asyc task manager
	lc.cancel()

	// wait for all clients to close
	<-lc.done
}
