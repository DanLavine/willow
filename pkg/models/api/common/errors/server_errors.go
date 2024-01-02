package errors

import (
	"fmt"
	"net/http"
)

var (
	// Server state errors
	ServerShutdown = &ServerError{Message: "Server is shutting down", StatusCode: http.StatusServiceUnavailable}

	// Storage errors
	UnknownStorageType = &ServerError{Message: "Unkown storage type", StatusCode: http.StatusInternalServerError}

	// unexpectd errors
	InternalServerError = &ServerError{Message: "Internal Server Error", StatusCode: http.StatusInternalServerError}
)

func ServerErrorNoAPIModel() *ServerError {
	return &ServerError{
		Message:    "unable to decode api model",
		StatusCode: http.StatusInternalServerError,
	}
}

func ServerErrorReadingRequestBody(err error) *ServerError {
	return &ServerError{
		Message:    fmt.Sprintf("failed to read http request body: %s", err.Error()),
		StatusCode: http.StatusInternalServerError,
	}
}

func ServerErrorDecodeingJson(err error) *ServerError {
	return &ServerError{
		Message:    fmt.Sprintf("failed to decode request: %s", err.Error()),
		StatusCode: http.StatusBadRequest,
	}
}

func ServerErrorModelRequestValidation(err error) *ServerError {
	return &ServerError{
		Message:    fmt.Sprintf("failed validation: %s", err.Error()),
		StatusCode: http.StatusBadRequest,
	}
}

func ServerErrorModelResponseValidation(err error) *ServerError {
	return &ServerError{
		Message:    fmt.Sprintf("failed valdating response object: %s", err.Error()),
		StatusCode: http.StatusBadRequest,
	}
}

// UnknownContentType is used if an API request has an unkown `Content-Type` header
func ServerUnknownContentType(contentType string) *ServerError {
	return &ServerError{
		Message:    fmt.Sprintf("server recieved unkown content type: '%s'", contentType),
		StatusCode: http.StatusBadRequest,
	}
}
