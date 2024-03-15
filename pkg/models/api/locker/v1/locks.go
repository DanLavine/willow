package v1

import (
	"fmt"
)

// Locks show any locks that match an AssociatedQuery.
// In the future, this will be an ADMIN API as the SessionID should be hiden from
// any malicious actors. But for now, it stands a useful debug API to inspect the
// state of the world.
type Locks []*Lock

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the LockQueryResponse has all required fields set
func (locks *Locks) Validate() error {
	if len(*locks) == 0 {
		return nil
	}

	for index, lock := range *locks {
		if lock == nil {
			return fmt.Errorf("invalid Lock at index %d: lock cannot be nil", index)
		}

		if err := lock.Validate(); err != nil {
			return fmt.Errorf("invalid Lock at index %d: %w", index, err)
		}
	}

	return nil
}
