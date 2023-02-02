package v1server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/DanLavine/willow/internal/logger"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

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

		messageRequest := &v1.EnqueMessage{}
		if err = json.Unmarshal(messageBody, messageRequest); err != nil {
			logger.Error("failed parsing publish request", zap.Error(err))

			// never seen this actualy fail, so just ignore it for now
			errResp, _ := json.Marshal(v1.Error{Message: err.Error()})

			w.WriteHeader(http.StatusBadRequest)
			w.Write(errResp)
			return
		}

		switch messageRequest.BrokerType {
		case v1.Queue:
			if enqueErr := qh.deadLetterQueue.Enqueue(messageRequest.Data, messageRequest.Updateable, messageRequest.BrokerTags); enqueErr != nil {
				errResp, _ := json.Marshal(enqueErr)
				w.WriteHeader(enqueErr.StatusCode)
				w.Write(errResp)
				return
			}

			logger.Debug("processed enque request")
			w.WriteHeader(http.StatusOK)
		default:
			err := (&v1.Error{Message: "Broker Type not supported", StatusCode: http.StatusBadRequest}).With("queue", messageRequest.BrokerType.ToString())
			errResp, _ := json.Marshal(err)

			w.WriteHeader(err.StatusCode)
			w.Write(errResp)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (qh *queueHandler) RetrieveMessage(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("RetrieveMessage"), r)

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
		case v1.Queue:
			message, retErr := qh.deadLetterQueue.Message(context.Background(), readyRequest.BrokerTagsMatch, readyRequest.BrokerTags)
			if retErr != nil {
				logger.Error("failed obtaining next message", zap.Error(retErr))
				errResp, _ := json.Marshal(retErr)

				w.WriteHeader(retErr.StatusCode)
				w.Write(errResp)
				return
			}

			responseBody, err := json.Marshal(message)
			if err != nil {
				logger.Error("failed parsing request", zap.Error(err))

				// never seen this actualy fail, so just ignore it for now
				errResp, _ := json.Marshal(v1.Error{Message: err.Error(), StatusCode: http.StatusInternalServerError})

				w.WriteHeader(http.StatusInternalServerError)
				w.Write(errResp)
				return
			}

			logger.Debug("processed ready request")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		default:
			err := (&v1.Error{Message: "Broker Type not supported", StatusCode: http.StatusBadRequest}).With("queue", readyRequest.BrokerType.ToString())
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

			w.WriteHeader(http.StatusBadRequest)
			w.Write(errResp)
			return
		}

		switch ACKRequest.BrokerType {
		case v1.Queue:
			if ackErr := qh.deadLetterQueue.ACK(ACKRequest.ID, ACKRequest.Passed, ACKRequest.BrokerTags); ackErr != nil {
				errResp, _ := json.Marshal(v1.Error{Message: ackErr.Error()})

				w.WriteHeader(ackErr.StatusCode)
				w.Write(errResp)
				return
			}

			logger.Debug("processed enque request")
			w.WriteHeader(http.StatusOK)
		default:
			err := (&v1.Error{Message: "Broker Type not supported", StatusCode: http.StatusBadRequest}).With("queue", ACKRequest.BrokerType.ToString())
			errResp, _ := json.Marshal(err)

			w.WriteHeader(err.StatusCode)
			w.Write(errResp)
		}

		logger.Debug("handled ack message")
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
