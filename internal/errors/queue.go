package errors

import (
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api"
)

var (
	// queue setup errors
	UnknownQueueStorage = &api.Error{Message: "Unkown queue storage type.", StatusCode: http.StatusBadRequest}
	NoCreateParams      = &api.Error{Message: "No Create Params.", StatusCode: http.StatusInternalServerError}
	NoEncoder           = &api.Error{Message: "No encoder received.", StatusCode: http.StatusInternalServerError}
	NoReaders           = &api.Error{Message: "No readers received.", StatusCode: http.StatusInternalServerError}
	NilReader           = &api.Error{Message: "Null reader received.", StatusCode: http.StatusInternalServerError}

	// queue errors
	QueueNotFound = &api.Error{Message: "Queue not found.", StatusCode: http.StatusNotAcceptable}

	// queue item errors
	ItemNotfound            = &api.Error{Message: "Item not found.", StatusCode: http.StatusNotFound}
	ItemNotProcessing       = &api.Error{Message: "Item not processing.", StatusCode: http.StatusBadRequest}
	MessageTypeNotSupported = &api.Error{Message: "Message type not supported", StatusCode: http.StatusBadRequest}
	MaxEnqueuedItems        = &api.Error{Message: "Queue reached max number of in flight items. Cannot queue request", StatusCode: http.StatusTooManyRequests}
)
