package v1

import (
	"encoding/json"
	"io"
	"sort"
)

type Create struct {
	// type of broker we want to create
	BrokerType BrokerType

	// Tag for the broker
	BrokerTags []string

	// params for queue
	QueueParams QueueParams

	// Leve this empty for no dead letter queue to be configured
	DeadLetterQueueParams *DeadLetterQueueParams
}

type QueueParams struct {
	// max size of the dead letter queue
	// Default: 100 if set at 0
	MaxSize uint64

	// Number of times to retry a queue item before sending it to the dead letter queue
	RetryCount uint64
}

type DeadLetterQueueParams struct {
	// Max Number of items to keep in the dead letter queue. If full, any new items will just be dropped
	// untill the queue is cleared by an admin.
	MaxSize uint64
}

func ParseCreateRequest(reader io.ReadCloser) (Create, *Error) {
	createRequestBody, err := io.ReadAll(reader)
	if err != nil {
		return Create{}, InvalidRequestBody.With("", err.Error())
	}

	return ParseCreateQuery(createRequestBody)
}

func ParseCreateQuery(b []byte) (Create, *Error) {
	var create Create
	if err := json.Unmarshal(b, &create); err != nil {
		return create, ParseRequestBodyError.With("", err.Error())
	}

	// always sort tags
	sort.Strings(create.BrokerTags)

	// setup defaults
	if create.QueueParams.MaxSize == 0 {
		create.QueueParams.MaxSize = 100
	}

	return create, create.Validate()
}

func (c Create) Validate() *Error {
	if c.DeadLetterQueueParams != nil {
		if c.DeadLetterQueueParams.MaxSize == 0 {
			return InvalidRequestBody.With("DeadLetterQueueParams.MaxSize to be greater than 0", "0")
		}
	}

	return nil
}
