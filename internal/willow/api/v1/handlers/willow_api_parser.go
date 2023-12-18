package handlers

import (
	"encoding/json"
	"io"

	servererrors "github.com/DanLavine/willow/internal/server_errors"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

// Parse create queue request
func ParseCreateRequest(reader io.ReadCloser) (*v1willow.Create, *servererrors.ApiError) {
	createRequestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	create := &v1willow.Create{}
	if err := json.Unmarshal(createRequestBody, create); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	return create, nil
}

// func (c *Create) SetDefaults(queueConfig *config.QueueConfig) *errors.Error {
// 	// check max size
// 	if c.QueueMaxSize > *queueConfig.MaxSize {
// 		return errors.InvalidRequestBody.With("QueueMaxSize is larger than allowed max", fmt.Sprintf("allowed max: %d", queueConfig.MaxSize))
// 	}

// 	// set default
// 	if c.QueueMaxSize == 0 {
// 		return errors.InvalidRequestBody.With("QueueMaxSize is set to 0. The queue won't allow any items", "to be at least 1")
// 	}

// 	if c.DeadLetterQueueMaxSize > *queueConfig.DeadLetterMaxSize {
// 		return errors.InvalidRequestBody.With("QueueParams.DeadLetterQueueMaxSize is larger than allowed max", fmt.Sprintf("max allowed %d", queueConfig.DeadLetterMaxSize))
// 	}

// 	return nil
// }

// Enqueue item request parser
func ParseEnqueueItemRequest(reader io.ReadCloser) (*v1willow.EnqueueItemRequest, *servererrors.ApiError) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}
	defer reader.Close()

	enqueueItem := &v1willow.EnqueueItemRequest{}
	if err := json.Unmarshal(body, enqueueItem); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("enqueue query to be valid json", err.Error())
	}

	if err := enqueueItem.Validate(); err != nil {
		return nil, servererrors.InvalidRequestBody.With("", err.Error())
	}

	return enqueueItem, nil
}

// Dequeue item request parser
func ParseDequeueItemRequest(reader io.ReadCloser) (*v1willow.DequeueItemRequest, *servererrors.ApiError) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}
	defer reader.Close()

	dequeueItemRequest := &v1willow.DequeueItemRequest{}
	if err := json.Unmarshal(body, dequeueItemRequest); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("dequeue query to be valid json", err.Error())
	}

	if validateErr := dequeueItemRequest.Validate(); err != nil {
		return nil, servererrors.InvalidRequestBody.With("", validateErr.Error())
	}

	return dequeueItemRequest, nil
}

// Ack item request parser
func ParseACKRequest(reader io.ReadCloser) (*v1willow.ACK, *servererrors.ApiError) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, servererrors.ReadRequestBodyError.With("", err.Error())
	}

	obj := &v1willow.ACK{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, servererrors.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, servererrors.InvalidRequestBody.With("", validateErr.Error())
	}

	return obj, nil
}
