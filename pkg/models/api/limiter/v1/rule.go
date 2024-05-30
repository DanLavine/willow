package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Single Rule that is returned from a "Match" or "Query"
type Rule struct {
	// Name of the Rule
	Name string `json:"Name"`

	// GroupBy contains the logical "key" grouping of any KeyValues for the counters.
	GroupByKeyValues datatypes.KeyValues `json:"GroupByKeyValues"`

	// Limit dictates what value of grouped counter KeyValues to allow untill a limit is reached.
	// Setting this value to -1 means unlimited
	Limit int64 `json:"Limit"`

	// Overrides for the Rule that matched the lookup parameters. These are a read-only
	// parameter that are returned on the FindLimits calls
	Overrides Overrides `json:"Overrides,omitempty"`
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the Rule has all required fields set
func (rule *Rule) Validate() *errors.ModelError {
	if rule.Name == "" {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("is the empty string")}
	}

	if err := rule.GroupByKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return &errors.ModelError{Field: "GroupByKeyValues", Child: err}
	}

	if rule.Limit < -1 {
		return &errors.ModelError{Field: "Limit", Err: fmt.Errorf("is set below the minimum value of -1. Value must be [-1 (ulimited) | 0+ (zero or more specific limit)")}
	}

	return nil
}
