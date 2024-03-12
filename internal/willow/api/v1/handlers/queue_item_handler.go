package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
	"go.uber.org/zap"
)

func (qh queueHandler) ItemACK(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), qh.logger.Named("ItemAck"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	ack := &v1willow.ACK{}
	if err := api.DecodeAndValidateHttpRequest(r, ack); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.Ack(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], ack); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
}

func (qh queueHandler) ItemHeartbeat(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), qh.logger.Named("ItemHeartbeat"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the enueue item
	heartbeat := &v1willow.Heartbeat{}
	if err := api.DecodeAndValidateHttpRequest(r, heartbeat); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if err := qh.queueClient.Heartbeat(ctx, urlrouter.GetNamedParamters(r.Context())["queue_name"], heartbeat); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
}
