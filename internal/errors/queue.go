package errors

import (
	"net/http"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var (
	// queue setup errors
	UnknownQueueStorage = &v1.Error{Message: "Unkown queue storage type.", StatusCode: http.StatusBadRequest}
	NoCreateParams      = &v1.Error{Message: "No Create Params.", StatusCode: http.StatusInternalServerError}
	NoEncoder           = &v1.Error{Message: "No encoder received.", StatusCode: http.StatusInternalServerError}
	NoReaders           = &v1.Error{Message: "No readers received.", StatusCode: http.StatusInternalServerError}
	NilReader           = &v1.Error{Message: "Null reader received.", StatusCode: http.StatusInternalServerError}

	// queue errors
	QueueNotFound = &v1.Error{Message: "Queue not found.", StatusCode: http.StatusBadRequest}

	// queue item errors
	ItemNotfound            = &v1.Error{Message: "Item not found.", StatusCode: http.StatusNotFound}
	ItemNotProcessing       = &v1.Error{Message: "Item not processing.", StatusCode: http.StatusBadRequest}
	MessageTypeNotSupported = &v1.Error{Message: "Message type not supported", StatusCode: http.StatusBadRequest}
)
