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
	shutdown context.Context
	cancel   func()

	// association tree for all possible locks
	locks      btreeassociated.BTreeAssociated
	lockTimers btreeassociated.BTreeAssociated

	// task manger ensures shutdown requests are processsed properly
	taskManager goasync.AsyncTaskManager
}

// nothing to do here for now.
func (generalLocker *generalLocker) Initialize() error { return nil }

// nothing to do here
func (generalLocker *generalLocker) Cleanup() error { return nil }

// execution for handling shutdown of all the possible locks
func (generalLocker *generalLocker) Execute(ctx context.Context) error {
	// NOTE: this blocks until all locks have processed the shutdown request
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = generalLocker.taskManager.Run(ctx)
	}()

	// when the service is told to stop, we want to cancel all the current heartberter operations
	<-ctx.Done()
	generalLocker.cancel()

	// wait until all the heartbeater operations have finished processing
	<-done
	return nil
}

func NewGeneralLocker(lockTimers btreeassociated.BTreeAssociated) *generalLocker {
	if lockTimers == nil {
		lockTimers = btreeassociated.NewThreadSafe()
	}

	locks := btreeassociated.NewThreadSafe()

	shutdown, cancel := context.WithCancel(context.Background())

	return &generalLocker{
		shutdown:    shutdown,
		cancel:      cancel,
		locks:       locks,
		lockTimers:  lockTimers,
		taskManager: goasync.NewTaskManager(goasync.RelaxedConfig()),
	}
}

// List all locks held
func (generalLocker *generalLocker) LocksQuery(query *v1common.AssociatedQuery) v1locker.Locks {
	locks := v1locker.Locks{}

	onPaginate := func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		generalTimer := associatedKeyValues.Value().(*generalTimer)

		locks = append(locks, &v1locker.Lock{
			SessionID:          generalTimer.sessionID,
			KeyValues:          associatedKeyValues.KeyValues(),
			Timeout:            time.Duration(generalTimer.timeout.Load()),
			TimeTillExipre:     time.Since(generalTimer.GetLastHeartbeat()),
			LocksHeldOrWaiting: generalTimer.clientsWaiting.Load(),
		})

		return true
	}

	_ = generalLocker.lockTimers.Query(query.AssociatedKeyValues, onPaginate)

	return locks
}

// Obtain a lock for the given key values.
// This blocks until one of of the contexts is canceled, or the lock is obtained
func (generalLocker *generalLocker) ObtainLock(clientCtx context.Context, createLockRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse {
	created := false
	var lockTimer *generalTimer

	onCreate := func() any {
		created = true
		lockTimer = newGeneralLock(func() bool { return generalLocker.freeLock(createLockRequest.KeyValues) })
		return lockTimer
	}

	onFind := func(item btreeassociated.AssociatedKeyValues) {
		lockTimer = item.Value().(*generalTimer)
		lockTimer.AddWaitingClient()
	}

	timerID, err := generalLocker.lockTimers.CreateOrFind(createLockRequest.KeyValues, onCreate, onFind)
	if err != nil {
		panic(err)
	}

	if created {
		_ = generalLocker.taskManager.AddExecuteTask(timerID, lockTimer)
	}

	// this is the case when we found a timer that already exists
	select {
	// obtained the lock timer.
	case lockTimer.claimTrigger <- createLockRequest.LockTimeout:
		// this has a bug and can panic. if there are many clients and shutdown is called.
		// this can panic on the send. so this channel should never be closed?

		lockID, err := generalLocker.locks.Create(createLockRequest.KeyValues, func() any {
			return lockTimer.GetHeartbeater()
		})

		if err != nil {
			panic(err)
		}

		lockTimer.sessionID = lockID
		return &v1locker.LockCreateResponse{
			SessionID:   lockID,
			LockTimeout: createLockRequest.LockTimeout,
		}

	// client has disconnected
	case <-clientCtx.Done():
		// want to decrment the counters for timer and possible removal
		_ = generalLocker.lockTimers.Delete(createLockRequest.KeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			generalTimer := associatedKeyValues.Value().(*generalTimer)

			// remove the lock timer if there are no other clients waiting
			return generalTimer.RemoveWaitingClient() <= 0
		})

	// service told to shutdown
	case <-generalLocker.shutdown.Done():
		// if we created the timer, can also creat the lock on the shutdown request
		if created {
			lockID, err := generalLocker.locks.Create(createLockRequest.KeyValues, func() any {
				return lockTimer.GetHeartbeater()
			})

			if err != nil {
				panic(err)
			}

			lockTimer.sessionID = lockID
			return &v1locker.LockCreateResponse{
				SessionID:   lockID,
				LockTimeout: createLockRequest.LockTimeout,
			}
		}
	}

	return nil
}

// heartbeat a particualr session id
func (generalLocker *generalLocker) Heartbeat(sessionID string) *errors.ServerError {
	heartbeaterErr := &errors.ServerError{Message: "SessionID could not be found", StatusCode: http.StatusGone}

	_ = generalLocker.locks.FindByAssociatedID(sessionID, func(item btreeassociated.AssociatedKeyValues) {
		heartbeat := item.Value().(chan struct{})

		_, ok := <-heartbeat
		if ok {
			heartbeaterErr = nil
		}
	})

	return heartbeaterErr
}

// release a lock from the tree. if there are no more clients waitiing, then it will be deleted
func (generalLocker *generalLocker) ReleaseLock(sessionID string) {
	// remove the unique heartbeat channel
	_ = generalLocker.locks.DeleteByAssociatedID(sessionID, func(item btreeassociated.AssociatedKeyValues) bool {

		// attempt to remove the time trigger if there are no other clients
		_ = generalLocker.lockTimers.Delete(item.KeyValues(), func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			generalTimer := associatedKeyValues.Value().(*generalTimer)
			generalTimer.sessionID = ""

			// wait for the timer to process the release
			<-generalTimer.release

			return generalTimer.RemoveWaitingClient() <= 0
		})

		// always delete the saved locks heartbeater channel
		return true
	})
}

func (generalLocker *generalLocker) freeLock(keyValues datatypes.KeyValues) bool {
	processed := false

	// remove the unique heartbeat channel
	_ = generalLocker.locks.Delete(keyValues, func(item btreeassociated.AssociatedKeyValues) bool {
		processed = true

		// attempt to remove the time trigger if there are no other clients
		_ = generalLocker.lockTimers.Delete(keyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			generalTimer := associatedKeyValues.Value().(*generalTimer)
			generalTimer.sessionID = ""

			return generalTimer.RemoveWaitingClient() <= 0
		})

		// always delete the saved locks heartbeater channel
		return true
	})

	return processed
}
