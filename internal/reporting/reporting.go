package reporting

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	constTraceHeader = "X-Request-Id"
)

type logger int
type header int

const (
	CustomLogger   logger = 1
	xRequestHeader header = 1
)

// Setup a context with a logger that has any tracing data setup on the logger
func SetupContextWithLoggerFromRequest(ctx context.Context, logger *zap.Logger, req *http.Request) (context.Context, *zap.Logger) {
	var loggerWithRequestID *zap.Logger

	if requestID := req.Header.Get(constTraceHeader); requestID != "" {
		loggerWithRequestID = logger.With(zap.String(constTraceHeader, requestID))
	} else {
		loggerWithRequestID = logger.With(zap.String(constTraceHeader, uuid.New().String()))
	}

	return SaveTraceHeaders(context.WithValue(ctx, CustomLogger, loggerWithRequestID), req.Header), loggerWithRequestID
}

// // Add a logger to a context
// func NamedLoggerFromContext(ctx context.Context, name string) (context.Context, *zap.Logger) {
// 	currentLogger := ctx.Value(CustomLogger).(*zap.Logger)
// 	newLogger := currentLogger.Named(name)

// 	return context.WithValue(ctx, CustomLogger, newLogger), newLogger
// }

// Get the logger saved in the context. This will panic if the logger is nil
func GetLogger(ctx context.Context) *zap.Logger {
	return ctx.Value(CustomLogger).(*zap.Logger)
}

func UpdateLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, CustomLogger, logger)
}

// Save any headers from an http request that might be interesting
func SaveTraceHeaders(ctx context.Context, headers http.Header) context.Context {
	xRequestHeaders := headers[constTraceHeader]
	if len(xRequestHeaders) == 1 {
		return context.WithValue(ctx, xRequestHeader, xRequestHeaders[0])
	}

	// malformed or missing
	return ctx
}

// Get the headers for a trace
func GetTraceHeaders(ctx context.Context) http.Header {
	headers := http.Header{}

	traceHeader := ctx.Value(xRequestHeader)
	if traceHeader != nil && traceHeader != "" {
		headers[constTraceHeader] = []string{traceHeader.(string)}
	}

	return headers
}

// obtain a base context with a base logger. This has no trace headeers
func StripedContext(logger *zap.Logger) context.Context {
	return context.WithValue(context.Background(), CustomLogger, zap.New(logger.Core()))
}
