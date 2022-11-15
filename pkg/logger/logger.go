package logger

import (
	"github.com/DanLavine/gomultiplex"
	"go.uber.org/zap"
)

type Logger struct {
	zapLogger *zap.Logger
}

func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{zapLogger: zapLogger}
}

func (l *Logger) Info(ctx *gomultiplex.Context, msg string) {
	if ctx == nil {
		l.zapLogger.Info(msg)
		return
	}

	fields := []zap.Field{}

	for key, value := range ctx.Fields {
		fields = append(fields, zap.Any(key, value))
	}

	l.zapLogger.Info(msg, fields...)
}

func (l *Logger) Error(ctx *gomultiplex.Context, msg string) {
	if ctx == nil {
		l.zapLogger.Error(msg)
		return
	}

	fields := []zap.Field{}

	for key, value := range ctx.Fields {
		fields = append(fields, zap.Any(key, value))
	}

	l.zapLogger.Error(msg, fields...)
}
