package api

import (
	"io"
)

// Required functions for Service response object
type APIResponseObject interface {
	EncodeJSON() []byte
}

// All pkg models need to suport the encoding types
type APIObject interface {
	Validate() error

	EncodeJSON() []byte

	Decode(contentType ContentType, reader io.ReadCloser) error
}
