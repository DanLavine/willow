package middleware

import (
	"context"
	"net/http"

	"github.com/DanLavine/willow/pkg/clients"
	"github.com/google/uuid"
)

func SetupTracer(child http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xRequestHeaders := r.Header.Get(clients.TraceHeader)

		// set the traceID if it is missing
		if xRequestHeaders == "" {
			xRequestHeaders = uuid.New().String()
			r.Header.Add(clients.TraceHeader, xRequestHeaders)
		}

		// save the trace ID to the context as well
		r = r.Clone(context.WithValue(r.Context(), clients.TraceID, xRequestHeaders))

		// call the child
		child(w, r)
	}
}
