package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DanLavine/willow/internal/server/versions/v1willow"
	"github.com/DanLavine/willow/pkg/config"
	"go.uber.org/zap"
)

type metrics struct {
	logger *zap.Logger
	config *config.WillowConfig

	metricsHandler v1willow.MetricsHandler
}

func NewMetrics(logger *zap.Logger, config *config.WillowConfig, metricsHandler v1willow.MetricsHandler) *metrics {
	return &metrics{
		logger:         logger.Named("tcp_server"),
		config:         config,
		metricsHandler: metricsHandler,
	}
}

func (m *metrics) Initialize() error { return nil }
func (m *metrics) Cleanup() error    { return nil }
func (m *metrics) Execute(ctx context.Context) error {
	errChan := make(chan error)
	mux := http.NewServeMux()
	// get metrics for all queues, and dead-lettere queues
	mux.HandleFunc("/v1/metrics", m.metricsHandler.Metrics)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", *m.config.MetricsPort),
		Handler: mux,
	}

	go func() {
		errChan <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(shutdownContext)
	case err := <-errChan:
		return err
	}
}
