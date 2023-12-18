package handlers

import (
	"net/http"

	"github.com/DanLavine/willow/internal/logger"
)

// Enqueue handler adds an item onto a message queue, or updates a message queue that is waiting to process
func (qh *queueHandler) Enqueue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Enqueue"), r)
	defer logger.Debug("processed enqueue request")

	switch method := r.Method; method {
	case "POST":
		logger.Debug("POST enqueue")

		enqueueItem, err := ParseEnqueueItemRequest(r.Body)
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
func (qh *queueHandler) Dequeue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Dequeue"), r)
	defer logger.Debug("processed dequeue request")

	switch method := r.Method; method {
	case "GET":
		logger.Debug("GET")

		dequeueItemRequest, err := ParseDequeueItemRequest(r.Body)
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

		dequeueItem, success, failure, err := queue.Dequeue(logger, r.Context(), dequeueItemRequest.Query)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		w.WriteHeader(http.StatusOK)
		_, writeErr := w.Write(dequeueItem.ToBytes())
		if writeErr == nil {
			// sent to the client, so advance item in the queue
			success()
		} else {
			// failed to send to the client, so ensure the item will be picked up again or squashed
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

		ack, err := ParseACKRequest(r.Body)
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
