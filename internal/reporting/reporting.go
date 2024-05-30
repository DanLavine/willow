package reporting

import (
	"context"
	"net/http"

	"github.com/DanLavine/willow/internal/middleware"
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
	ctx, traceID := saveTraceHeaders(ctx, req.Header)
	loggerWithRequestID := logger.With(zap.String(constTraceHeader, traceID))

	return context.WithValue(ctx, CustomLogger, loggerWithRequestID), loggerWithRequestID
}

// Get the logger saved in the context. This will panic if the logger is nil
func GetLogger(ctx context.Context) *zap.Logger {
	return ctx.Value(CustomLogger).(*zap.Logger)
}

func UpdateLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, CustomLogger, logger)
}

// Save any headers from an http request that might be interesting
func saveTraceHeaders(ctx context.Context, headers http.Header) (context.Context, string) {
	xRequestHeaders := headers.Get(constTraceHeader)
	if xRequestHeaders != "" {
		return context.WithValue(ctx, xRequestHeader, xRequestHeaders), xRequestHeaders
	}

	// malformed or missing
	traceID := uuid.New().String()
	return context.WithValue(ctx, xRequestHeader, traceID), traceID
}

// // Get the headers for a trace
// func GetTraceHeaders(ctx context.Context) http.Header {
// 	headers := http.Header{}

// 	traceHeader := ctx.Value(xRequestHeader)
// 	if traceHeader != nil && traceHeader != "" {
// 		headers[constTraceHeader] = []string{traceHeader.(string)}
// 	}

// 	return headers
// }

// obtain a base context with a base logger. This has no trace headeers
func StripedContext(logger *zap.Logger) context.Context {
	return context.WithValue(context.Background(), middleware.LoggerCtxKey, zap.New(logger.Core()))
}
