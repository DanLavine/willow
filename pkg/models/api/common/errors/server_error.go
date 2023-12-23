package errors

import (
	"encoding/json"
	"fmt"
)

// Error can be used to parse errors returned from services if the request was invalid
type ServerError struct {
	StatusCode int `json:"-"` // ignore this field in the json response since it is set in a header

	Message string
}

//	RETURNS:
//	- error - error describing any possible issues with the ServerError and the steps to rectify them
//
// Validate ensures the ServerError has all required fields set
func (err *ServerError) Validate() error {
	if err.StatusCode == 0 {
		return fmt.Errorf("malformed error, there is no 'StatusCode' set")
	}

	if err.Message == "" {
		return fmt.Errorf("malformed error, there is no 'Message'")
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Lock
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (err *ServerError) EncodeJSON() ([]byte, error) {
	return json.Marshal(err)
}

//	PARAMETERS
//	- data - encoded JSON data to parse the Error from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// DecodeJSON can easily parse the response body from an http create request into the object
func (e *ServerError) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, e); err != nil {
		return fmt.Errorf("failed to decode error from the server: %w", err)
	}

	return nil
}

//	RETURNS:
//	- string - error message
//
// Error returns the original message
func (e *ServerError) Error() string {
	return e.Message
}
