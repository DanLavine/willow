package v1

import (
	"encoding/json"
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type OverridesToInclude string

const (
	None  OverridesToInclude = ""
	All   OverridesToInclude = "all"
	Match OverridesToInclude = "match"
)

// Rule query is used to find any Rules that match the provided KeyValues
type RuleQuery struct {
	// Find any rules that match the provided key values. If this is nil, then OverrideQuery can be None or All.
	// If this has a value, then OverrideQuery can be Match
	KeyValues *datatypes.KeyValues

	// type of overrrides to query for
	// 1. Empty string - returns nothing
	// 2. All - returns all overrides
	// 3. Match - returns all override that match the KeyValues if there are any
	OverridesToInclude OverridesToInclude
}

// Used to validate on the server side that all parameters are valid. Client's can also call this
// validation beforehand to ensure that the request is valid before sending
func (rq RuleQuery) Validate() *errors.Error {
	if rq.KeyValues != nil {
		if err := rq.KeyValues.Validate(); err != nil {
			return errors.InvalidRequestBody.With("KeyValues to be valid", err.Error())
		}
	}

	switch rq.OverridesToInclude {
	case None, All, Match:
		// these are all valid
	default:
		return errors.InvalidRequestBody.With(fmt.Sprintf("OverrideQuery is %s", string(rq.OverridesToInclude)), "OverridesToInclude to be one of ['' | all | match]")
	}

	return nil
}

func (rq RuleQuery) ToBytes() []byte {
	data, _ := json.Marshal(rq)
	return data
}
