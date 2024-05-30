package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/middleware"

	lockmanager "github.com/DanLavine/willow/internal/locker/lock_manager"
	"github.com/DanLavine/willow/pkg/models/api"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"

	"go.uber.org/zap"
)

// Handles CRUD operations for Limit operations
type V1LockerHandler interface {
	// create or grab a lock
	Create(w http.ResponseWriter, r *http.Request)

	// query what locks exist
	Query(w http.ResponseWriter, r *http.Request)

	// releas a lock for another client to obtain, or delete the lock if there are no waiting clients
	Release(w http.ResponseWriter, r *http.Request)

	// Heartbeat is used to ensure that clients are still active and have the obtained locks
	Heartbeat(w http.ResponseWriter, r *http.Request)
}

type lockerHandler struct {
	cfg *config.LockerConfig

	generalLocker lockmanager.ExcluiveLocker
}

func NewLockHandler(cfg *config.LockerConfig, locker lockmanager.ExcluiveLocker) *lockerHandler {
	return &lockerHandler{
		cfg:           cfg,
		generalLocker: locker,
	}
}

func (lh *lockerHandler) Create(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Create")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the create lock request
	createLockerRequest := &v1locker.Lock{}
	if err := api.ObjectDecodeRequest(r, createLockerRequest); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// set the defaults for the lock request
	if createLockerRequest.Spec.Timeout == nil {
		createLockerRequest.Spec.Timeout = lh.cfg.LockDefaultTimeout
	}

	if lockResponse := lh.generalLocker.ObtainLock(r.Context(), createLockerRequest); lockResponse != nil {
		// obtained lock, send response to the client
		if _, respErr := api.ModelEncodeResponse(w, http.StatusOK, lockResponse); respErr != nil {
			// failing to write the response to the client means we should free the lock
			logger.Error("Failed to write lock response to client", zap.Error(respErr))
			_ = lh.generalLocker.Release(ctx, lockResponse.State.LockID, &v1locker.LockClaim{SessionID: lockResponse.State.SessionID})
		}

		return
	}

	// in this case, the client should be disconnected or we are shutting down and they need to retry
	_, _ = api.ModelEncodeResponse(w, http.StatusServiceUnavailable, nil)
}

func (lh *lockerHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Heartbeat")

	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the claim
	lockClaim := &v1locker.LockClaim{}
	if err := api.ModelDecodeRequest(r, lockClaim); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// obtain named parameters from the url
	namedParameters := urlrouter.GetNamedParamters(r.Context())

	// heartbeat the lock
	if heartbeatError := lh.generalLocker.Heartbeat(ctx, namedParameters["lock_id"], lockClaim); heartbeatError != nil {
		_, _ = api.ModelEncodeResponse(w, heartbeatError.StatusCode, heartbeatError)
		return
	}

	// heartbeat was successful
	_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
}

func (lh *lockerHandler) Query(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Query")

	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the associaed query
	query := &queryassociatedaction.AssociatedActionQuery{}
	if err := api.ModelDecodeRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// find the locks
	locks := lh.generalLocker.LocksQuery(ctx, query)

	// respond and ignore the errors
	_, _ = api.ModelEncodeResponse(w, http.StatusOK, locks)
}

func (lh *lockerHandler) Release(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "Query")

	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the claim
	lockClaim := &v1locker.LockClaim{}
	if err := api.ModelDecodeRequest(r, lockClaim); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// obtain named parameters from the url
	namedParameters := urlrouter.GetNamedParamters(r.Context())

	// release the lock
	lh.generalLocker.Release(ctx, namedParameters["lock_id"], lockClaim)

	// respond and ignore the errors
	_, _ = api.ModelEncodeResponse(w, http.StatusNoContent, nil)
}
