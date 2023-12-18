package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"go.uber.org/zap"
)

type metrics struct {
	logger *zap.Logger
	config *config.WillowConfig

	mux *urlrouter.Router
}

func NewMetrics(logger *zap.Logger, config *config.WillowConfig, mux *urlrouter.Router) *metrics {
	return &metrics{
		logger: logger.Named("tcp_server"),
		config: config,
		mux:    mux,
	}
}

func (m *metrics) Initialize() error { return nil }
func (m *metrics) Cleanup() error    { return nil }
func (m *metrics) Execute(ctx context.Context) error {
	errChan := make(chan error)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", *m.config.MetricsPort),
		Handler: m.mux,
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
