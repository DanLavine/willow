package v1server

import (
	"context"
	"net/http"

	"github.com/DanLavine/willow/internal/locker"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/server/client"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"go.uber.org/zap"
)

// Handles CRUD operations for Limit operations
//
//go:generate mockgen -destination=v1serverfakes/locker_mock.go -package=v1serverfakes github.com/DanLavine/willow/internal/server/versions/v1server LockerHandler
type LockerHandler interface {
	// create a group rule
	Create(context context.Context) func(w http.ResponseWriter, r *http.Request)

	// read group rules
	List(context context.Context) func(w http.ResponseWriter, r *http.Request)

	// delete a group rule
	Delete(context context.Context) func(w http.ResponseWriter, r *http.Request)
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

func (lh *lockerHandler) Create(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

			if disconnectCallback := lh.generalLocker.ObtainLocks(ctx, r.Context(), lockerRequest.KeyValues); disconnectCallback != nil {
				clientTracker := r.Context().Value("clientTracker").(client.LockerTracker)
				clientTracker.AddReleaseCallback(lockerRequest.KeyValues, disconnectCallback)

				w.WriteHeader(http.StatusOK)
			} else {
				select {
				case <-ctx.Done():
					// in this case the server is restarting
					w.WriteHeader(http.StatusServiceUnavailable)
				default:
					// in this case, the client should be disconnected, but try to send something anyways
					w.WriteHeader(http.StatusBadGateway)
				}
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (lh *lockerHandler) List(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.AddRequestID(lh.logger.Named("List"), r)
		logger.Debug("starting request")
		defer logger.Debug("processed request")

		switch method := r.Method; method {
		case "GET":
			lockResponse := &v1locker.LockResponse{
				Locks: lh.generalLocker.ListLocks(),
			}

			w.WriteHeader(http.StatusOK)
			w.Write(lockResponse.ToBytes())
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (lh *lockerHandler) Delete(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

			// only remove the items if the client has the locks for the desired key values
			clientTracker := r.Context().Value("clientTracker").(client.LockerTracker)
			if clientTracker.HasLocks(lockerRequest.KeyValues) {
				clientTracker.ClearLocks(lockerRequest.KeyValues)

				w.WriteHeader(http.StatusNoContent)
			} else {
				err := &api.Error{Message: "client does not have the lock for key values group", StatusCode: http.StatusBadRequest}
				w.WriteHeader(err.StatusCode)
				w.Write(err.ToBytes())
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
