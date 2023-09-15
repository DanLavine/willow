package v1willow

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/query"
)

type EnqueueItemRequest struct {
	// common broker info
	BrokerInfo

	// Message body that will be used by clients receiving this message
	Data []byte

	// If the message should be updatable
	// If set to true:
	//   1. Will colapse on the previous message if it has not been processed and is also updateable
	// If set to false:
	//   1. Will enque the messge as unique and won't be collapsed on. Can still collapse the previous message iff that was true
	Updateable bool
}

func ParseEnqueueItemRequest(reader io.ReadCloser) (*EnqueueItemRequest, *api.Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}
	defer reader.Close()

	enqueueItem := &EnqueueItemRequest{}
	if err := json.Unmarshal(body, enqueueItem); err != nil {
		return nil, api.ParseRequestBodyError.With("enqueue query to be valid json", err.Error())
	}

	if err := enqueueItem.Validate(); err != nil {
		return nil, err
	}

	return enqueueItem, nil
}

func (eir *EnqueueItemRequest) Validate() *api.Error {
	if eir.BrokerInfo.Name == "" {
		return api.InvalidRequestBody.With("Name to be provided", "Name is the empty string")
	}

	if err := eir.BrokerInfo.validate(); err != nil {
		return err
	}

	return nil
}

type DequeueItemRequest struct {
	// common broker info
	Name string

	// query for what readeers to select
	Selection query.Select
}

func ParseDequeueItemRequest(reader io.ReadCloser) (*DequeueItemRequest, *api.Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}
	defer reader.Close()

	dequeueItemRequest := &DequeueItemRequest{}
	if err := json.Unmarshal(body, dequeueItemRequest); err != nil {
		return nil, api.ParseRequestBodyError.With("dequeue query to be valid json", err.Error())
	}

	if err := dequeueItemRequest.Validate(); err != nil {
		return nil, err
	}

	return dequeueItemRequest, nil
}

func (dir *DequeueItemRequest) Validate() *api.Error {
	if dir.Name == "" {
		return api.InvalidRequestBody.With("Name to be provided", "Name is the empty string")
	}

	if err := dir.Selection.Validate(); err != nil {
		return api.InvalidRequestBody.With("", err.Error())
	}

	return nil
}

type DequeueItemResponse struct {
	// common broker info
	BrokerInfo BrokerInfo

	// ID of the message that can be ACKed
	ID string

	// Message body that will be used by clients receiving this message
	Data []byte
}

func (dqr *DequeueItemResponse) ToBytes() ([]byte, *api.Error) {
	data, err := json.Marshal(dqr)
	if err != nil {
		return nil, api.MarshelModelFailed.With("", err.Error())
	}

	return data, nil
}
