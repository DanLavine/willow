package v1

import (
	"fmt"
)

// Collection of Rules that is returned from a "Match" or "Query"
type Rules []*Rule

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the Rules has all required fields set
func (rules Rules) Validate() error {
	if len(rules) == 0 {
		return nil
	}

	for index, rule := range rules {
		if rule == nil {
			return fmt.Errorf("invalid Rule at index %d. Rule can not be nil", index)
		}

		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid Rule at index %d: %w", index, err)
		}
	}

	return nil
}
