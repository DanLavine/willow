package errors

import (
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api"
)

var (
	// errors when creating inital dirs for saving state files
	FailedToCreateDir = &api.Error{Message: "Failed to create dir.", StatusCode: http.StatusInternalServerError}
	FailedToStatDir   = &api.Error{Message: "Failed to stat dir.", StatusCode: http.StatusInternalServerError}
	PathAlreadyExists = &api.Error{Message: "Path already exists and is not a directory.", StatusCode: http.StatusInternalServerError}

	// errors for actual files
	FileNotFound       = &api.Error{Message: "File not found.", StatusCode: http.StatusInternalServerError}
	FailedToCreateFile = &api.Error{Message: "Failed to create file.", StatusCode: http.StatusInternalServerError}

	// encode errors
	WriteFailed  = &api.Error{Message: "Failed to write data to disk.", StatusCode: http.StatusInternalServerError}
	EncodeFailed = &api.Error{Message: "Failed to encode data.", StatusCode: http.StatusInternalServerError}

	// decode errors
	ReadFailed   = &api.Error{Message: "Failed to read data from disk.", StatusCode: http.StatusInternalServerError}
	DecodeFailed = &api.Error{Message: "Failed to decode data from disk.", StatusCode: http.StatusInternalServerError}

	// cleanup errors
	TruncateError = &api.Error{Message: "Failed to truncate error.", StatusCode: http.StatusInternalServerError}
)
