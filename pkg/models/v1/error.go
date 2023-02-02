package v1

import "fmt"

type Error struct {
	Message  string
	expected string
	actual   string

	StatusCode int
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
		err = fmt.Sprintf("%s Expected %s.", err, e.expected)
	}

	if e.actual != "" {
		err = fmt.Sprintf("%s Actual %s.", err, e.actual)
	}

	return err
}
