package v1server

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/locker"
	"github.com/DanLavine/willow/internal/logger"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	"go.uber.org/zap"
)

// Handles CRUD operations for Limit operations
//
//go:generate mockgen -destination=v1serverfakes/locker_mock.go -package=v1serverfakes github.com/DanLavine/willow/internal/server/versions/v1server LockerHandler
type LockerHandler interface {
	// create a group rule
	Create(w http.ResponseWriter, r *http.Request)

	// Heartbeat is used to ensure that clients are still active and have the obtained locks
	Heartbeat(w http.ResponseWriter, r *http.Request)

	// read group rules
	List(w http.ResponseWriter, r *http.Request)

	// delete a group rule
	Delete(w http.ResponseWriter, r *http.Request)
}

type lockerHandler struct {
	logger *zap.Logger
	cfg    *config.LockerConfig

	generalLocker locker.GeneralLocker
}

func NewLockHandler(logger *zap.Logger, cfg *config.LockerConfig, locker locker.GeneralLocker) *lockerHandler {
	return &lockerHandler{
		logger:        logger.Named("LockHandler"),
		cfg:           cfg,
		generalLocker: locker,
	}
}

func (lh *lockerHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("Create"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	lockerRequest, err := v1locker.ParseLockRequest(r.Body, *lh.cfg.LockDefaultTimeout)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	if lockResponse := lh.generalLocker.ObtainLock(r.Context(), lockerRequest); lockResponse != nil {
		// obtained lock, send response to the client
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write(lockResponse.ToBytes()); err != nil {
			// failing to write the response to the client means we should free the lock
			logger.Error("Failed to write lock response to client", zap.Error(err))
			lh.generalLocker.ReleaseLock(lockResponse.SessionID)
		}
	}

	// in this case, the client should be disconnected or we are shutting down and they need to retry
	w.WriteHeader(http.StatusServiceUnavailable)
}

func (lh *lockerHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("Heartbeat"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	heartbeatError := lh.generalLocker.Heartbeat(namedParameters["_associated_id"])

	if heartbeatError == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusConflict)
		w.Write(heartbeatError.ToBytes())
	}
}

func (lh *lockerHandler) List(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("List"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	generalQuery, err := v1common.ParseGeneralAssociatedQuery(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	locks := v1locker.NewListLockResponse(lh.generalLocker.ListLocks(generalQuery))

	w.WriteHeader(http.StatusOK)
	w.Write(locks.ToBytes())
}

func (lh *lockerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("Delete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())

	lh.generalLocker.ReleaseLock(namedParameters["_associated_id"])
	w.WriteHeader(http.StatusNoContent)
}
