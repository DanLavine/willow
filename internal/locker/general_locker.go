package locker

import (
	"context"
	"fmt"
	"time"

	"github.com/DanLavine/goasync"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

type GeneralLocker interface {
	// obtain all the locks that make up a collection
	ObtainLock(clientCtx context.Context, createRequest *v1locker.CreateLockRequest) *v1locker.CreateLockResponse

	// Heartbeat any number of locks so we know they are still running properly
	Heartbeat(sessions []string) []v1locker.HeartbeatError

	// Find all locks currently held in the tree
	ListLocks() []v1locker.Lock

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
func (generalLocker *generalLocker) ListLocks() []v1locker.Lock {
	var locks []v1locker.Lock

	onPaginate := func(item any) bool {
		associatedKeyValues := item.(*btreeassociated.AssociatedKeyValues)

		generalLock := associatedKeyValues.Value().(*generalLock)
		generalLock.counterLock.RLock()
		defer generalLock.counterLock.RUnlock()

		locks = append(locks, v1locker.Lock{KeyValues: associatedKeyValues.KeyValues().StripKey(btreeassociated.ReservedID), LocksHeldOrWaiting: generalLock.counter})

		return true
	}

	_ = generalLocker.locks.Query(query.AssociatedKeyValuesQuery{}, onPaginate)

	return locks
}

// Obtain a lock for the given key values.
// This blocks until one of of the contexts is canceled, or the lock is obtained
func (generalLocker *generalLocker) ObtainLock(clientCtx context.Context, createLockRequest *v1locker.CreateLockRequest) *v1locker.CreateLockResponse {
	var sessionID string
	var timeout time.Duration
	var lockChan chan struct{}

	onCreate := func() any {
		generalLock := newGeneralLock(createLockRequest.Timeout, func() bool { return generalLocker.freeLock(createLockRequest.KeyValues) })
		timeout = createLockRequest.Timeout

		// start running the heartbeat timer in the background
		_ = generalLocker.taskManager.AddExecuteTask("lock timer", generalLock)

		return generalLock
	}

	// NOTE: it is very important that the all procces in the order that they were found
	onFind := func(item any) {
		generalLock := item.(*btreeassociated.AssociatedKeyValues).Value().(*generalLock)
		generalLock.counterLock.Lock()
		defer generalLock.counterLock.Unlock()

		generalLock.counter++
		timeout = generalLock.timeout

		// set the channel to wait for
		lockChan = generalLock.lockChan
	}

	// lock every single possible tag combination we might be using
	sessionID, _ = generalLocker.locks.CreateOrFind(createLockRequest.KeyValues, onCreate, onFind)

	switch lockChan {
	case nil:
		// this is the created case. Can alwys try and report the sessionID to the client
	default:
		// this is the found case
		select {
		case <-lockChan:
			// a lock has been freed and we were able to claim it
		case <-clientCtx.Done():
			// now we need to try and cleanup any locks we currently have recored a counter on
			generalLocker.freeLock(createLockRequest.KeyValues)
			return nil
		case <-generalLocker.done:
			// now we need to try and cleanup any locks we currently have recored a counter on
			generalLocker.freeLock(createLockRequest.KeyValues)
			return nil
		}
	}

	// at this point, we have obtained the "locks" for all tag groups
	return &v1locker.CreateLockResponse{
		SessionID: sessionID,
		Timeout:   timeout,
	}
}

// heartbeat any number of key values
func (generalLocker *generalLocker) Heartbeat(sessions []string) []v1locker.HeartbeatError {
	var heartbeatErrors []v1locker.HeartbeatError

	found := false
	onFind := func(item any) {
		generalLock := item.(*btreeassociated.AssociatedKeyValues).Value().(*generalLock)

		select {
		case generalLock.hertbeat <- struct{}{}:
			// set a heartbeat to the lock
		default:
			// the lock was deleted or we are shutting down
		}

		found = true
	}

	for _, session := range sessions {
		_ = generalLocker.locks.FindByAssociatedID(session, onFind)

		if !found {
			heartbeatErrors = append(heartbeatErrors, v1locker.HeartbeatError{Session: session, Error: "session could not be found"})
		}

		found = false
	}

	return heartbeatErrors
}

// delete a lock from the
func (generalLocker *generalLocker) ReleaseLock(lockID string) {
	var keyValues datatypes.StringMap

	// only need to find the 1 item
	onQuery := func(item any) bool {
		keyValues = item.(*btreeassociated.AssociatedKeyValues).KeyValues().StripKey(query.ReservedID)
		return false
	}

	idString := datatypes.String(lockID)
	generalLocker.locks.Query(query.AssociatedKeyValuesQuery{
		KeyValueSelection: &query.KeyValueSelection{
			KeyValues: map[string]query.Value{
				query.ReservedID: {Value: &idString, ValueComparison: query.EqualsPtr()},
			},
		},
	}, onQuery)

	if len(keyValues) != 0 {
		_ = generalLocker.freeLock(keyValues)
	}
}

func (generalLocker *generalLocker) freeLock(keyValues datatypes.StringMap) bool {
	removed := false

	fmt.Println("Calling free lock")

	canDelete := func(item any) bool {
		generalLock := item.(*btreeassociated.AssociatedKeyValues).Value().(*generalLock)

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
