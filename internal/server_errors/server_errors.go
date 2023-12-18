package servererrors

import (
	"net/http"
)

var (
	// Server state errors
	ServerShutdown = &ApiError{Message: "Server is shutting down.", StatusCode: http.StatusServiceUnavailable}

	// Storage errors
	UnknownStorageType = &ApiError{Message: "Unkown storage type.", StatusCode: http.StatusInternalServerError}

	// unexpectd errors
	InternalServerError = &ApiError{Message: "Internal Server Error.", StatusCode: http.StatusInternalServerError}
)
