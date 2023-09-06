package errors

import (
	"net/http"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var (
	ServerShutdown = &v1.Error{Message: "Server is shutting down.", StatusCode: http.StatusServiceUnavailable}
	QueueClosed    = &v1.Error{Message: "Queue has been closed.", StatusCode: http.StatusGone}

	ProcessNotSet = &v1.Error{Message: "Process not set?", StatusCode: http.StatusInternalServerError}

	InternalServerError = &v1.Error{Message: "Internal Server Error.", StatusCode: http.StatusInternalServerError}
)
