package v1willow

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

type RequeueLocation uint

const (
	RequeueFront RequeueLocation = iota
	RequeueEnd
	RequeueNone
)

type ACK struct {
	// common broker info
	BrokerInfo

	// ID of the original message being acknowledged
	ID string

	// Indicate a success or failure of the message
	Passed          bool
	RequeueLocation RequeueLocation // only used when set to false
}

func ParseACKRequest(reader io.ReadCloser) (*ACK, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &ACK{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (ack *ACK) Validate() *api.Error {
	if err := ack.BrokerInfo.validate(); err != nil {
		return err
	}

	if ack.ID == "" {
		return api.InvalidRequestBody.With("ID cannot be an empty string", "")
	}

	return nil
}
