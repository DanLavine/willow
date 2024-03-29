package errors

import (
	"encoding/json"
	"fmt"
)

// Error can be used to parse errors returned from services if the request was invalid
type Error struct {
	Message string
}

//	RETURNS:
//	- error - error describing any possible issues with the Error message and the steps to rectify them
//
// Validate ensures the Error has all required fields set
func (err *Error) Validate() error {
	if err.Message == "" {
		return fmt.Errorf("malformed error, there is no 'Message'")
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Error
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (err *Error) EncodeJSON() ([]byte, error) {
	return json.Marshal(err)
}

//	PARAMETERS
//	- data - encoded JSON data to parse the Error from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// DecodeJSON can easily parse the response body from an http create request into the object
func (e *Error) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, e); err != nil {
		return &Error{Message: fmt.Sprintf("failed to decode api error response: %s", err.Error())}
	}

	return nil
}

//	RETURNS:
//	- string - error message
//
// Error returns the original message sent
func (e *Error) Error() string {
	return e.Message
}
