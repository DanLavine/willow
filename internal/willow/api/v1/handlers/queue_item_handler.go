package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/models/api"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
	"go.uber.org/zap"
)

func (qh queueHandler) ItemACK(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "ItemACK")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	ack := &v1willow.ACK{}
	if err := api.ModelDecodeRequest(r, ack); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.Ack(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], ack); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
}

func (qh queueHandler) ItemHeartbeat(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "ItemHeartbeat")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	heartbeat := &v1willow.Heartbeat{}
	if err := api.ModelDecodeRequest(r, heartbeat); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.Heartbeat(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], heartbeat); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
}
