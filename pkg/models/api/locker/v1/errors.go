package v1

import "fmt"

var (
	sessionIDEmpty              = fmt.Errorf("'SessionID' is the empty string")
	keyValuesLenghtInvalid      = fmt.Errorf("'KeyValues' is empty, but requires a length of at least 1")
	timeoutIsInvalid            = fmt.Errorf("'Timeout' is set to 0 and should be set to a reasonable time duration (in nano seconds)")
	locksHeldOrWaitingIsInvalid = fmt.Errorf("'LocksHeldOrWaiting' is set to 0. There is a server error as this lock should be deleted")
)
