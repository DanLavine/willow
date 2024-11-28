package logging

import (
	"context"
	"log"

	"go.uber.org/zap"
)

func NewZapLogger(logLevel string) *zap.Logger {
	zapCfg := zap.NewProductionConfig()
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.DisableCaller = true
	zapCfg.DisableStacktrace = true
	zapCfg.Sampling = nil

	switch logLevel {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	default:
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	return logger
}

func BaseLogger(logger *zap.Logger) *zap.Logger {
	return zap.New(logger.Core())
}

func StripedContext(logger *zap.Logger) context.Context {
	return context.WithValue(context.Background(), LoggerCtxKey, zap.New(logger.Core()))
}
