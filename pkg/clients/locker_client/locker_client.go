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
	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"golang.org/x/net/http2"
)

type LockerClient interface {
	Healthy() error

	ObtainLock(ctx context.Context, keyValues datatypes.KeyValues, timeout time.Duration) (Lock, error)

	Done() <-chan struct{}
}

type lockerclient struct {
	// used to know if the async manager is done
	done chan struct{}

	// client to connect with the remote Locker service
	url    string
	client *http.Client

	// callback to client if a lock was somehow released on the server beccause heartbeats are failing
	heartbeatErrorCallback func(keyValues datatypes.KeyValues, err error)

	// each item in the locks tree's value is a lock
	locks btreeassociated.BTreeAssociated

	// async manager to handle heartbeats
	asyncManager goasync.AsyncTaskManager
}

//	PARAMS
//	- ctx - Context for the http(s) client. This must be closed to close the managed client and will trigger a close of all held locks
//	- cfg - configuration for the http client
//	- heartbeatErrorCallback (optional) - callback for heartbeat errors. Mainly used to log any errors the managed client to the locker service might be experiencing
//
//	RETURNS:
//
// Setup a new client to the remote locker service. This client automatically manages hertbeats for any obtained locks and
// will notify the user if a lock is lost at some point
func NewLockerClient(ctx context.Context, cfg *clients.Config, heartbeatErrorCallback func(keyValue datatypes.KeyValues, err error)) (LockerClient, error) {
	if ctx == nil || ctx == context.TODO() || ctx == context.Background() {
		return nil, fmt.Errorf("cannot use provided context. The context must be canceled by the caller to cleanup async resource management")
	}

	if err := cfg.Vaidate(); err != nil {
		return nil, err
	}

	done := make(chan struct{})
	asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())

	locks := btreeassociated.NewThreadSafe()
	httpClient := &http.Client{}

	if cfg.CAFile != "" {
		httpClient.Transport = &http2.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cfg.Cert()},
				RootCAs:      cfg.RootCAs(),
			},
		}
	}

	lockerClient := &lockerclient{
		done:                   done,
		url:                    cfg.URL,
		client:                 httpClient,
		heartbeatErrorCallback: heartbeatErrorCallback,
		locks:                  locks,
		asyncManager:           asyncManager,
	}

	go func() {
		defer close(done)
		_ = asyncManager.Run(ctx)
	}()

	return lockerClient, nil
}

func (lc *lockerclient) Healthy() error {
	// setup and make request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/health", lc.url), nil)
	if err != nil {
		// this should never actually hit
		return fmt.Errorf("internal error setting up http request: %w", err)
	}

	resp, err := lc.client.Do(request)
	if err != nil {
		return fmt.Errorf("unable to make request to locker service: %w", err)
	}

	// parse the response
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("unexpected status code checking health: %d", resp.StatusCode)
	}
}

//	PARAMS:
//	- ctx - Context that can be used to cancel the blocking requst trying to obtain the lock. NOTE: once a lock is obtained, release must be called
//	- keyValues - Key Values to obtain the unique lock for
//	- timeout - How long the lock should remain valid for if the heartbeats are failing
//
//	RETURNS
//	- Lock - lock object that can be used to release a lock, and monitor if a lock is lost for some reason
//	- error - any errors encountered when obtaining the lock
//	NOTE: is both Lock and error are nil, the context must have been canceled obtaining the lock
//
// Obtain a lock for a particular set of KeyValues. This blocks until the desired lock is obtained, or the context is canceled.
// The returned lock will automatically heartbeat to ensure that the lock remains valid. If the heartbeat fails for some reason,
// the channel returned from the `lock.Done()` call will be closed. It is up to the clients to monitor for a lock being lost
func (lc *lockerclient) ObtainLock(ctx context.Context, keyValues datatypes.KeyValues, timeout time.Duration) (Lock, error) {
	select {
	case <-lc.done:
		return nil, fmt.Errorf("locker client has already been canceled and won't process heartbeats. Refusing to obtain the lock")
	default:
		// nothing to do here, can still obtain the lock
	}

	// create lock request body
	createLockRequest := v1locker.CreateLockRequest{
		KeyValues: keyValues,
		Timeout:   timeout,
	}
	body, err := json.Marshal(createLockRequest)
	if err != nil {
		// should never actually hit this
		return nil, err
	}

	obtainedLock := make(chan struct{})
	defer close(obtainedLock)

	var returnLock *lock
	var lockErr error

	// should check to make sure we don't already have the lock
	onCreate := func() any {
		cancelContext, cancel := context.WithCancel(ctx)
		defer cancel()

		go func() {
			select {
			case <-obtainedLock:
				// obtained the lock properly
			case <-cancelContext.Done():
				// caller or canceled the context
			case <-lc.done:
				// locker client is canceled
				cancel()
			}
		}()

		for {
			req, err := http.NewRequestWithContext(cancelContext, "POST", fmt.Sprintf("%s/v1/locker/create", lc.url), bytes.NewBuffer(body))
			if err != nil {
				// this should never actually hit
				lockErr = fmt.Errorf("internal error setting up http request: %w", err)
				return nil
			}

			resp, err := lc.client.Do(req)
			if err != nil {
				// we didn't make the request or was canceled
				select {
				case <-cancelContext.Done():
					select {
					case <-lc.done:
						// locker client was closed
						lockErr = fmt.Errorf("locker client has been canceled and won't process heartbeats. Refusing to obtain the lock: %w", err)
						return nil
					default:
						// client was canceled so don't return an error
						return nil
					}
				default:
					lockErr = fmt.Errorf("unable to make request to locker service: %w", err)
					return nil
				}
			}

			switch resp.StatusCode {
			case http.StatusCreated:
				// created the lock, need to record the session id and start hearbeating
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					// server sent back a bad body for some reason
					lockErr = fmt.Errorf("failed to read response body: %v", err)
					return nil
				}

				createLockResponse := &v1locker.CreateLockResponse{}
				if err = json.Unmarshal(respBody, createLockResponse); err != nil {
					// server sent back an unexpeded response body
					lockErr = fmt.Errorf("failed to parse server response: %v", err)
					return nil
				}

				// wrapper is needed so we ensure that the lock is removed from the tree on removal
				lostLockWrapper := func() {
					canDelete := func(_ any) bool {
						return true
					}
					lc.locks.Delete(btreeassociated.ConverDatatypesKeyValues(keyValues), canDelete)
				}

				if lc.heartbeatErrorCallback != nil {
					returnLock = newLock(createLockResponse.SessionID, createLockRequest.Timeout, lc.client, lc.url, nil, lostLockWrapper)
				} else {
					errCallback := func(err error) {
						lc.heartbeatErrorCallback(keyValues, err)
					}
					returnLock = newLock(createLockResponse.SessionID, createLockRequest.Timeout, lc.client, lc.url, errCallback, lostLockWrapper)
				}

				// setup the heartbeat operation
				if err = lc.asyncManager.AddExecuteTask(createLockResponse.SessionID, returnLock); err != nil {
					// could not add the task to execute, must be shutting down so release the lock
					lockErr = fmt.Errorf("client has been closed, will not obtain lock")
					returnLock.release()

					returnLock = nil
				}

				return returnLock
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

	onFind := func(item any) {
		// nothing to do here
		returnLock = item.(*btreeassociated.AssociatedKeyValues).Value().(*lock)
	}

	// create or find the lock if we already have it
	_, _ = lc.locks.CreateOrFind(btreeassociated.ConverDatatypesKeyValues(keyValues), onCreate, onFind)

	return returnLock, lockErr
}

func (lc *lockerclient) Done() <-chan struct{} {
	return lc.done
}
