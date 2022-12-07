package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DanLavine/willow/pkg/brokers"
	"go.uber.org/zap"
)

type metrics struct {
	logger *zap.Logger
	port   string

	brokerManager brokers.BrokerManager
}

func NewAdmin(logger *zap.Logger, port string, brokerManager brokers.BrokerManager) *metrics {
	return &metrics{
		logger:        logger.Named("tcp_server"),
		port:          port,
		brokerManager: brokerManager,
	}
}

func (m *metrics) Initialize() error { return nil }
func (m *metrics) Cleanup() error    { return nil }
func (m *metrics) Execute(ctx context.Context) error {
	errChan := make(chan error)
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", m.metrics)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", m.port),
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

func (m *metrics) metrics(res http.ResponseWriter, req *http.Request) {
	metrics := m.brokerManager.Metrics()
	body, err := json.Marshal(&metrics)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
		return
	}

	res.WriteHeader(200)
	res.Write(body)
}
