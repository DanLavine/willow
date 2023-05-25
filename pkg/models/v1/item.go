package v1

import (
	"encoding/json"
	"io"
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

//type DequeueItemRequest struct {
//	// common broker info
//	BrokerInfo BrokerInfo
//
//	// type of requuest we want
//
//}

type DequeueItemResponse struct {
	// common broker info
	BrokerInfo BrokerInfo

	// ID of the message that can be ACKed
	ID uint64

	// Message body that will be used by clients receiving this message
	Data []byte
}

func ParseEnqueueItemRequest(reader io.ReadCloser) (*EnqueueItemRequest, *Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}
	defer reader.Close()

	enqueueItem := &EnqueueItemRequest{}
	if err := json.Unmarshal(body, enqueueItem); err != nil {
		return nil, ParseRequestBodyError.With("enqueue query to be valid json", err.Error())
	}

	if enqueueItem.BrokerInfo.Name == "" {
		return nil, InvalidRequestBody.With("Name to be provided", "Name is the empty string")
	}

	if err := enqueueItem.BrokerInfo.validate(); err != nil {
		return nil, err
	}

	return enqueueItem, nil
}

//func ParseDequeueItemRequest(reader io.ReadCloser) (*DequeueItemRequest, *Error) {
//	body, err := io.ReadAll(reader)
//	if err != nil {
//		return nil, InvalidRequestBody.With("", err.Error())
//	}
//	defer reader.Close()
//
//	dequeueItemRequest := &DequeueItemRequest{}
//	if err := json.Unmarshal(body, dequeueItemRequest); err != nil {
//		return nil, ParseRequestBodyError.With("dequeue query to be valid json", err.Error())
//	}
//
//	if err := dequeueItemRequest.BrokerInfo.validate(); err != nil {
//		return nil, err
//	}
//
//	return dequeueItemRequest, nil
//}

func (dqr *DequeueItemResponse) ToBytes() []byte {
	data, _ := json.Marshal(dqr)
	return data
}
