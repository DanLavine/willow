package errors

import (
	"net/http"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var (
	// queue errors
	QueueNotFound = &v1.Error{Message: "Queue not found.", StatusCode: http.StatusBadRequest}
	NoReaders     = &v1.Error{Message: "No readers received.", StatusCode: http.StatusInternalServerError}
	NullReader    = &v1.Error{Message: "Null reader received.", StatusCode: http.StatusInternalServerError}

	// queue item errors
	ItemNotfound            = &v1.Error{Message: "Item not found.", StatusCode: http.StatusNotFound}
	MessageTypeNotSupported = &v1.Error{Message: "Message type not supported", StatusCode: http.StatusBadRequest}
)
