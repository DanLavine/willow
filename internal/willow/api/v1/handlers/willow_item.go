package handlers

import (
	"net/http"

	"github.com/DanLavine/willow/internal/logger"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

// Enqueue handler adds an item onto a message queue, or updates a message queue that is waiting to process
func (qh *queueHandler) Enqueue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Enqueue"), r)
	defer logger.Debug("processed enqueue request")

	switch method := r.Method; method {
	case "POST":
		logger.Debug("POST enqueue")

		// parse the enqueue request
		enqueueItem := &v1willow.EnqueueItemRequest{}
		if err := enqueueItem.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
			_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
			return
		}

		// find the queue
		queue, err := qh.queueManager.Find(logger, enqueueItem.BrokerInfo.Name)
		if err != nil {
			_, _ = api.HttpResponse(r, w, err.StatusCode, err)
			return
		}

		// enqueu the item
		if err = queue.Enqueue(logger, enqueueItem); err != nil {
			_, _ = api.HttpResponse(r, w, err.StatusCode, err)
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

		// parse the dequeue request
		dequeueItemRequest := &v1willow.DequeueItemRequest{}
		if err := dequeueItemRequest.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
			_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
			return
		}

		// Find the queue
		queue, err := qh.queueManager.Find(logger, dequeueItemRequest.Name)
		if err != nil {
			_, _ = api.HttpResponse(r, w, err.StatusCode, err)
			return
		}

		// dequeue an item
		dequeueItem, success, failure, err := queue.Dequeue(logger, r.Context(), dequeueItemRequest.Query)
		if err != nil {
			_, _ = api.HttpResponse(r, w, err.StatusCode, err)
			return
		}

		// respond to the client with dequeued item
		_, writeErr := api.HttpResponse(r, w, http.StatusOK, dequeueItem)
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

		// parse the dequeue request
		ack := &v1willow.ACK{}
		if err := ack.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
			_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
			return
		}

		// find the queue
		queue, err := qh.queueManager.Find(logger, ack.BrokerInfo.Name)
		if err != nil {
			_, _ = api.HttpResponse(r, w, err.StatusCode, err)
			return
		}

		// ack the item
		if err = queue.ACK(logger, ack); err != nil {
			_, _ = api.HttpResponse(r, w, err.StatusCode, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
