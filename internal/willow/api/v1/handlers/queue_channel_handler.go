package handlers

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

func (qh queueHandler) ChannelEnqueue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("ChannelEnqueue"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	queueItem := &v1willow.EnqueueQueueItem{}
	if err := api.DecodeAndValidateHttpRequest(r, queueItem); err != nil {
		logger.Error("failed to decode enqueue item request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.Enqueue(logger, urlrouter.GetNamedParamters(r.Context())["queue_name"], queueItem); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusCreated, nil)
}

func (qh queueHandler) ChannelDequeue(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("ChannelDequeue"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	query := &datatypes.AssociatedKeyValuesQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Error("failed to decode enqueue item request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	dequeueItem, successCallback, failureCallback, err := qh.queueClient.Dequeue(logger, r.Context(), urlrouter.GetNamedParamters(r.Context())["queue_name"], *query)
	fmt.Printf("dequeue item: %#v\n", dequeueItem)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if _, responseErr := api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, dequeueItem); responseErr != nil {
		logger.Info("Failed so send the response back to the client")
		failureCallback()
	} else {
		successCallback()
	}
}

func (qh queueHandler) ChannelDelete(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(qh.logger.Named("ChannelDelete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the delte request
	keyValues := &datatypes.KeyValues{}
	if err := api.DecodeAndValidateHttpRequest(r, keyValues); err != nil {
		logger.Error("failed to decode enqueue item request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.DeleteChannel(logger, urlrouter.GetNamedParamters(r.Context())["queue_name"], *keyValues); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusNoContent, nil)
}
