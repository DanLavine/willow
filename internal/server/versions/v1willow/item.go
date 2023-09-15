package v1willow

import (
	"net/http"

	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/server/client"
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
)

// Enqueue handler adds an item onto a message queue, or updates a message queue that is waiting to process
func (qh *queueHandler) Enqueue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Enqueue"), r)
	defer logger.Debug("processed enqueue request")

	switch method := r.Method; method {
	case "POST":
		logger.Debug("POST enqueue")

		enqueueItem, err := v1willow.ParseEnqueueItemRequest(r.Body)
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

// Dequeue handler removes an item from a message queue. If there are no messages waiting to process,
// the connection stays open until:
// 1. the client closes the connection
// 2. a message is processed and sent to the client
func (qh *queueHandler) Dequeue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Dequeue"), r)
	defer logger.Debug("processed dequeue request")

	switch method := r.Method; method {
	case "GET":
		logger.Debug("GET")

		dequeueItemRequest, err := v1willow.ParseDequeueItemRequest(r.Body)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		queue, err := qh.queueManager.Find(logger, dequeueItemRequest.Name)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		dequeueItem, success, failure, err := queue.Dequeue(logger, r.Context(), dequeueItemRequest.Selection)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		data, err := dequeueItem.ToBytes()
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			failure()
			return
		}

		w.WriteHeader(http.StatusOK)
		_, writeErr := w.Write(data)
		if writeErr == nil {
			// record which client is processing which item
			clientTracker := r.Context().Value("clientTracker").(client.Tracker)
			clientTracker.Add(dequeueItem.ID, dequeueItem.BrokerInfo)

			// successfuly sent to the client
			success()
		} else {
			// failed to send to the client
			failure()
		}
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

		ack, err := v1willow.ParseACKRequest(r.Body)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		queue, err := qh.queueManager.Find(logger, ack.BrokerInfo.Name)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		if err = queue.ACK(logger, ack); err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
