package errors

import (
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
func (err *Error) Validate() *ModelError {
	if err.Message == "" {
		return &ModelError{Field: "Message", Err: fmt.Errorf("no error message was received")}
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
