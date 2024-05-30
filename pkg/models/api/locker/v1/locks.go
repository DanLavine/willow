package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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
func (locks Locks) Validate() *errors.ModelError {
	if len(locks) == 0 {
		return nil
	}

	for index, lock := range locks {
		if lock == nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Err: fmt.Errorf("Lock cannot be null")}
		}

		if err := lock.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Child: err}
		}
	}

	return nil
}
