package v1server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DanLavine/willow-message/protocol"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/brokers/v1brokers"
	"github.com/DanLavine/willow/pkg/errors"
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
	logger       *zap.Logger
	queueManager v1brokers.QueueManager
}

func NewQueueHandler(logger *zap.Logger, queueManager v1brokers.QueueManager) *queueHandler {
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

		if err := qh.queueManager.CreateQueue(logger, createRequest); err != nil {
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(err.StatusCode)
			w.Write(errResp)
			return
		}

		logger.Debug("processed create queue request")
		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (qh *queueHandler) RetrieveMessage(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Ready"), r)

	switch method := r.Method; method {
	case "GET":
		logger.Debug("processing retrieve request")

		readyBody, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("failed reading ready request", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		readyRequest := &v1.Ready{}
		if err = json.Unmarshal(readyBody, readyRequest); err != nil {
			logger.Error("failed parsing request", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		switch readyRequest.BrokerType {
		case protocol.Queue:
			message, retErr := qh.queueManager.RetrieveMessage(logger, readyRequest)
			if err != nil {
				errResp, _ := json.Marshal(v1.Error{Message: retErr.Error()})

				w.WriteHeader(retErr.StatusCode)
				w.Write(errResp)
				return
			}

			responseBody, err := json.Marshal(message)
			if err != nil {
				logger.Error("failed parsing request", zap.Error(err))

				// never seen this actualy fail, so just ignore it for now
				errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

				w.WriteHeader(http.StatusInternalServerError)
				w.Write(errResp)
				return
			}

			logger.Debug("processed ready request")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		default:
			err := errors.ProtocolNotSupportedError.Expected("queue").Actual(readyRequest.BrokerType.ToString())
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(err.StatusCode)
			w.Write(errResp)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (qh *queueHandler) Message(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("Message"), r)

	switch method := r.Method; method {
	case "POST":
		logger.Debug("processing enque request")

		messageBody, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("failed reading message body", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		messageRequest := &v1.Message{}
		if err = json.Unmarshal(messageBody, messageRequest); err != nil {
			logger.Error("failed parsing publish request", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		switch messageRequest.BrokerType {
		case protocol.Queue:
			if enqueErr := qh.queueManager.EnqueMessage(logger, messageRequest); enqueErr != nil {
				errResp, _ := json.Marshal(v1.Error{Message: enqueErr.Error()})

				w.WriteHeader(enqueErr.StatusCode)
				w.Write(errResp)
				return
			}

			logger.Debug("processed enque request")
			w.WriteHeader(http.StatusOK)
		default:
			err := errors.ProtocolNotSupportedError.Expected("queue").Actual(messageRequest.BrokerType.ToString())
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(err.StatusCode)
			w.Write(errResp)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (qh *queueHandler) ACK(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("ACK"), r)

	switch method := r.Method; method {
	case "POST":
		logger.Debug("handling ack message")

		ACKBody, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("failed reading ack body", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		ACKRequest := &v1.ACK{}
		if err = json.Unmarshal(ACKBody, ACKRequest); err != nil {
			logger.Error("failed parsing ack request", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errResp)
			return
		}

		switch ACKRequest.BrokerType {
		case protocol.Queue:
			if ackErr := qh.queueManager.ACKMessage(logger, ACKRequest); ackErr != nil {
				errResp, _ := json.Marshal(v1.Error{Message: ackErr.Error()})

				w.WriteHeader(ackErr.StatusCode)
				w.Write(errResp)
				return
			}

			logger.Debug("processed enque request")
			w.WriteHeader(http.StatusOK)
		default:
			err := errors.ProtocolNotSupportedError.Expected("queue").Actual(ACKRequest.BrokerType.ToString())
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(err.StatusCode)
			w.Write(errResp)
		}

		logger.Debug("handled ack message")
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
