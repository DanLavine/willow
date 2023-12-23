package v1

import "fmt"

var (
	errorNameIsInvalid        = fmt.Errorf("'Name' is the empty string")
	errorGroupByInvalidLength = fmt.Errorf("'GroupBy' tags requres at least 1 Key")
	errorGroupByInvalidKeys   = fmt.Errorf("'GroupBy' has duplicate keys")
)
