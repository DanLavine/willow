package logger

import (
	"log"
	"net/http"

	"github.com/DanLavine/willow/internal/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func NewZapLogger(config *config.Config) *zap.Logger {
	zapCfg := zap.NewProductionConfig()
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.DisableCaller = true
	zapCfg.DisableStacktrace = true

	switch config.LogLevel {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	return logger
}

func AddRequestID(logger *zap.Logger, req *http.Request) *zap.Logger {
	if requestID := req.Header.Get("request_id"); requestID != "" {
		return logger.With(zap.String("request_id", requestID))
	}

	return logger.With(zap.String("request_id", uuid.New().String()))
}
