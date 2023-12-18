package handlers

import (
	"encoding/json"
	"io"
	"time"

	servererrors "github.com/DanLavine/willow/internal/server_errors"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
)

// Parse the request to create a Lock object
func ParseLockRequest(reader io.ReadCloser, defaultTimeout time.Duration) (*v1locker.CreateLockRequest, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1locker.CreateLockRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if obj.Timeout == 0 {
		obj.Timeout = defaultTimeout
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", validateErr.Error())
	}

	return obj, nil
}
