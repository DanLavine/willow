package v1

import (
	"fmt"
)

// LockClaim is used to refresh a heartbeat and eventually reclaim a lock when a service restarts
type LockClaim struct {
	// SessionID for the currently held lock
	SessionID string
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the LockCreateRequest has all required fields set
func (claim *LockClaim) Validate() error {
	if claim.SessionID == "" {
		return fmt.Errorf("'SessionID' is required, but received an empty string")
	}

	return nil
}
