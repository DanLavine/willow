package reporting

import (
	"log"

	"github.com/DanLavine/willow/internal/config"
	"go.uber.org/zap"
)

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

	return logger
}

func BaseLogger(logger *zap.Logger) *zap.Logger {
	return zap.New(logger.Core())
}
