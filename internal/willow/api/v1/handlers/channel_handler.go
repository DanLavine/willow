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
)

func (qh queueHandler) ChannelQuery(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "ChannelQuery")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the request
	query := &queryassociatedaction.AssociatedActionQuery{}
	if err := api.ModelDecodeRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	channels, err := qh.queueClient.QueryChannels(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], query)
	if err != nil {
		logger.Warn("failed to query channels", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, channels)
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
