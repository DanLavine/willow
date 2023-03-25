package errors

import (
	"net/http"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var (
	ServerShutdown = &v1.Error{Message: "Server is shutting down.", StatusCode: http.StatusServiceUnavailable}

	ProcessNotSet = &v1.Error{Message: "Process not set?", StatusCode: http.StatusInternalServerError}

	InternalServerError = &v1.Error{Message: "Internal Server Error", StatusCode: http.StatusInternalServerError}
)
