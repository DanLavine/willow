package v1

import (
	"encoding/json"
	"io"
	"net/http"
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

func ParseACKRequest(reader io.ReadCloser) (*ACK, *Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	obj := &ACK{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (ack *ACK) Validate() *Error {
	if err := ack.BrokerInfo.validate(); err != nil {
		return err
	}

	if ack.ID == "" {
		return &Error{Message: "ID cannot be an empty string", StatusCode: http.StatusBadRequest}
	}

	return nil
}
