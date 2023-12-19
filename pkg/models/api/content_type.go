package api

import "net/http"

type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"

	// Add this eventual to encode/decode bytes directly witch will be faster.
	// useful when will is a high traffic queue, with no rules
	//ContentTypeOctetStream ContentType = "application/octet-stream"

	ContentTypeInvalid ContentType = "invalid"
)

func ContentHeaderFromRequest(r *http.Request) string {
	return r.Header.Get("Content-Type")
}

func ContentTypeFromRequest(r *http.Request) ContentType {
	switch ContentHeaderFromRequest(r) {
	case "application/json":
		return ContentTypeJSON
	default:
		return ContentTypeInvalid
	}
}

func ContentHeaderFromResponse(r *http.Response) string {
	return r.Header.Get("Content-Type")
}

func ContentTypeFromResponse(r *http.Response) ContentType {
	switch ContentHeaderFromResponse(r) {
	case "application/json":
		return ContentTypeJSON
	default:
		return ContentTypeInvalid
	}
}
