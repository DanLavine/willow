package errors

import (
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api"
)

var (
	ServerShutdown = &api.Error{Message: "Server is shutting down.", StatusCode: http.StatusServiceUnavailable}
	QueueClosed    = &api.Error{Message: "Queue has been closed.", StatusCode: http.StatusGone}

	ProcessNotSet = &api.Error{Message: "Process not set?", StatusCode: http.StatusInternalServerError}

	InternalServerError = &api.Error{Message: "Internal Server Error.", StatusCode: http.StatusInternalServerError}
)
