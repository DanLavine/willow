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
	Heartbeat(lockID string, lockClaim *v1locker.LockClaim) *errors.ServerError

	// Find all locks currently held in the tree
	LocksQuery(query *v1common.AssociatedQuery) v1locker.Locks

	// Release or delete a specific lock
	Release(lockID string, lockClaim *v1locker.LockClaim) *errors.ServerError
}

type exclusiveLocker struct {
	// association trees for all possible locks
	exclusiveLocks btreeassociated.BTreeAssociated

	// task manger ensures shutdown requests are processsed properly
	taskManager goasync.AsyncTaskManager
}

func NewExclusiveLocker() *exclusiveLocker {
	return &exclusiveLocker{
		exclusiveLocks: btreeassociated.NewThreadSafe(),
		taskManager:    goasync.NewTaskManager(goasync.RelaxedConfig()),
	}
}

func (exclusiveLocker *exclusiveLocker) Initialize() error { return nil }
func (exclusiveLocker *exclusiveLocker) Cleanup() error    { return nil }

// execution for handling shutdown of all the possible locks
func (exclusiveLocker *exclusiveLocker) Execute(ctx context.Context) error {
	// NOTE: this blocks until all locks have processed the shutdown request
	// when the service is told to stop, we want to cancel all the current claim operations
	_ = exclusiveLocker.taskManager.Run(ctx)

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

		var expireTime time.Duration
		lastHeartbeatTime := exclusiveLock.getLastHeartbeatTime()
		if lastHeartbeatTime == (time.Time{}) {
			expireTime = 0
		} else {
			expireTime = time.Since(lastHeartbeatTime)
		}

		locks = append(locks, &v1locker.Lock{
			LockID:             associatedKeyValues.AssociatedID(),
			SessionID:          exclusiveLock.getSessionID(),
			KeyValues:          associatedKeyValues.KeyValues(),
			Timeout:            exclusiveLock.getLockTimeout(),
			TimeTillExipre:     expireTime,
			LocksHeldOrWaiting: exclusiveLock.clientsWaitingForClaim.Load(),
		})

		return true
	})

	return locks
}

func (exclusiveLocker *exclusiveLocker) ObtainLock(clientCtx context.Context, createLockRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse {
	var claimChannel <-chan func(time.Duration) string

	onCreate := func() any {
		lock := newExclusiveLock(func() { exclusiveLocker.timeout(createLockRequest.KeyValues) })

		// add the task to the task manager
		if err := exclusiveLocker.taskManager.AddExecuteTask("", lock); err == nil {
			claimChannel = lock.GetClaimChannel()
		} else {
			// on an error, just set thest to a closed channel to exit from the claim operation. The client will
			// need to retry and the lock should not be saved
			closedChannel := make(chan func(time.Duration) string)
			close(closedChannel)

			claimChannel = closedChannel
		}

		return lock
	}

	onFind := func(item btreeassociated.AssociatedKeyValues) {
		claimChannel = item.Value().(*exclusiveLock).GetClaimChannel()
	}

	lockID, err := exclusiveLocker.exclusiveLocks.CreateOrFind(createLockRequest.KeyValues, onCreate, onFind)
	if err != nil {
		panic(err)
	}

	select {
	// client disconnected
	case <-clientCtx.Done():
		// decrement the claim and remove the object from the tree if there are no other clients waiting

		_ = exclusiveLocker.exclusiveLocks.Delete(createLockRequest.KeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			return associatedKeyValues.Value().(*exclusiveLock).LostClient()
		})

	// claim was obtained or lock stopped processing for shutdown
	case processClaim, ok := <-claimChannel:
		// service was told to shutdown and the client did not obtain the lock
		if !ok {
			return nil
		}

		if ok {
			// at this point, we have successfuly have the exclusive claim so create the lock
			sessionID := processClaim(createLockRequest.LockTimeout)

			return &v1locker.LockCreateResponse{
				LockID:      lockID,
				SessionID:   sessionID,
				LockTimeout: createLockRequest.LockTimeout,
			}
		}
	}

	return nil
}

func (exclusiveLocker *exclusiveLocker) Heartbeat(lockID string, claim *v1locker.LockClaim) *errors.ServerError {
	heartbeaterErr := &errors.ServerError{Message: "LockID could not be found", StatusCode: http.StatusNotFound}

	exclusiveLocker.exclusiveLocks.FindByAssociatedID(lockID, func(associatedKeyValues btreeassociated.AssociatedKeyValues) {
		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)
		heartbeaterErr = exclusiveLock.Heartbeat(claim)
	})

	return heartbeaterErr
}

func (exclusiveLocker *exclusiveLocker) Release(lockID string, claim *v1locker.LockClaim) *errors.ServerError {
	releaseError := &errors.ServerError{Message: "LockID could not be found", StatusCode: http.StatusNotFound}

	exclusiveLocker.exclusiveLocks.DeleteByAssociatedID(lockID, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		var destroy bool
		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)
		destroy, releaseError = exclusiveLock.ReleaseLock(claim)
		return destroy
	})

	return releaseError
}

// callback for the lock when a timeout occurs.
func (exclusiveLocker *exclusiveLocker) timeout(lockKeyValues datatypes.KeyValues) {
	exclusiveLocker.exclusiveLocks.Delete(lockKeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)
		shouldTimeout := exclusiveLock.TimeOut()

		return shouldTimeout
	})
}
