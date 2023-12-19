package servererrors

import (
	"encoding/json"
	"fmt"
)

type ApiError struct {
	Message  string
	expected string
	actual   string

	StatusCode int `json:"-"` // ignore this as part of the response. This is to set server side as the response code
}

// With is used to generate a more specific message for any of the provided default Error messages.
// if Expected or Actual is anything other than than the empty string, the message will be formated
// for calls to 'Error()' will the additional details:
//
// [Original Error Message]. Expected [expected string provided]. Actual [actual string provided].
//
//	PARAMETERS:
//	- expected - Optional string that will be used to generate a more tailored message
//	- actual - Optional string that will be used to generate a more tailored message
//
//	RETURNS:
//	- *Error - New Error with all the strings encapsulated so calls to `Error()` are properly formatted
func (e *ApiError) With(expected, actual string) *ApiError {
	newErr := &ApiError{
		Message:    e.Message,
		expected:   expected,
		actual:     actual,
		StatusCode: e.StatusCode,
	}

	return newErr
}

// Error returns a formatted message with optinal 'expected' and 'actual' values
//
//	RETURNS:
//	- string - error message
func (e *ApiError) Error() string {
	err := e.Message

	if e.expected != "" {
		err = fmt.Sprintf("%s Expected: %s.", err, e.expected)
	}

	if e.actual != "" {
		err = fmt.Sprintf("%s Actual: %s.", err, e.actual)
	}

	return err
}

// ToBytes converts a model to a []bytes that can be written and parsed over http successfully
//
//	RETURNS:
//	- []byte - byte array for the model that can be converted through standard json pacakge.
func (e *ApiError) EncodeJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}
