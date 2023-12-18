package errors

import (
	"encoding/json"
	"fmt"
	"io"
)

var (
	// errors when sending a message
	RequestEncodeFailed = &Error{Message: "Failed to encode request."}
	InvalidRequestBody  = &Error{Message: "Request is invalid."}

	// client side errors
	ReadResponseBodyError  = &Error{Message: "Failed to read server response body."}
	ParseResponseBodyError = &Error{Message: "Failed to parse server response."}
)

type Error struct {
	Message string

	expected string
	actual   string
}

// ParseError is used to easily parse the response from a HTTP response body
//
//	PARAMETERS:
//	- reader - stream that contains the encoded message in JSON
//
//	RETURNS:
//	- *Error - will be the error responded by the server, or an error parsing the server response
func ParseError(reader io.ReadCloser) *Error {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return ReadResponseBodyError.With("", err.Error())
	}

	obj := &Error{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return ParseResponseBodyError.With("", err.Error())
	}

	return obj
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
func (e *Error) With(expected, actual string) *Error {
	newErr := &Error{
		Message:  e.Message,
		expected: expected,
		actual:   actual,
	}

	return newErr
}

// Error returns a formatted message with optinal 'expected' and 'actual' values
//
//	RETURNS:
//	- string - error message
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
