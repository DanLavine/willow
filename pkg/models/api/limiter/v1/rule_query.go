package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type OverrideQuery string

const (
	None  OverrideQuery = ""
	All   OverrideQuery = "all"
	Match OverrideQuery = "match"
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
	OverrideQuery OverrideQuery
}

// Server side logic to parse a Rule to know it is valid
func ParseRuleQuery(reader io.ReadCloser) (*RuleQuery, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := RuleQuery{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return &obj, nil
}

// Used to validate on the server side that all parameters are valid. Client's can also call this
// validation beforehand to ensure that the request is valid before sending
func (rq RuleQuery) Validate() *api.Error {
	if rq.KeyValues != nil {
		if err := rq.KeyValues.Validate(); err != nil {
			return api.InvalidRequestBody.With("KeyValues to be valid", err.Error())
		}
	}

	switch rq.OverrideQuery {
	case None, All, Match:
		// these are all valid
	default:
		return api.InvalidRequestBody.With(fmt.Sprintf("OverrideQuery is %s", string(rq.OverrideQuery)), "OverrideQuery to be one of ['' | all | match]")
	}

	return nil
}

func (rq RuleQuery) ToBytes() []byte {
	data, _ := json.Marshal(rq)
	return data
}
