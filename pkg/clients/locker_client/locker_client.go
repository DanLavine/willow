package lockerclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/clients"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// LockerClient interface defines a the methods for a *LockClient
//
//go:generate mockgen -destination=lockerclientfakes/locker_client_mock.go -package=lockerclientfakes github.com/DanLavine/willow/pkg/clients/locker_client LockerClient
type LockerClient interface {
	//	RETURNS:
	//	- error - error if the Locker service cannot be reached
	//
	// Healthy is used to ensure that the Locker service can be reached
	Healthy() error

	//	PARAMS:
	//	- ctx - Context that can be used to cancel the blocking requst trying to obtain the lock. NOTE: once a lock is obtained, release must be called
	//	- lockRequest - request for the lock to obtain
	//	- headers (optional) - optional http headers to add to the http request
	//	- heartbeatErrorCallback (optional) - callback for heartbeat errors. Mainly used to log any errors the managed client to the locker service might be experiencing
	//
	//	RETURNS
	//	- Lock - lock object that can be used to release a lock, and monitor if a lock is lost for some reason
	//	- error - any errors encountered when obtaining the lock
	//	NOTE: if both Lock and error are nil, the context must have been canceled obtaining the lock
	//
	// Obtain a lock for a particular set of KeyValues. This blocks until the desired lock is obtained, or the context is canceled.
	// The returned lock will automatically heartbeat to ensure that the lock remains valid. If the heartbeat fails for some reason,
	// the channel returned from the `lock.Done()` call will be closed. It is up to the clients to monitor for a lock being lost
	ObtainLock(ctx context.Context, lockRequest *v1locker.Lock, heartbeatErrorCallback func(keyValue datatypes.KeyValues, err error)) (Lock, error)
}

// LockClient interacts with the Locker service.
//
// One useful strategy for claiming multiple locks is to always sort the KeyValues for the obtained locks by.
// As long as these rules are followed by each of the services, then there will be no deadlocks
//  1. Sort lenght of KeyValues, with min first
//  2. Sort each of the KeyValues by their Keys to know which locks to obtaini first
type LockClient struct {
	// client to connect with the remote Locker service
	url    string
	client *http.Client

	// each item in the locks tree's value is a lock
	locks btreeassociated.BTreeAssociated
}

//	PARAMS
//	- cfg - configuration for the HTTP(s) client
//
//	RETURNS:
//	- LockClient - properly configured locker client that manages all held locks
//	- error - any errors setting up the client
//
// Setup a new client to the remote locker service. This client automatically manages heartbeats for any obtained locks and
// will notify the user if a lock is lost at some point.
func NewLockClient(cfg *clients.Config) (*LockClient, error) {
	httpClient, err := clients.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	lockerClient := &LockClient{
		url:    cfg.URL,
		client: httpClient,
		locks:  btreeassociated.NewThreadSafe(),
	}

	return lockerClient, nil
}

//	RETURNS:
//	- error - error if the Locker service cannot be reached
//
// Healthy is used to ensure that the Locker service can be reached
func (lc *LockClient) Healthy() error {
	// setup and make the health request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/health", lc.url), nil)
	if err != nil {
		return fmt.Errorf("failed to setup request to healthy api")
	}

	resp, err := lc.client.Do(req)
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
//	- headers (optional) - optional http headers to add to the http request
//	- heartbeatErrorCallback (optional) - callback for heartbeat errors. Mainly used to log any errors the managed client to the locker service might be experiencing
//
//	RETURNS
//	- Lock - lock object that can be used to release a lock, and monitor if a lock is lost for some reason
//	- error - any errors encountered when obtaining the lock
//	NOTE: if both Lock and error are nil, the context must have been canceled obtaining the lock
//
// Obtain a lock for a particular set of KeyValues. This blocks until the desired lock is obtained, or the context is canceled.
// The returned lock will automatically heartbeat to ensure that the lock remains valid. If the heartbeat fails for some reason,
// the channel returned from the `lock.Done()` call will be closed. It is up to the clients to monitor for a lock being lost
func (lc *LockClient) ObtainLock(ctx context.Context, lockRequest *v1locker.Lock, heartbeatErrorCallback func(keyValue datatypes.KeyValues, err error)) (Lock, error) {
	var returnLock Lock
	var lockErr error

	// obtain the desired lock
	onCreate := func() any {
		for {
			// encode and validate the model
			data, err := api.ObjectEncodeRequest(lockRequest)
			if err != nil {
				lockErr = err
				return nil
			}

			// setup the http request
			obtainLockReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/locks", lc.url), bytes.NewBuffer(data))
			if err != nil {
				lockErr = err
				return nil
			}
			clients.AddHeadersFromContext(obtainLockReq, ctx)
			obtainLockReq.Header.Set("Content-Type", "application/json")

			// Obtain the lock
			resp, err := lc.client.Do(obtainLockReq)
			if err != nil {
				select {
				case <-ctx.Done():
					// request was canceled
				default:
					// failed to obtain the lock
					lockErr = fmt.Errorf("unable to make request to locker service: %w", err)
				}

				// don't save anything for create since there was an error obtaining the lock
				return nil
			}

			switch resp.StatusCode {
			case http.StatusOK:
				// created the lock, need to record the session id and start hearbeating
				createLockResponse := &v1locker.Lock{}
				if err := api.ModelDecodeResponse(resp, createLockResponse); err != nil {
					lockErr = err
					return nil
				}

				// wrapper is needed so we ensure that the lock is removed from the tree on removal
				lostLockWrapper := func() {
					canDelete := func(_ btreeassociated.AssociatedKeyValues) bool {
						return true
					}

					if err = lc.locks.Delete(lockRequest.Spec.DBDefinition.KeyValues, canDelete); err != nil {
						panic(fmt.Sprintf("failed to relase the lock's memeory footprint: %s", err.Error()))
					}
				}

				if heartbeatErrorCallback != nil {
					errCallback := func(err error) {
						heartbeatErrorCallback(lockRequest.Spec.DBDefinition.KeyValues, err)
					}
					returnLock = newLock(createLockResponse, lc.url, lc.client, errCallback, lostLockWrapper)
				} else {
					returnLock = newLock(createLockResponse, lc.url, lc.client, nil, lostLockWrapper)
				}

				return returnLock
			case http.StatusBadRequest, http.StatusInternalServerError:
				// there was an error with the request. possibly a mismatch on client server versions
				apiError := &errors.Error{}
				if err := api.ModelDecodeResponse(resp, apiError); err != nil {
					lockErr = err
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

	onFind := func(item btreeassociated.AssociatedKeyValues) {
		// just return the instance of the found lock
		returnLock = item.Value().(Lock)
	}

	// create or find the lock if we already have it
	_, _ = lc.locks.CreateOrFind(lockRequest.Spec.DBDefinition.KeyValues, onCreate, onFind)

	return returnLock, lockErr
}
