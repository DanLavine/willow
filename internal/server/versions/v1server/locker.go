package v1server

import (
	"net/http"

	"github.com/DanLavine/willow/internal/locker"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"go.uber.org/zap"
)

// Handles CRUD operations for Limit operations
//
//go:generate mockgen -destination=v1serverfakes/locker_mock.go -package=v1serverfakes github.com/DanLavine/willow/internal/server/versions/v1server LockerHandler
type LockerHandler interface {
	// create a group rule
	Create(w http.ResponseWriter, r *http.Request)

	// read group rules
	List(w http.ResponseWriter, r *http.Request)

	// delete a group rule
	Delete(w http.ResponseWriter, r *http.Request)
}

type lockerHandler struct {
	logger *zap.Logger

	generalLocker locker.GeneralLocker
}

func NewLockHandler(logger *zap.Logger) *lockerHandler {
	return &lockerHandler{
		logger:        logger.Named("LockHandler"),
		generalLocker: locker.NewGeneralLocker(nil),
	}
}

func (lh *lockerHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("Create"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	switch method := r.Method; method {
	case "POST":
		lockerRequest, err := v1locker.ParseLockRequest(r.Body)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		// create the lock rules, and wait until they are obtained
		// TODO: need to make use of the callback for disconnects to free up locks
		//disconnectCallback := lh.generalLocker.ObtainLocks(r.Context(), lockerRequest.KeyValues)
		_ = lh.generalLocker.ObtainLocks(r.Context(), lockerRequest.KeyValues)

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (lh *lockerHandler) List(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("List"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	w.WriteHeader(http.StatusNotImplemented)
}

func (lh *lockerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(lh.logger.Named("Delete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	switch method := r.Method; method {
	case "DELETE":
		lockerRequest, err := v1locker.ParseLockRequest(r.Body)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		lh.generalLocker.FreeLocks(lockerRequest.KeyValues)

		// TODO: clear out the value in the client being tracked.

		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
