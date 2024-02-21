package lockmanager

import (
	"context"
	"net/http"
	"time"

	"github.com/DanLavine/goasync"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// TODO DSL:
//	I think there is an issue when releasing the lock. It is not acctually releasing the generalLock resource.
//	that needs to also be stopped. The callback will eventually trigger when the timeout hits, but thats not great.
//	should be released immediately

type GeneralLocker interface {
	// obtain all the locks that make up a collection
	ObtainLock(clientCtx context.Context, createRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse

	// Heartbeat any number of locks so we know they are still running properly
	Heartbeat(sessions string) *errors.ServerError

	// Find all locks currently held in the tree
	LocksQuery(query *v1common.AssociatedQuery) v1locker.Locks

	// Release or delete a specific lock
	ReleaseLock(lockID string)
}

type generalLocker struct {
	// closed when the server shuts down
	done chan struct{}

	// association tree for all possible locks
	locks btreeassociated.BTreeAssociated

	// task manger ensures shutdown requests are processsed properly
	taskManager goasync.AsyncTaskManager
}

// nothing to do here for now.
// TODO: read from disk all the locks that already exist and re-instate them
func (generalLocker *generalLocker) Initialize() error { return nil }

// nothing to do here
func (generalLocker *generalLocker) Cleanup() error { return nil }

// execution for handling shutdown of all the possible locks
func (generalLocker *generalLocker) Execute(ctx context.Context) error {
	// NOTE: this blocks until all locks have processed the shutdown request
	go func() {
		defer close(generalLocker.done)
		_ = generalLocker.taskManager.Run(ctx)
	}()

	<-generalLocker.done
	return nil
}

func NewGeneralLocker(tree btreeassociated.BTreeAssociated) *generalLocker {
	if tree == nil {
		tree = btreeassociated.NewThreadSafe()
	}

	return &generalLocker{
		done:        make(chan struct{}),
		locks:       tree,
		taskManager: goasync.NewTaskManager(goasync.RelaxedConfig()),
	}
}

// List all locks held
func (generalLocker *generalLocker) LocksQuery(query *v1common.AssociatedQuery) v1locker.Locks {
	locks := v1locker.Locks{}

	onPaginate := func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		generalLock := associatedKeyValues.Value().(*generalLock)
		generalLock.counterLock.RLock()
		defer generalLock.counterLock.RUnlock()

		locks = append(locks, &v1locker.Lock{
			SessionID:          associatedKeyValues.AssociatedID(),
			KeyValues:          associatedKeyValues.KeyValues(),
			Timeout:            generalLock.timeout,
			TimeTillExipre:     time.Since(generalLock.getLastHeartbeat()),
			LocksHeldOrWaiting: generalLock.counter,
		})

		return true
	}

	_ = generalLocker.locks.Query(query.AssociatedKeyValues, onPaginate)

	return locks
}

// Obtain a lock for the given key values.
// This blocks until one of of the contexts is canceled, or the lock is obtained
func (generalLocker *generalLocker) ObtainLock(clientCtx context.Context, createLockRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse {
	var lockChan chan struct{}
	var lockDone chan struct{}
	var lockUpdateHeartbeatTimeout chan time.Duration

	onCreate := func() any {
		generalLock := newGeneralLock(createLockRequest.LockTimeout, func() bool { return generalLocker.freeLock(createLockRequest.KeyValues) })

		// start running the heartbeat timer in the background
		_ = generalLocker.taskManager.AddExecuteTask("lock timer", generalLock)

		return generalLock
	}

	// NOTE: it is very important that they all procces in the order that they were found
	onFind := func(item btreeassociated.AssociatedKeyValues) {
		generalLock := item.Value().(*generalLock)
		generalLock.counterLock.Lock()
		defer generalLock.counterLock.Unlock()

		generalLock.counter++

		// set the channel to wait for
		lockChan = generalLock.lockChan

		// grab the channels to update the timer
		lockDone = generalLock.done
		lockUpdateHeartbeatTimeout = generalLock.updateHearbeatTimeout
	}

	// lock every single possible tag combination we might be using
	sessionID, err := generalLocker.locks.CreateOrFind(createLockRequest.KeyValues, onCreate, onFind)
	if err != nil {
		panic(err)
	}

	switch lockChan {
	case nil:
		// this is the created case. Can always try and report the sessionID to the client
	default:
		// this is the found case
		select {
		case <-lockChan:
			// a lock has been freed and we were able to claim it
			select {
			case lockUpdateHeartbeatTimeout <- createLockRequest.LockTimeout:
				// updated the heartbeat to the new requests timeout
			case <-lockDone:
				// the lock must have timed out when wantint to update the new timeout duration. Server is running super slow
				// for some reason. I think its probably best to refactor that this just re-adds to the async task manager though
				panic("failed to update the lock in time. This panic is here for now because this need to be refactored!")
			}
		case <-clientCtx.Done(): // client canceled
			// now we need to try and cleanup any locks we currently have recored a counter on
			generalLocker.freeLock(createLockRequest.KeyValues)
			return nil
		case <-generalLocker.done: // server shutdown
			// now we need to try and cleanup any locks we currently have recorded a counter on
			generalLocker.freeLock(createLockRequest.KeyValues)
			return nil
		}
	}

	// at this point, we have obtained the "locks" for all tag groups
	return &v1locker.LockCreateResponse{
		SessionID:   sessionID,
		LockTimeout: createLockRequest.LockTimeout,
	}
}

// heartbeat a particualr session key values
func (generalLocker *generalLocker) Heartbeat(sessionID string) *errors.ServerError {
	found := false
	onFind := func(item btreeassociated.AssociatedKeyValues) {
		generalLock := item.Value().(*generalLock)

		select {
		case generalLock.hertbeat <- struct{}{}:
			// set a heartbeat to the lock
		default:
			// the lock was deleted or we are shutting down
		}

		found = true
	}

	if err := generalLocker.locks.FindByAssociatedID(sessionID, onFind); err != nil {
		panic(err)
	}

	if !found {
		return &errors.ServerError{Message: "SessionID could not be found", StatusCode: http.StatusGone}
	}

	return nil
}

// delete a lock from the tree
func (generalLocker *generalLocker) ReleaseLock(lockID string) {
	var keyValues datatypes.KeyValues

	// only need to find the 1 item
	onQuery := func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		keyValues = associatedKeyValues.KeyValues()
		return false
	}

	idString := datatypes.String(lockID)
	generalLocker.locks.Query(datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{
				btreeassociated.ReservedID: {Value: &idString, ValueComparison: datatypes.EqualsPtr()},
			},
		},
	}, onQuery)

	if len(keyValues) != 0 {
		_ = generalLocker.freeLock(keyValues)
	}
}

func (generalLocker *generalLocker) freeLock(keyValues datatypes.KeyValues) bool {
	removed := false

	canDelete := func(item btreeassociated.AssociatedKeyValues) bool {
		generalLock := item.Value().(*generalLock)

		// don't need to grab the lock here since this is already write protected on the tree
		generalLock.counter--

		if generalLock.counter == 0 {
			// nothing to do here, no clients are waiting
		} else {
			select {
			case generalLock.lockChan <- struct{}{}:
				// this will trigger a client waiting to proceed
			default:
				// if this case is hit. Then the client waiting could have also disconnected, and will be cleaned up
				// on the next call. If delete is next to process as well
			}
		}

		removed = generalLock.counter == 0
		return removed
	}

	// delete or at least signal to other waiting locks that we are freeing the currently held lock
	_ = generalLocker.locks.Delete(keyValues, canDelete)

	return removed
}
