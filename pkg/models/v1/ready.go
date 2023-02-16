package v1

import (
	"encoding/json"
	"io"
)

// Used to notify Willow that a client is ready to accept a message for a particular broker
type Ready struct {
	Name string
}

func ParseReadyRequest(reader io.ReadCloser) (*Ready, *Error) {
	readyRequestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	return ParseReadyBody(readyRequestBody)
}

func ParseReadyBody(b []byte) (*Ready, *Error) {
	readyQuery := &Ready{}
	if err := json.Unmarshal(b, readyQuery); err != nil {
		return nil, ParseRequestBodyError.With("ready query to be valid json", err.Error())
	}

	return readyQuery, nil
}
