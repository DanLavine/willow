package v1

import (
	"fmt"
)

// CountersQueryResponse show any locks that match an AssociatedQuery.
type Counters []*Counter

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the CountersQueryResponse has all required fields set
func (counters Counters) Validate() error {
	if len(counters) == 0 {
		return nil
	}

	for index, counter := range counters {
		if counter == nil {
			return fmt.Errorf("invalid Counter at index %d. Counter can not be nil", index)
		}

		if err := counter.Validate(); err != nil {
			return fmt.Errorf("invalid Counter at index %d: %w", index, err)
		}
	}

	return nil
}
