package handlers

import (
	"net/http"

	"github.com/DanLavine/willow/internal/willow/brokers/queues"
	"go.uber.org/zap"
)

type V1MetricsHandler interface {
	// Get all metrics for all queues
	Metrics(res http.ResponseWriter, req *http.Request)
}

type metricsHandler struct {
	logger *zap.Logger

	queueManager queues.QueueManager
}

func NewV1MetricsHandler(logger *zap.Logger, queueManager queues.QueueManager) *metricsHandler {
	return &metricsHandler{
		logger:       logger,
		queueManager: queueManager,
	}
}

func (mh *metricsHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	//logger := logger.AddRequestID(mh.logger.Named("Metrics"), r)

	switch method := r.Method; method {
	case "GET":
		metrics := mh.queueManager.Metrics()

		w.WriteHeader(http.StatusOK)
		w.Write(metrics.ToBytes())
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
