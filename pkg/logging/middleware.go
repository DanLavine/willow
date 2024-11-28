package logging

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type requestLogger string

const (
	RequestIDHeader requestLogger = "X-Request-Id"
	LoggerCtxKey    requestLogger = "logger"
)

func MiddlewareSetLogger(baseLogger *zap.Logger, child http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xRequestHeaders := r.Header.Get(string(RequestIDHeader))
		if xRequestHeaders == "" {
			xRequestHeaders = uuid.New().String()
		}

		// setup the logger and save it to the request's original id as well as our generated id
		traceLogger := baseLogger.With(
			zap.String(string(RequestIDHeader), xRequestHeaders), // set the global trace id
		)

		r = r.Clone(
			context.WithValue(r.Context(), LoggerCtxKey, traceLogger),
		)

		// call the child
		child(w, r)
	}
}

func MiddlewareRequestID(ctx context.Context) string {
	return ctx.Value(RequestIDHeader).(string)
}

// GetMiddlewareLogger grabs the logger set on all http requests through the shared middleware
func MiddlewareLogger(ctx context.Context) *zap.Logger {
	return ctx.Value(LoggerCtxKey).(*zap.Logger)
}

// NameMiddlewareLogger grabs the logger set on all http requests through the shared middleware. It
// also updates the context with the newly named logger that can be used to child calls
func NamedMiddlewareLogger(ctx context.Context, name string) (context.Context, *zap.Logger) {
	namedLogger := ctx.Value(LoggerCtxKey).(*zap.Logger).Named(name)
	return context.WithValue(ctx, LoggerCtxKey, namedLogger), namedLogger
}
