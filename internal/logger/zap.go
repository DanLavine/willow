package logger

import (
	"log"
	"net/http"

	"github.com/DanLavine/willow/internal/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// set the bas logger to a Nop for testing. In real code, this will be overwritten to the default logger
var BaseLogger = zap.NewNop()

func NewZapLogger(config config.Config) *zap.Logger {
	zapCfg := zap.NewProductionConfig()
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.DisableCaller = true
	zapCfg.DisableStacktrace = true
	zapCfg.Sampling = nil

	switch config.LogLevel() {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	// set the base logger for this package
	BaseLogger = logger

	return logger
}

func AddRequestID(logger *zap.Logger, req *http.Request) *zap.Logger {
	if requestID := req.Header.Get("request_id"); requestID != "" {
		return logger.With(zap.String("request_id", requestID))
	}

	return logger.With(zap.String("request_id", uuid.New().String()))
}

func StripRequestID(logger *zap.Logger) *zap.Logger {
	return logger.With(zap.String("request_id", ""))
}
