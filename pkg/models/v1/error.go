package v1

import "fmt"

type Error struct {
	Message  string
	expected string
	actual   string

	StatusCode int
}

func (e *Error) Expected(expected string) *Error {
	e.expected = expected
	return e
}

func (e *Error) Actual(actual string) *Error {
	e.actual = actual
	return e
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
