package v1

import (
	"encoding/json"
	"io"
	"net/http"
)

type ACK struct {
	// common broker info
	BrokerInfo

	// ID of the original message being acknowledged
	ID uint64

	// Indicate a success or failure of the message
	Passed bool
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

	if ack.ID == 0 {
		return &Error{Message: "ID cannot be 0", StatusCode: http.StatusBadRequest}
	}

	return nil
}
