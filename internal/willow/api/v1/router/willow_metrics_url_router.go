package router

import (
	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/willow/api/v1/handlers"
)

func AddV1WillowMetricsRoutes(mux *urlrouter.Router, v1Handler handlers.V1MetricsHandler) {
	mux.HandleFunc("GET", "/v1/metrics", v1Handler.Metrics)
}
