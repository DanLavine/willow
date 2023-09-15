package errors

import (
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api"
)

var (
	// No Dead Letter queue Configured
	DeadLetterQueueNotConfigured = &api.Error{Message: "No dead letter queue configured.", StatusCode: http.StatusNotAcceptable}

	// dead letter setup queue errors
	DeadLetterQueueInvalidCreateParams = &api.Error{Message: "Invalid create params.", StatusCode: http.StatusBadRequest}
	DeadLetterQueueNilEncoder          = &api.Error{Message: "No encoder received.", StatusCode: http.StatusInternalServerError}

	// dead letter queue is full
	DeadLetterQueueFull = &api.Error{Message: "Dead Letter Queue is full.", StatusCode: http.StatusBadRequest}

	// dead letter queue item errors
	DeadLetterItemNotfound = &api.Error{Message: "Dead Letter Item not found.", StatusCode: http.StatusNotFound}
)
