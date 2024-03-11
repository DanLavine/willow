package errors

import "fmt"

var (
	// Common error when the KeyValues for an API mode has a len of 0
	KeyValuesLenghtInvalid = fmt.Errorf("'KeyValues' is empty, but requires a length of at least 1")
)
