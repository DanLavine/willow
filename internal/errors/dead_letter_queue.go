package errors

import (
	"net/http"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var (
	// No Dead Letter queue Configured
	DeadLetterQueueNotConfigured = &v1.Error{Message: "No dead letter queue configured.", StatusCode: http.StatusNotAcceptable}
	// dead letter setup queue errors
	DeadLetterQueueInvalidCreateParams = &v1.Error{Message: "Invalid create params.", StatusCode: http.StatusBadRequest}
	DeadLetterQueueNilEncoder          = &v1.Error{Message: "No encoder received.", StatusCode: http.StatusInternalServerError}

	// dead letter queue is full
	DeadLetterQueueFull = &v1.Error{Message: "Dead Letter Queue is full.", StatusCode: http.StatusBadRequest}

	// dead letter queue item errors
	DeadLetterItemNotfound = &v1.Error{Message: "Dead Letter Item not found.", StatusCode: http.StatusNotFound}
)
