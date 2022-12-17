package errors

import (
	"fmt"
	"net/http"
)

var (
	Unimplemented = &Error{Message: "Unimplemented.", StatusCode: http.StatusNotImplemented}

	QueueNotFound = &Error{Message: "Queue not found"}

	// queue errors
	RetrieveError = &Error{Message: "no message retrieved", StatusCode: http.StatusNoContent}

	// general errors
	ValidationError           = &Error{Message: "Invalid request received.", StatusCode: http.StatusBadRequest}
	ProtocolNotSupportedError = &Error{Message: "Prtocol not supported.", StatusCode: http.StatusBadRequest}
)

type Error struct {
	Message  string
	expected string
	actual   string

	StatusCode int
}

func (e *Error) Expected(expected string) *Error {
	newErr := e.duplicate()
	newErr.expected = expected

	return newErr
}

func (e *Error) Actual(actual string) *Error {
	newErr := e.duplicate()
	newErr.actual = actual

	return newErr
}

func (e *Error) Error() string {
	err := e.Message

	if e.expected != "" {
		err = fmt.Sprintf("%s Expected '%s'.", err, e.expected)
	}

	if e.actual != "" {
		err = fmt.Sprintf("%s Actual '%s'.", err, e.actual)
	}

	return err
}

func (e *Error) duplicate() *Error {
	return &Error{
		Message:    e.Message,
		expected:   e.expected,
		actual:     e.actual,
		StatusCode: e.StatusCode,
	}
}
