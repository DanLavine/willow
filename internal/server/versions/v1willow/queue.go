package v1willow

import (
	"encoding/json"
	"net/http"

	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"go.uber.org/zap"
)

type QueueHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Enqueue(w http.ResponseWriter, r *http.Request)
	Dequeue(w http.ResponseWriter, r *http.Request)
	ACK(w http.ResponseWriter, r *http.Request)
}

type queueHandler struct {
	logger *zap.Logger

	queueManager queues.QueueManager
}

func NewQueueHandler(logger *zap.Logger, queueManager queues.QueueManager) *queueHandler {
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

		createRequest, err := v1willow.ParseCreateRequest(r.Body)
		if err != nil {
			logger.Error("failed parsing request", zap.Error(err))

			w.WriteHeader(err.StatusCode)
			w.Write([]byte(err.Error()))
			return
		}

		if createErr := qh.queueManager.Create(logger, createRequest); createErr != nil {
			logger.Error("failed creating queue", zap.Error(createErr))
			errResp, _ := json.Marshal(createErr)

			w.WriteHeader(createErr.StatusCode)
			w.Write(errResp)
			return
		}

		logger.Debug("processed create queue request")
		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
