package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Overrides []Override

func (ovr Overrides) ToBytes() []byte {
	data, _ := json.Marshal(ovr)
	return data
}

// Override can be thought of as a "sub query" for which the rule resides. Any request that matches all the
// given tags for an override will use the new override value. If multiple overrides match a particular set of tags,
// then the override with the lowest value will be used
type Override struct {
	// name for the override. Must be unique for all overrides attached to a rule
	Name string

	// When checking a rule, if it has these exact keys, then the limit will be applied.
	// In the case of an override matchin many key values, the smallest Limit will be enforced
	KeyValues datatypes.KeyValues

	// The new limit to use for the paricular mapping
	Limit uint64
}

func (or Override) Validate() *errors.Error {
	if or.Name == "" {
		return errors.InvalidRequestBody.With("Name to be provided", "received empty sting")
	}

	if err := or.KeyValues.Validate(); err != nil {
		return errors.InvalidRequestBody.With("KeyValues to be valid", err.Error())
	}

	return nil
}

func (or Override) ToBytes() []byte {
	data, _ := json.Marshal(or)
	return data
}
