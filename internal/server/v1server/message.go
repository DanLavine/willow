package v1server

import (
	"net/http"
	"reflect"

	"github.com/DanLavine/willow/internal/brokers/tags"
	"github.com/DanLavine/willow/internal/logger"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

// Enqueue handler adds an item onto a message queue, or updates a message queue that is waiting to process
func (qh *queueHandler) Enqueue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Enqueue"), r)
	defer logger.Debug("processed enqueue request")

	switch method := r.Method; method {
	case "POST":
		logger.Debug("POST enqueue")

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

		query, err := v1.ParseReaderSelect(r.Body)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		queue, err := qh.queueManager.Find(logger, query.BrokerName)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		readers, err := queue.Readers(logger, query)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		} else if len(readers) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// setup select cases
		selectCases := []reflect.SelectCase{reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(r.Cancel)}}
		for _, reader := range readers {
			selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(reader)})
		}

		id, value, _ := reflect.Select(selectCases)
		if id == 0 {
			// In this case, the client was disconnected so just return
			return
		}

		// call the dequeue function
		dequeueResponse := value.Interface().(tags.Tag)()

		// TODO: on an error, we need to mark the message as failed?
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(dequeueResponse.ToBytes())
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
