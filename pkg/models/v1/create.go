package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/config"
)

type Create struct {
	// Name of the broker object
	Name string

	// max size of the dead letter queue
	// Cannot be set to  0
	QueueMaxSize uint64

	// Number of times to retry a queue item before sending it to the dead letter queue
	ItemRetryAttempts uint64

	// Max Number of items to keep in the dead letter queue. If full,
	// any new items will just be dropped untill the queue is cleared by an admin.
	// To not have a dead letter queue, this must be set to 0
	DeadLetterQueueMaxSize *uint64
}

func ParseCreateRequest(reader io.ReadCloser) (*Create, *Error) {
	createRequestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	return ParseCreateBody(createRequestBody)
}

func ParseCreateBody(b []byte) (*Create, *Error) {
	var create *Create
	if err := json.Unmarshal(b, create); err != nil {
		return nil, ParseRequestBodyError.With("", err.Error())
	}

	return create, nil
}

func (c *Create) SetDefaults(queueConfig *config.QueueConfig) *Error {
	// check max size
	if c.QueueMaxSize > queueConfig.MaxSize {
		return (&Error{Message: "Error: QueueParams.QueueMaxSize is larger than allowed max", StatusCode: http.StatusBadRequest}).With(fmt.Sprintf("requested %d", c.QueueMaxSize), fmt.Sprintf("to be less than max allowed %d", queueConfig.MaxSize))
	}

	// set default
	if c.QueueMaxSize == 0 {
		c.QueueMaxSize = queueConfig.DefaultSize
	}

	if c.DeadLetterQueueMaxSize != nil {
		if *c.DeadLetterQueueMaxSize > queueConfig.DeadLetterMaxSize {
			return (&Error{Message: "Error: QueueParams.DeadLetterQueueMaxSize is larger than allowed max", StatusCode: http.StatusBadRequest}).With(fmt.Sprintf("requested %d", c.DeadLetterQueueMaxSize), fmt.Sprintf("to be less than max allowed %d", queueConfig.DeadLetterMaxSize))
		}
	}

	return nil
}

func (c *Create) ToBytes() ([]byte, *Error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, MarshelModelFailed.With("", err.Error())
	}

	return data, nil
}