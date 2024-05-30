package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// LockClaim is used to refresh a heartbeat and eventually reclaim a lock when a service restarts
type LockClaim struct {
	// SessionID for the currently held lock
	SessionID string
}

//	RETURNS:
//	- *errors.ModelError - error describing any possible issues and the steps to rectify them
//
// Validate ensures the LockClaim is valid for the held lock
func (claim *LockClaim) Validate() *errors.ModelError {
	if claim.SessionID == "" {
		return &errors.ModelError{Field: "SessionID", Err: fmt.Errorf("received an empty string")}
	}

	return nil
}
