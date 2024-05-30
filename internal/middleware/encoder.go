package middleware

import (
	"context"
	"net/http"

	"github.com/DanLavine/willow/pkg/encoding"
)

func ValidateReqHeaders(child http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqContentType := r.Header.Get(encoding.ContentTypeHeader)

		if reqContentType == "" {
			logger := GetMiddlewareLogger(r.Context())
			logger.Error("failed to get Content-Type header")

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"Message":"Content-Type header not specified"}`))
			return
		}

		// save the Content-Type for the encoder and save it to the request's original context
		r = r.Clone(context.WithValue(r.Context(), encoding.EncoderType, reqContentType))

		// save the content type to the response writer
		w.Header().Add(encoding.ContentTypeHeader, reqContentType)

		// call the child
		child(w, r)
	}
}
