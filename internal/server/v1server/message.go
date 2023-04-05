package v1server

import (
	"net/http"
	"reflect"

	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/internal/v1/tags"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

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

func (qh *queueHandler) Dequeue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Dequeue"), r)
	defer logger.Debug("processed dequeue request")

	switch method := r.Method; method {
	case "GET":
		logger.Debug("GET dequeue")

		matchRequst, err := v1.ParseMatchQueryRequest(r.Body)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		queue, err := qh.queueManager.Find(logger, matchRequst.BrokerName)
		if err != nil {
			w.WriteHeader(err.StatusCode)
			w.Write(err.ToBytes())
			return
		}

		readers := queue.Readers(matchRequst)
		if len(readers) == 0 {
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

		w.WriteHeader(http.StatusOK)
		w.Write(dequeueResponse.ToBytes())
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
