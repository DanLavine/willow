package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
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

func (eir *EnqueueItemRequest) Validate() *errors.Error {
	if eir.BrokerInfo.Name == "" {
		return errors.InvalidRequestBody.With("Name to be provided", "Name is the empty string")
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
	Query datatypes.AssociatedKeyValuesQuery
}

func (dir *DequeueItemRequest) Validate() *errors.Error {
	if dir.Name == "" {
		return errors.InvalidRequestBody.With("Name to be provided", "Name is the empty string")
	}

	if err := dir.Query.Validate(); err != nil {
		return errors.InvalidRequestBody.With("", err.Error())
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

func (dqr *DequeueItemResponse) ToBytes() []byte {
	data, _ := json.Marshal(dqr)
	return data
}
