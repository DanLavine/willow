package lockerclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DanLavine/goasync"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

//go:generate mockgen -destination=lockerclientfakes/locker_client_mock.go -package=lockerclientfakes github.com/DanLavine/willow/pkg/clients/locker_client LockerClient
type LockerClient interface {
	// Healthy is used to ensure that the locker service is up and running
	Healthy() error

	//	PARAMS:
	//	- ctx - Context that can be used to cancel the blocking requst trying to obtain the lock. NOTE: once a lock is obtained, release must be called
	//	- lockRequest - request for the lock to obtain with a configured timeout
	//
	//	RETURNS
	//	- Lock - lock object that can be used to release a lock, and monitor if a lock is lost for some reason
	//	- error - any errors encountered when obtaining the lock
	//	NOTE: is both Lock and error are nil, the context must have been canceled obtaining the lock
	//
	// Obtain a lock for a particular set of KeyValues. This blocks until the desired lock is obtained, or the context is canceled.
	// The returned lock will automatically heartbeat to ensure that the lock remains valid. If the heartbeat fails for some reason,
	// the channel returned from the `lock.Done()` call will be closed. It is up to the clients to monitor for a lock being lost
	ObtainLock(ctx context.Context, lockRequest *v1locker.CreateLockRequest) (Lock, error)

	// Done is closed if the LockerClient's contex is closed and no longer heartbeating
	Done() <-chan struct{}
}

type lockerclient struct {
	// used to know if the async manager is done
	done chan struct{}

	// client to connect with the remote Locker service
	url         string
	client      clients.HttpClient
	contentType api.ContentType

	// callback to client if a lock was somehow released on the server beccause heartbeats are failing
	heartbeatErrorCallback func(keyValues datatypes.KeyValues, err error)

	// each item in the locks tree's value is a lock
	locks btreeassociated.BTreeAssociated

	// async manager to handle heartbeats
	asyncManager goasync.AsyncTaskManager
}

//	PARAMS
//	- ctx - Context for the http(s) client. This must be closed to close the LockerCLient and will trigger a close of all held locks
//	- cfg - configuration for the http client
//	- heartbeatErrorCallback (optional) - callback for heartbeat errors. Mainly used to log any errors the managed client to the locker service might be experiencing
//
//	RETURNS:
//	- LockerClient - properly configured locker client that manages all held locks
//	- error - any errors setting up the client
//
// Setup a new client to the remote locker service. This client automatically manages heartbeats for any obtained locks and
// will notify the user if a lock is lost at some point.
func NewLockerClient(ctx context.Context, cfg *clients.Config, heartbeatErrorCallback func(keyValue datatypes.KeyValues, err error)) (LockerClient, error) {
	if ctx == nil || ctx == context.TODO() || ctx == context.Background() {
		return nil, fmt.Errorf("cannot use provided context. The context must be canceled by the caller to cleanup async resource management")
	}

	httpClient, err := clients.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	done := make(chan struct{})
	asyncManager := goasync.NewTaskManager(goasync.RelaxedConfig())

	lockerClient := &lockerclient{
		done:                   done,
		url:                    cfg.URL,
		client:                 httpClient,
		contentType:            cfg.ContentType,
		heartbeatErrorCallback: heartbeatErrorCallback,
		locks:                  btreeassociated.NewThreadSafe(),
		asyncManager:           asyncManager,
	}

	go func() {
		defer close(done)
		_ = asyncManager.Run(ctx)
	}()

	return lockerClient, nil
}

// Healthy is used to ensure that the /health endpoint on the Locker service can be reached
//
//	RETURNS:
//	- error - error if the Locker service cannot be reached
func (lc *lockerclient) Healthy() error {
	// setup and make the request
	resp, err := lc.client.Do(&clients.RequestData{
		Method: "GET",
		Path:   fmt.Sprintf("%s/health", lc.url),
		Model:  nil,
	})

	if err != nil {
		return err
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
//	- lockRequest - request for the lock to obtain
//
//	RETURNS
//	- Lock - lock object that can be used to release a lock, and monitor if a lock is lost for some reason
//	- error - any errors encountered when obtaining the lock
//	NOTE: if both Lock and error are nil, the context must have been canceled obtaining the lock
//
// Obtain a lock for a particular set of KeyValues. This blocks until the desired lock is obtained, or the context is canceled.
// The returned lock will automatically heartbeat to ensure that the lock remains valid. If the heartbeat fails for some reason,
// the channel returned from the `lock.Done()` call will be closed. It is up to the clients to monitor for a lock being lost
func (lc *lockerclient) ObtainLock(ctx context.Context, lockRequest *v1locker.CreateLockRequest) (Lock, error) {
	select {
	case <-lc.done:
		return nil, fmt.Errorf("locker client has already been canceled and won't process heartbeats. Refusing to obtain the lock")
	default:
		// nothing to do here, can still obtain the lock
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
			// setup and make request
			resp, err := lc.client.DoWithContext(cancelContext, &clients.RequestData{
				Method: "POST",
				Path:   fmt.Sprintf("%s/v1/locks", lc.url),
				Model:  lockRequest,
			})

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
				createLockResponse := &v1locker.CreateLockResponse{}
				if err := createLockResponse.Decode(api.ContentTypeFromResponse(resp), resp.Body); err != nil {
					lockErr = errors.ClientError(err)
					return nil
				}

				// wrapper is needed so we ensure that the lock is removed from the tree on removal
				lostLockWrapper := func() {
					canDelete := func(_ any) bool {
						return true
					}
					lc.locks.Delete(btreeassociated.ConverDatatypesKeyValues(lockRequest.KeyValues), canDelete)
				}

				if lc.heartbeatErrorCallback != nil {
					errCallback := func(err error) {
						lc.heartbeatErrorCallback(lockRequest.KeyValues, err)
					}
					returnLock = newLock(createLockResponse.SessionID, createLockResponse.Timeout, lc.url, lc.client, lc.contentType, errCallback, lostLockWrapper)
				} else {
					returnLock = newLock(createLockResponse.SessionID, createLockResponse.Timeout, lc.url, lc.client, lc.contentType, nil, lostLockWrapper)
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
				apiError := &errors.Error{}
				if err := apiError.Decode(api.ContentTypeFromResponse(resp), resp.Body); err != nil {
					lockErr = errors.ClientError(err)
				} else {
					lockErr = apiError
				}

				return nil
			case http.StatusServiceUnavailable:
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
	_, _ = lc.locks.CreateOrFind(btreeassociated.ConverDatatypesKeyValues(lockRequest.KeyValues), onCreate, onFind)

	return returnLock, lockErr
}

//	RETURNS:
//	- <-chan struct{} - struc that can be used to monitor when a client has been closed
//
// Done is closed when the LockerClient's context is closed and all held locks have successfully been released
func (lc *lockerclient) Done() <-chan struct{} {
	return lc.done
}
