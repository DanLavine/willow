package v1

import (
	"encoding/json"
	"io"
	"sort"
)

type EnqueueItem struct {
	// specific queue name for the message
	// For a "private" queue, this will be needed. Hard to do auh on only "tags"
	Name string

	// Tags for an item. Can be used to update specific item if the previous item has not yet processed
	// OR so the queue pulls the items in a first in, first out order.
	Tags []string

	// Message body that will be used by clients receiving this message
	Data []byte

	// If the message should be updatable
	// If set to true:
	//   1. Will colapse on the previous message if it has not been processed and is also updateable
	// If set to false:
	//   1. Will enque the messge as unique and won't be collapsed on
	Updateable bool
}

type DequeueItemRequest struct {
	// specific queue name for the message
	Name string

	// specific tag that this message was pulled from
	Tags []string
}

type DequeueItemResponse struct {
	// ID of the message that can be ACKed
	ID uint64

	// specific queue name for the message
	Name string

	// specific tag that this message was pulled from
	Tags []string

	// Message body that will be used by clients receiving this message
	Data []byte
}

func ParseEnqueueItem(reader io.ReadCloser) (*EnqueueItem, *Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}
	defer reader.Close()

	enqueueItem := &EnqueueItem{}
	if err := json.Unmarshal(body, enqueueItem); err != nil {
		return nil, ParseRequestBodyError.With("enqueue query to be valid json", err.Error())
	}

	sort.Strings(enqueueItem.Tags)

	return enqueueItem, nil
}

func ParseDequeueItemRequest(reader io.ReadCloser) (*DequeueItemRequest, *Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}
	defer reader.Close()

	dequeueItemRequest := &DequeueItemRequest{}
	if err := json.Unmarshal(body, dequeueItemRequest); err != nil {
		return nil, ParseRequestBodyError.With("dequeue query to be valid json", err.Error())
	}

	sort.Strings(dequeueItemRequest.Tags)

	return dequeueItemRequest, nil
}

func (dqr *DequeueItemResponse) ToBytes() []byte {
	data, _ := json.Marshal(dqr)
	return data
}
