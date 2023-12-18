package servererrors

import "net/http"

var (
	// queue errors
	MaxEnqueuedItems = &ApiError{Message: "Queue has reached max allowed enqueued items.", StatusCode: http.StatusConflict}
	QueueClosed      = &ApiError{Message: "Queue is closed.", StatusCode: http.StatusConflict}
	QueueNotFound    = &ApiError{Message: "Queue not found.", StatusCode: http.StatusNotAcceptable}
)
