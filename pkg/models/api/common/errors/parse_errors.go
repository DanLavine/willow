package errors

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api"
)

// UnknownContentType is used if an API request has an unkown `Content-Type` header
func UnknownContentType(contentType api.ContentType) error {
	return fmt.Errorf("unkown content type '%s'", contentType)
}

// FailedToReadStreamBody is used when reading the body of an http request fails
func FailedToReadStreamBody(err error) error {
	return fmt.Errorf("failed to read stream body: %w", err)
}

// FailedToDecodeBody is used when decoding the Stream's data fails
func FailedToDecodeBody(err error) error {
	return fmt.Errorf("failed to decode stream body: %w", err)
}

// wrapper to know if an error happend client side
func ClientError(err error) error {
	return fmt.Errorf("client error: %w", err)
}

// wrapper for server side logic to encode an error back to the clients
func ServerError(err error) *Error {
	return &Error{Message: fmt.Errorf("server error: %w", err).Error()}
}
