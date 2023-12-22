package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"

	lockmanager "github.com/DanLavine/willow/internal/locker/lock_manager"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	"go.uber.org/zap"
)

// Handles CRUD operations for Limit operations
type V1LockerHandler interface {
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

	generalLocker lockmanager.GeneralLocker
}

func NewLockHandler(logger *zap.Logger, cfg *config.LockerConfig, locker lockmanager.GeneralLocker) *lockerHandler {
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

	// parse the create lock request
	createLockerRequest := &v1locker.LockCreateRequest{}
	if err := createLockerRequest.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// set the defaults for the lock request
	if createLockerRequest.Timeout == 0 {
		createLockerRequest.Timeout = *lh.cfg.LockDefaultTimeout
	}

	if lockResponse := lh.generalLocker.ObtainLock(r.Context(), createLockerRequest); lockResponse != nil {
		// obtained lock, send response to the client
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		if _, respErr := api.HttpResponse(r, w, http.StatusCreated, lockResponse); respErr != nil {
			// failing to write the response to the client means we should free the lock
			logger.Error("Failed to write lock response to client", zap.Error(respErr))
			lh.generalLocker.ReleaseLock(lockResponse.SessionID)
		}
	}

	// in this case, the client should be disconnected or we are shutting down and they need to retry
	_, _ = api.HttpResponse(r, w, http.StatusServiceUnavailable, nil)
}

func (lh *lockerHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("Heartbeat"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// obtain named parameters from the url
	namedParameters := urlrouter.GetNamedParamters(r.Context())

	// heartbeat the lock
	heartbeatError := lh.generalLocker.Heartbeat(namedParameters["_associated_id"])

	if heartbeatError == nil {
		// heartbeat was successful
		_, _ = api.HttpResponse(r, w, http.StatusOK, nil)
	} else {
		// there was an error heartbeating
		_, _ = api.HttpResponse(r, w, heartbeatError.StatusCode, heartbeatError)
	}
}

func (lh *lockerHandler) List(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("List"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the associaed query
	query := &v1common.AssociatedQuery{}
	if err := query.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// find the locks
	locks := lh.generalLocker.LocksQuery(query)

	// respond and ignore the errors
	_, _ = api.HttpResponse(r, w, http.StatusOK, locks)
}

func (lh *lockerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("Delete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// obtain named parameters from the url
	namedParameters := urlrouter.GetNamedParamters(r.Context())

	// release the lock
	lh.generalLocker.ReleaseLock(namedParameters["_associated_id"])

	// respond and ignore the errors
	_, _ = api.HttpResponse(r, w, http.StatusNoContent, nil)
}
