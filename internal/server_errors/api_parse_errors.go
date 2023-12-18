package servererrors

import (
	"net/http"
)

var (
	// server side errors
	InvalidRequestBody    = &ApiError{Message: "Invalid request body.", StatusCode: http.StatusBadRequest}
	ReadRequestBodyError  = &ApiError{Message: "Failed to read request body.", StatusCode: http.StatusBadRequest}
	ParseRequestBodyError = &ApiError{Message: "Failed to parse request.", StatusCode: http.StatusBadRequest}

	// client side errors
	MarshelModelFailed     = &ApiError{Message: "Failed to encode response.", StatusCode: http.StatusInternalServerError}
	ReadResponseBodyError  = &ApiError{Message: "Failed to read server response body."}
	ParseResponseBodyError = &ApiError{Message: "Failed to parse server response."}
)
