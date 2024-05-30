package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// CountersQueryResponse show any locks that match an AssociatedQuery.
type Counters []*Counter

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the CountersQueryResponse has all required fields set
func (counters Counters) Validate() *errors.ModelError {
	if len(counters) == 0 {
		return nil
	}

	for index, counter := range counters {
		if counter == nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Err: fmt.Errorf("child cannot be null")}
		}

		if err := counter.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Child: err}
		}
	}

	return nil
}
