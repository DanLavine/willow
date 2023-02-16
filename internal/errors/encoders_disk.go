package errors

import (
	"net/http"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

var (
	// errors when creating inital dirs for saving state files
	FailedToCreateDir = &v1.Error{Message: "Failed to create dir.", StatusCode: http.StatusInternalServerError}
	FailedToStatDir   = &v1.Error{Message: "Failed to stat dir.", StatusCode: http.StatusInternalServerError}
	PathAlreadyExists = &v1.Error{Message: "Path already exists and is not a directory.", StatusCode: http.StatusInternalServerError}

	// errors for actual files
	FileNotFound       = &v1.Error{Message: "File not found.", StatusCode: http.StatusInternalServerError}
	FailedToCreateFile = &v1.Error{Message: "Failed to create file.", StatusCode: http.StatusInternalServerError}

	// encode errors
	WriteFailed  = &v1.Error{Message: "Failed to write data to disk.", StatusCode: http.StatusInternalServerError}
	EncodeFailed = &v1.Error{Message: "Failed to encode data.", StatusCode: http.StatusInternalServerError}

	// decode errors
	ReadFailed   = &v1.Error{Message: "Failed to read data from disk.", StatusCode: http.StatusInternalServerError}
	DecodeFailed = &v1.Error{Message: "Failed to decode data from disk.", StatusCode: http.StatusInternalServerError}

	// cleanup errors
	TruncateError = &v1.Error{Message: "Failed to truncate error.", StatusCode: http.StatusInternalServerError}
)
