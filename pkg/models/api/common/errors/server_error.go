package errors

import (
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
func (err *ServerError) Validate() *ModelError {
	if err.StatusCode == 0 {
		return &ModelError{Field: "StatusCode", Err: fmt.Errorf("is not set")}
	}

	if err.Message == "" {
		return &ModelError{Field: "Message", Err: fmt.Errorf("is set tot the empty strinf")}
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
