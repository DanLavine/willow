package v1

import (
	"fmt"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Lock is the full representation of the Lock Object
type Lock struct {
	// LockID to identify the lock
	LockID string

	// SessionID associated with the client currently holding the lock. Used to heartbeat or release a lock
	SessionID string

	// KeyValues collection that deffines the lock
	KeyValues datatypes.KeyValues

	// Timeout for the lock
	Timeout time.Duration
	// Time until the lock expires if no heartbeats are received
	TimeTillExipre time.Duration

	// LocksHeldOrWaiting show how many clients are all trying to obtaiin the same lock
	LocksHeldOrWaiting uint64
}

//	RETURNS:
//	- error - error describing any possible issues with the Lock and the steps to rectify them
//
// Validate ensures the Lock has all required fields set
func (lock *Lock) Validate() error {
	if lock.SessionID == "" {
		return fmt.Errorf("'SessionID' is required, but received an empty string")
	}

	if err := lock.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return err
	}

	if lock.Timeout == 0 {
		return fmt.Errorf("'Timeout' is required, but received an empty string")
	}

	if lock.LocksHeldOrWaiting == 0 {
		return fmt.Errorf("'Timeout' is required, but received 0")
	}

	return nil
}
