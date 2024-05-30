package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// Collection of Rules that is returned from a "Match" or "Query"
type Rules []*Rule

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the Rules has all required fields set
func (rules Rules) Validate() *errors.ModelError {
	if len(rules) == 0 {
		return nil
	}

	for index, rule := range rules {
		if rule == nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Err: fmt.Errorf("Rule cannot be null")}
		}

		if err := rule.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Child: err}
		}
	}

	return nil
}
