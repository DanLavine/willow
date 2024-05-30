package clients

import (
	"context"
	"net/http"

	"github.com/DanLavine/willow/pkg/encoding"
)

func AppendHeaders(req *http.Request, headers http.Header) {
	for headerKey, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerKey, headerValue)
		}
	}
}

func AddHeadersFromContext(req *http.Request, ctx context.Context) {
	if ctx == nil {
		req.Header.Set(encoding.ContentTypeHeader, "application/json")
		return
	}

	// Get the X-Request-ID if it is set
	traceId := ctx.Value(TraceID)
	if traceId != nil {
		req.Header.Set(TraceHeader, traceId.(string))
	}

	// Get the Content type header
	contentType := ctx.Value(encoding.EncoderType)
	if contentType != nil {
		req.Header.Set(encoding.ContentTypeHeader, contentType.(string))
	} else {
		req.Header.Set(encoding.ContentTypeHeader, "application/json")
	}
}
