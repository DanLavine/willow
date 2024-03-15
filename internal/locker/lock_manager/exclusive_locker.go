package lockmanager

import (
	"context"
	"net/http"
	"time"

	"github.com/DanLavine/goasync"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

type ExcluiveLocker interface {
	// obtain all the locks that make up a collection
	ObtainLock(ctx context.Context, createRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse

	// Heartbeat any number of locks so we know they are still running properly
	Heartbeat(ctx context.Context, lockID string, lockClaim *v1locker.LockClaim) *errors.ServerError

	// Find all locks currently held in the tree
	LocksQuery(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) v1locker.Locks

	// Release or delete a specific lock
	Release(ctx context.Context, lockID string, lockClaim *v1locker.LockClaim) *errors.ServerError
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
func (exclusiveLocker *exclusiveLocker) LocksQuery(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) v1locker.Locks {
	logger := reporting.GetLogger(ctx).Named("LocksQuery")
	logger.Debug("querying locks")
	defer logger.Debug("done querying lock")

	locks := v1locker.Locks{}

	exclusiveLocker.exclusiveLocks.QueryAction(query, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
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

func (exclusiveLocker *exclusiveLocker) ObtainLock(ctx context.Context, createLockRequest *v1locker.LockCreateRequest) *v1locker.LockCreateResponse {
	logger := reporting.GetLogger(ctx).Named("ObtainLock")
	logger.Debug("obtaining lock")
	defer logger.Debug("done obtaining lock")

	var claimChannel <-chan func(time.Duration) string

	onCreate := func() any {
		lock := newExclusiveLock(func() { exclusiveLocker.timeout(reporting.BaseLogger(logger), createLockRequest.KeyValues) })

		// add the task to the task manager
		if err := exclusiveLocker.taskManager.AddExecuteTask("", lock); err == nil {
			claimChannel = lock.GetClaimChannel()
		} else {
			// on an error, just set these to a closed channel to exit from the claim operation. The client will
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
	case <-ctx.Done():
		// decrement the claim and remove the object from the tree if there are no other clients waiting
		logger.Info("client disconnected")
		_ = exclusiveLocker.exclusiveLocks.Delete(createLockRequest.KeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
			return associatedKeyValues.Value().(*exclusiveLock).LostClient()
		})

	// claim was obtained or lock stopped processing for shutdown
	case processClaim, ok := <-claimChannel:
		// service was told to shutdown and the client did not obtain the lock
		if !ok {
			logger.Info("server is shutting down. Client will need to retry obtaining the lock")
			return nil
		}

		if ok {
			// at this point, we have successfuly have the exclusive claim so create the lock
			logger.Info("claimed the lock")
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

func (exclusiveLocker *exclusiveLocker) Heartbeat(ctx context.Context, lockID string, claim *v1locker.LockClaim) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("Heartbeat")
	logger.Debug("attempting to heartbeat")

	heartbeaterErr := &errors.ServerError{Message: "LockID could not be found", StatusCode: http.StatusNotFound}

	idQuery := &queryassociatedaction.AssociatedActionQuery{
		Selection: &queryassociatedaction.Selection{
			IDs: []string{lockID},
		},
	}

	_ = exclusiveLocker.exclusiveLocks.QueryAction(idQuery, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)
		heartbeaterErr = exclusiveLock.Heartbeat(claim)
		return false
	})

	if heartbeaterErr == nil {
		logger.Debug("successfully heartbeat")
	} else {
		logger.Debug("failed to heartbeat", zap.Error(heartbeaterErr))
	}

	return heartbeaterErr
}

func (exclusiveLocker *exclusiveLocker) Release(ctx context.Context, lockID string, claim *v1locker.LockClaim) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("Release")
	logger.Debug("releasing the lock")

	releaseError := &errors.ServerError{Message: "LockID could not be found", StatusCode: http.StatusNotFound}

	exclusiveLocker.exclusiveLocks.DeleteByAssociatedID(lockID, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		var destroy bool
		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)
		destroy, releaseError = exclusiveLock.ReleaseLock(claim)
		return destroy
	})

	if releaseError == nil {
		logger.Debug("released the lock")
	} else {
		logger.Debug("failed to release the lock", zap.Error(releaseError))
	}

	return releaseError
}

// callback for the lock when a timeout occurs.
func (exclusiveLocker *exclusiveLocker) timeout(logger *zap.Logger, lockKeyValues datatypes.KeyValues) {
	exclusiveLocker.exclusiveLocks.Delete(lockKeyValues, func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		logger.Warn("lock timed out", zap.Any("key_values", lockKeyValues))

		exclusiveLock := associatedKeyValues.Value().(*exclusiveLock)
		shouldTimeout := exclusiveLock.TimeOut()

		return shouldTimeout
	})
}
