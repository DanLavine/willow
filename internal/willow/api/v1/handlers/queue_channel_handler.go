package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

func (qh queueHandler) ChannelEnqueue(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "ChannelEnqueue")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	queueItem := &v1willow.EnqueueQueueItem{}
	if err := api.ModelDecodeRequest(r, queueItem); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.Enqueue(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], queueItem); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusCreated, nil)
}

func (qh queueHandler) ChannelDequeue(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "ChannelDequeue")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	query := &queryassociatedaction.AssociatedActionQuery{}
	if err := api.ModelDecodeRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	dequeueItem, successCallback, failureCallback, err := qh.queueClient.Dequeue(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], query)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	if _, responseErr := api.ModelEncodeResponse(w, http.StatusOK, dequeueItem); responseErr != nil {
		logger.Warn("Failed so send the response back to the client")
		failureCallback()
	} else {
		successCallback()
	}
}

func (qh queueHandler) ChannelDelete(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "ChannelDelete")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the delte request
	keyValues := datatypes.KeyValues{}
	if err := json.NewDecoder(r.Body).Decode(&keyValues); err != nil {
		logger.Warn("failed to decode request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, http.StatusBadRequest, errors.ServerErrorModelRequestValidation(err))
	}

	if err := keyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		logger.Warn("failed to validated request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, http.StatusBadRequest, errors.ServerErrorModelRequestValidation(err))
	}

	if err := qh.queueClient.DeleteChannel(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], keyValues); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusNoContent, nil)
}
