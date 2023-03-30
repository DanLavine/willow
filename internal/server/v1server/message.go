package v1server

import (
	"net/http"

	"github.com/DanLavine/willow/internal/logger"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

func (qh *queueHandler) Enqueue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Enqueue"), r)
	defer logger.Debug("processed enqueue request")

	switch method := r.Method; method {
	case "POST":
		logger.Debug("GET enqueue")

		enqueueItem, err := v1.ParseEnqueueItemRequest(r.Body)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		queue, err := qh.queueManager.Find(logger, enqueueItem.BrokerInfo.Name)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		if err = queue.Enqueue(logger, enqueueItem); err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (qh *queueHandler) Dequeue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Dequeue"), r)
	defer logger.Debug("processed dequeue request")

	switch method := r.Method; method {
	case "GET":
		logger.Debug("GET dequeue")

		//dequeueItemRequest, err := v1.ParseDequeueItemRequest(r.Body)
		//if err != nil {
		//	w.WriteHeader(err.StatusCode)
		//	w.Write(err.ToBytes())
		//	return
		//}

		w.WriteHeader(http.StatusOK)
		//w.Write(responseBody)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (qh *queueHandler) ACK(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("ACK"), r)
	defer logger.Debug("processed ack request")

	switch method := r.Method; method {
	case "POST":
		logger.Debug("POST ACK")

		// TODO

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
