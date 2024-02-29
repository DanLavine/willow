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

type ExcluiveLocker interface {
	// obtain all the locks that make up a collection
	ObtainLock(clientCtx context.Context, createRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse

	// Heartbeat any number of locks so we know they are still running properly
	Heartbeat(sessions string) *errors.ServerError

	// Find all locks currently held in the tree
	LocksQuery(query *v1common.AssociatedQuery) v1locker.Locks

	// Release or delete a specific lock
	Release(lockID string)
}

type exclusiveLocker struct {
	// closed when the server shuts down
	shutdown context.Context
	cancel   func()

	// association trees for all possible locks
	exclusiveLocks  btreeassociated.BTreeAssociated
	exclusiveClaims btreeassociated.BTreeAssociated

	// task manger ensures shutdown requests are processsed properly
	taskManager goasync.AsyncTaskManager
}

func NewExclusiveLocker() *exclusiveLocker {
	shutdown, cancel := context.WithCancel(context.Background())

	return &exclusiveLocker{
		shutdown:        shutdown,
		cancel:          cancel,
		exclusiveLocks:  btreeassociated.NewThreadSafe(),
		exclusiveClaims: btreeassociated.NewThreadSafe(),
		taskManager:     goasync.NewTaskManager(goasync.RelaxedConfig()),
	}
}

func (exclusiveLocker *exclusiveLocker) Initialize() error { return nil }
func (exclusiveLocker *exclusiveLocker) Cleanup() error    { return nil }

// execution for handling shutdown of all the possible locks
func (exclusiveLocker *exclusiveLocker) Execute(ctx context.Context) error {
	// NOTE: this blocks until all locks have processed the shutdown request
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = exclusiveLocker.taskManager.Run(ctx)
	}()

	// when the service is told to stop, we want to cancel all the current claim operations
	<-ctx.Done()
	exclusiveLocker.cancel()

	// wait until all the heartbeater operations have finished processing
	<-done
	return nil
}

// Helper to find all the currently held locks and the number of clients also wanting to obtain the locks.
//
// NOTE: there is a slight race condition where a lock is release and before it is claimed again, this will
// not find that lock. However to do that I believe I would need to change the API soaving rather than having the
// '/v1/locks/:lock_id' apis, they would also require some sort of "session id" to identify themselves as actually
// having the lock. Eventually this would just be an "ADMIN" api, so I don't see it being used for real in the system
// as we don't actuallyl want to return LockIDs to clients. Those should be hidden for just the clients that obtained
// the original lock.
func (exclusiveLocker *exclusiveLocker) LocksQuery(query *v1common.AssociatedQuery) v1locker.Locks {
	locks := v1locker.Locks{}

	exclusiveLocker.exclusiveLocks.Query(query.AssociatedKeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)

		locks = append(locks, &v1locker.Lock{
			SessionID:          associatedKeyValues.AssociatedID(),
			KeyValues:          associatedKeyValues.KeyValues(),
			Timeout:            exclusiveLock.lockTimeout,
			TimeTillExipre:     time.Since(exclusiveLock.getLastHeartbeatTime()),
			LocksHeldOrWaiting: exclusiveLock.clientsWaitingForClaim.Load(),
		})

		return true
	})

	return locks
}

func (exclusiveLocker *exclusiveLocker) ObtainLock(clientCtx context.Context, createLockRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse {
	var claim *exclusiveClaim

	created := false
	onCreate := func() any {
		created = true

		claim = newExclusiveClaim()
		claim.addClientWaiting()

		return claim
	}

	onFind := func(item btreeassociated.AssociatedKeyValues) {
		claim = item.Value().(*exclusiveClaim)
		claim.addClientWaiting()
	}

	claimID, err := exclusiveLocker.exclusiveClaims.CreateOrFind(createLockRequest.KeyValues, onCreate, onFind)
	if err != nil {
		panic(err)
	}

	// if this was create, create the task with the claim id to the task manager
	if created {
		_ = exclusiveLocker.taskManager.AddExecuteTask(claimID, claim)
	}

	select {
	// service was told to shutdown
	case <-exclusiveLocker.shutdown.Done():
		return nil

	// client disconnected
	case <-clientCtx.Done():
		// decrement the claim and remove the object from the tree if there are no other clients waiting
		_ = exclusiveLocker.exclusiveClaims.Delete(createLockRequest.KeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			return associatedKeyValues.Value().(*exclusiveClaim).removeClientWaiting(false) == 0
		})

	// claim was obtained
	case _, ok := <-claim.claim:
		// if the claim channel was closed. then it could be a race with the serivce shutdown
		if !ok {
			return nil
		}

		// at this point, we have successfuly have the exclusive claim so create the lock
		var exclusiveLock *exclusiveLock
		lockID, err := exclusiveLocker.exclusiveLocks.Create(createLockRequest.KeyValues, func() any {
			exclusiveLock = newExclusiveLock(claim.clientsWaitingForClaim, createLockRequest.LockTimeout, func() bool { return exclusiveLocker.timeout(createLockRequest.KeyValues) })
			return exclusiveLock
		})

		if err != nil {
			panic(err)
		}

		// start processing the lock in the background
		_ = exclusiveLocker.taskManager.AddExecuteTask(lockID, exclusiveLock)

		return &v1locker.LockCreateResponse{
			SessionID:   lockID,
			LockTimeout: createLockRequest.LockTimeout,
		}
	}

	return nil
}

func (exclusiveLocker *exclusiveLocker) Heartbeat(lockID string) *errors.ServerError {
	heartbeaterErr := &errors.ServerError{Message: "SessionID could not be found", StatusCode: http.StatusGone}

	exclusiveLocker.exclusiveLocks.FindByAssociatedID(lockID, func(associatedKeyValues btreeassociated.AssociatedKeyValues) {
		_, ok := <-associatedKeyValues.Value().(*exclusiveLock).heartbeat

		// heartbeat was processed
		if ok {
			heartbeaterErr = nil
		}
	})

	return heartbeaterErr
}

func (exclusiveLocker *exclusiveLocker) Release(lockID string) {
	var keyValues datatypes.KeyValues

	exclusiveLocker.exclusiveLocks.DeleteByAssociatedID(lockID, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)

		_, ok := <-exclusiveLock.release

		// lock was released
		if ok {
			keyValues = associatedKeyValues.KeyValues()
			return true
		}

		// service shutdown. client will need to retry the request
		return false
	})

	// if this is true, we released the lock. now need to update the claim
	if len(keyValues) > 0 {
		_ = exclusiveLocker.exclusiveClaims.Delete(keyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			return associatedKeyValues.Value().(*exclusiveClaim).removeClientWaiting(true) == 0
		})
	}
}

// callback for the lock when a timeout occurs.
func (exclusiveLocker *exclusiveLocker) timeout(lockKeyValues datatypes.KeyValues) bool {
	lockRemoved := false

	// lock was removed
	exclusiveLocker.exclusiveLocks.Delete(lockKeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		lockRemoved = true
		return true
	})

	// if this is true, we released the lock. now need to update the claim
	if lockRemoved {
		_ = exclusiveLocker.exclusiveClaims.Delete(lockKeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			exclusiveClaim := associatedKeyValues.Value().(*exclusiveClaim)
			return exclusiveClaim.removeClientWaiting(true) == 0
		})
	}

	return lockRemoved
}
