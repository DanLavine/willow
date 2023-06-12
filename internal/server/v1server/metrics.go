package v1server

import (
	"net/http"

	"github.com/DanLavine/willow/internal/brokers/queues"
	"github.com/DanLavine/willow/internal/logger"
	"go.uber.org/zap"
)

type MetricsHandler interface {
	// Get all metrics for all queues
	Metrics(res http.ResponseWriter, req *http.Request)
}

type metricsHandler struct {
	logger *zap.Logger

	queueManager queues.QueueManager
}

func NewMetricsHandler(logger *zap.Logger, queueManager queues.QueueManager) *metricsHandler {
	return &metricsHandler{
		logger:       logger,
		queueManager: queueManager,
	}
}

func (mh *metricsHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(mh.logger.Named("Metrics"), r)

	switch method := r.Method; method {
	case "GET":
		//matchQuery, err := v1.ParseMatchQueryRequest(r.Body)
		//if err != nil {
		//	logger.Error("Failed to parse match request body", zap.Error(err))
		//	w.WriteHeader(err.StatusCode)
		//	w.Write([]byte(err.Error()))
		//	return
		//}

		metrics := mh.queueManager.Metrics()
		metricsData, err := metrics.ToBytes()
		if err != nil {
			logger.Error("Failed to encode metrics response", zap.Error(err))
			w.WriteHeader(err.StatusCode)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(metricsData)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
