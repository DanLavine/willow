package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Override can be thought of as a "sub query" for a Rule's KeyValues. Any request that matches all the
// given tags for an override will use the new override value. If multiple overrides match a particular set of tags,
// then each override will be validated for their KeyValue group.
type Override struct {
	// Name for the Override. Must be unique for all overrides attached to a rule
	Name string

	// When checking a rule, if it has these exact keys, then the limit will be applied.
	// In the case of an override matchin many key values, all Overrides will be checked
	// unless the Limit is 0. In this case all Overrides can be ignored as the request
	// is guranteed to reject adding the counter
	KeyValues datatypes.KeyValues

	// The new limit to use for the paricular KeyValues
	// Setting this value to -1 means unlimited
	Limit int64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Override has all required fields set
func (override *Override) Validate() *errors.ModelError {
	if override.Name == "" {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("is the empty string")}
	}

	if len(override.KeyValues) == 0 {
		return &errors.ModelError{Field: "KeyValue", Err: fmt.Errorf("needs at least a length of at least 1")}
	}

	if err := override.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}
