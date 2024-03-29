package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api"

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

	generalLocker lockmanager.ExcluiveLocker
}

func NewLockHandler(logger *zap.Logger, cfg *config.LockerConfig, locker lockmanager.ExcluiveLocker) *lockerHandler {
	return &lockerHandler{
		logger:        logger.Named("LockHandler"),
		cfg:           cfg,
		generalLocker: locker,
	}
}

func (lh *lockerHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), lh.logger.Named("Create"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the create lock request
	createLockerRequest := &v1locker.LockCreateRequest{}
	if err := api.DecodeAndValidateHttpRequest(r, createLockerRequest); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// set the defaults for the lock request
	if createLockerRequest.LockTimeout == 0 {
		createLockerRequest.LockTimeout = *lh.cfg.LockDefaultTimeout
	}

	if lockResponse := lh.generalLocker.ObtainLock(ctx, createLockerRequest); lockResponse != nil {
		// obtained lock, send response to the client
		if _, respErr := api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, lockResponse); respErr != nil {
			// failing to write the response to the client means we should free the lock
			logger.Error("Failed to write lock response to client", zap.Error(respErr))
			lh.generalLocker.Release(ctx, lockResponse.LockID, &v1locker.LockClaim{SessionID: lockResponse.SessionID})
		}

		return
	}

	// in this case, the client should be disconnected or we are shutting down and they need to retry
	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusServiceUnavailable, nil)
}

func (lh *lockerHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), lh.logger.Named("Heartbeat"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the claim
	lockClaim := &v1locker.LockClaim{}
	if err := api.DecodeAndValidateHttpRequest(r, lockClaim); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// obtain named parameters from the url
	namedParameters := urlrouter.GetNamedParamters(r.Context())

	// heartbeat the lock
	if heartbeatError := lh.generalLocker.Heartbeat(ctx, namedParameters["lock_id"], lockClaim); heartbeatError != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, heartbeatError.StatusCode, heartbeatError)
		return
	}

	// heartbeat was successful
	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
}

func (lh *lockerHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), lh.logger.Named("List"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the associaed query
	query := &v1common.AssociatedQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// find the locks
	locks := lh.generalLocker.LocksQuery(ctx, query)

	// respond and ignore the errors
	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, &locks)
}

func (lh *lockerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), lh.logger.Named("Delete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the claim
	lockClaim := &v1locker.LockClaim{}
	if err := api.DecodeAndValidateHttpRequest(r, lockClaim); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// obtain named parameters from the url
	namedParameters := urlrouter.GetNamedParamters(r.Context())

	// release the lock
	lh.generalLocker.Release(ctx, namedParameters["lock_id"], lockClaim)

	// respond and ignore the errors
	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusNoContent, nil)
}
