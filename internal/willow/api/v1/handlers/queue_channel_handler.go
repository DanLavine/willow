package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

func (qh queueHandler) ChannelEnqueue(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), qh.logger.Named("ChannelEnqueue"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	queueItem := &v1willow.EnqueueQueueItem{}
	if err := api.DecodeAndValidateHttpRequest(r, queueItem); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.Enqueue(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], queueItem); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusCreated, nil)
}

func (qh queueHandler) ChannelDequeue(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), qh.logger.Named("ChannelDequeue"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	query := &queryassociatedaction.AssociatedActionQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	dequeueItem, successCallback, failureCallback, err := qh.queueClient.Dequeue(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], query)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if _, responseErr := api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, dequeueItem); responseErr != nil {
		logger.Warn("Failed so send the response back to the client")
		failureCallback()
	} else {
		successCallback()
	}
}

func (qh queueHandler) ChannelDelete(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), qh.logger.Named("ChannelDelete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the delte request
	keyValues := datatypes.KeyValues{}
	if err := api.DecodeHttpRequest(r, &keyValues); err != nil {
		logger.Warn("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if err := keyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		logger.Warn("failed to validated request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusBadRequest, errors.ServerErrorModelRequestValidation(err))
		return
	}

	if err := qh.queueClient.DeleteChannel(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], keyValues); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusNoContent, nil)
}
