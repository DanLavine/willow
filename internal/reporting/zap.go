package reporting

import (
	"log"

	"github.com/DanLavine/willow/internal/config"
	"go.uber.org/zap"
)

// set the bas logger to a Nop for testing. In real code, this will be overwritten to the default logger
//var BaseLogger = zap.NewNop()

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
	//BaseLogger = logger

	return logger
}

func BaseLogger(logger *zap.Logger) *zap.Logger {
	return zap.New(logger.Core())
	//return logger.With(zap.String("x_request_id", ""))
}
