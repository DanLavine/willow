package middleware

import (
	"context"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type requestLogger string

const (
	LogCtxID     requestLogger = "Request-ID"
	LoggerCtxKey requestLogger = "logger"
)

func AddLogger(baseLogger *zap.Logger, child http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()

		// setup the logger and save it to the request's original id as well as our generated id
		traceLogger := baseLogger.With(
			zap.String(clients.TraceHeader, clients.GetTraceIDFromContext(r.Context())), // global trace from the client
			zap.String(string(LogCtxID), id),                                            // trace for the current request
		)

		r = r.Clone(
			context.WithValue(
				context.WithValue(r.Context(), LoggerCtxKey, traceLogger),
				LogCtxID, id,
			),
		)

		// call the child
		child(w, r)
	}
}

func GetMiddlewareRequestID(ctx context.Context) string {
	return ctx.Value(LogCtxID).(string)
}

// GetMiddlewareLogger grabs the logger set on all http requests through the shared middleware
func GetMiddlewareLogger(ctx context.Context) *zap.Logger {
	return ctx.Value(LoggerCtxKey).(*zap.Logger)
}

// GetNameMiddlewareLogger grabs the logger set on all http requests through the shared middleware. It
// also updates the context with the newly named logger that can be used to child calls
func GetNamedMiddlewareLogger(ctx context.Context, name string) (context.Context, *zap.Logger) {
	namedLogger := ctx.Value(LoggerCtxKey).(*zap.Logger).Named(name)
	return context.WithValue(ctx, LoggerCtxKey, namedLogger), namedLogger
}
