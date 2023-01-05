package v1server

import (
	"encoding/json"
	"io"
	"net/http"

	v1 "github.com/DanLavine/willow-message/protocol/v1"
	deadletterqueue "github.com/DanLavine/willow/pkg/dead-letter-queue"
	"github.com/DanLavine/willow/pkg/logger"
	"go.uber.org/zap"
)

type QueueHandler interface {
	Create(w http.ResponseWriter, r *http.Request)

	Message(w http.ResponseWriter, r *http.Request)
	RetrieveMessage(w http.ResponseWriter, r *http.Request)
	ACK(w http.ResponseWriter, r *http.Request)
}

type queueHandler struct {
	logger *zap.Logger

	deadLetterQueue deadletterqueue.DeadLetterQueue
}

func NewQueueHandler(logger *zap.Logger, deadLetterQueue deadletterqueue.DeadLetterQueue) *queueHandler {
	return &queueHandler{
		logger:          logger,
		deadLetterQueue: deadLetterQueue,
	}
}

func (qh *queueHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Create"), r)

	switch method := r.Method; method {
	case "POST":
		logger.Debug("processing create queue request")

		createRequestBody, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("failed reading request", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		createRequest := &v1.Create{}
		if err = json.Unmarshal(createRequestBody, createRequest); err != nil {
			logger.Error("failed parsing request", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		if createErr := qh.deadLetterQueue.Create(createRequest.BrokerName, createRequest.BrokerTag); createErr != nil {
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
