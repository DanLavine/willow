package v1willow

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/config"
	"github.com/DanLavine/willow/pkg/models/api"
)

type Create struct {
	// Name of the broker object
	Name string

	// max size of the dead letter queue
	// Cannot be set to  0
	QueueMaxSize uint64

	// Max Number of items to keep in the dead letter queue. If full,
	// any new items will just be dropped untill the queue is cleared by an admin.
	DeadLetterQueueMaxSize uint64
}

func ParseCreateRequest(reader io.ReadCloser) (*Create, *api.Error) {
	createRequestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	create := &Create{}
	if err := json.Unmarshal(createRequestBody, create); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	return create, nil
}

func (c *Create) SetDefaults(queueConfig *config.QueueConfig) *api.Error {
	// check max size
	if c.QueueMaxSize > *queueConfig.MaxSize {
		return api.InvalidRequestBody.With("QueueMaxSize is larger than allowed max", fmt.Sprintf("allowed max: %d", queueConfig.MaxSize))
	}

	// set default
	if c.QueueMaxSize == 0 {
		return api.InvalidRequestBody.With("QueueMaxSize is set to 0. The queue won't allow any items", "to be at least 1")
	}

	if c.DeadLetterQueueMaxSize > *queueConfig.DeadLetterMaxSize {
		return api.InvalidRequestBody.With("QueueParams.DeadLetterQueueMaxSize is larger than allowed max", fmt.Sprintf("max allowed %d", queueConfig.DeadLetterMaxSize))
	}

	return nil
}

func (c *Create) ToBytes() ([]byte, *api.Error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, api.MarshelModelFailed.With("", err.Error())
	}

	return data, nil
}
