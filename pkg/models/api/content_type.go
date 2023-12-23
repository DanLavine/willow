package api

import "net/http"

const (
	ContentTypeJSON = "application/json"

	// Add this eventual to encode/decode bytes directly which will be faster.
	// useful when will is a high traffic queue, with no rules
	//ContentTypeOctetStream ContentType = "application/octet-stream"
)

func ContentTypeHeader(headers http.Header) string {
	contentType := headers.Get("Content-Type")

	switch contentType {
	case "":
		// if this is not set (on a delete request for example)
		// still need to encode a type for the client to understand
		return ContentTypeJSON
	default:
		return contentType
	}
}
