package handlers

import (
	"net/http"

	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/willow/brokers/queues"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"

	"go.uber.org/zap"
)

type V1QueueHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Enqueue(w http.ResponseWriter, r *http.Request)
	Dequeue(w http.ResponseWriter, r *http.Request)
	ACK(w http.ResponseWriter, r *http.Request)
}

type queueHandler struct {
	logger *zap.Logger

	queueManager queues.QueueManager
}

func NewV1QueueHandler(logger *zap.Logger, queueManager queues.QueueManager) *queueHandler {
	return &queueHandler{
		logger:       logger,
		queueManager: queueManager,
	}
}

func (qh *queueHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Create"), r)

	switch method := r.Method; method {
	case "POST":
		logger.Debug("processing create queue request")
		defer logger.Debug("processed create queue request")

		// parse the create request
		createRequest := &v1willow.Create{}
		if err := createRequest.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
			_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
			return
		}

		if createErr := qh.queueManager.Create(logger, createRequest); createErr != nil {
			logger.Error("failed creating queue", zap.Error(createErr))
			_, _ = api.HttpResponse(r, w, createErr.StatusCode, createErr)
			return
		}

		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
