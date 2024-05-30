package clients

import "context"

type parentTrackerID string

const (
	TraceID parentTrackerID = "X-Request-ID"

	TraceHeader = "X-Request-ID"
)

// GetMiddlewareTraceID is the ID provided by the client's X-Request-ID and can be used to
// trace the logs for all the services that make up a single api request
func GetTraceIDFromContext(ctx context.Context) string {
	traceID := ctx.Value(TraceID)

	if traceID == nil {
		return ""
	}

	return traceID.(string)
}
