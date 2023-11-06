package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	MarshelModelFailed = &Error{Message: "Failed to encode response.", StatusCode: http.StatusInternalServerError}

	NotAcceptable         = &Error{Message: "Invalid request.", StatusCode: http.StatusNotAcceptable}
	InvalidRequestBody    = &Error{Message: "Invalid request body.", StatusCode: http.StatusBadRequest}
	ReadRequestBodyError  = &Error{Message: "Failed to read request body.", StatusCode: http.StatusInternalServerError}
	ParseRequestBodyError = &Error{Message: "Failed to parse request body.", StatusCode: http.StatusBadRequest}
)

type Error struct {
	Message  string `json:"Error"`
	expected string
	actual   string

	StatusCode int `json:"-"` // ignore this as part of the response
}

func (e *Error) With(expected, actual string) *Error {
	newErr := &Error{
		Message:    e.Message,
		expected:   expected,
		actual:     actual,
		StatusCode: e.StatusCode,
	}

	return newErr
}

func (e *Error) Error() string {
	err := e.Message

	if e.expected != "" {
		err = fmt.Sprintf("%s Expected: %s.", err, e.expected)
	}

	if e.actual != "" {
		err = fmt.Sprintf("%s Actual: %s.", err, e.actual)
	}

	return err
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"Error": "%s"}`, e.Error())), nil
}

func (e *Error) ToBytes() []byte {
	data, _ := json.Marshal(e)
	return data
}

func ParseError(reader io.ReadCloser) (*Error, error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	obj := &Error{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
