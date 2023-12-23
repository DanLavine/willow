package v1

import (
	"encoding/json"
	"fmt"

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

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (ruleQuery *RuleQuery) Validate() error {
	if ruleQuery.KeyValues != nil {
		if len(*ruleQuery.KeyValues) == 0 {
			return fmt.Errorf("'KeyValues' requres at least 1 key + value piar")
		}

		if err := ruleQuery.KeyValues.Validate(); err != nil {
			return err
		}
	}

	switch ruleQuery.OverridesToInclude {
	case None, All, Match:
		// these are all valid
	default:
		return fmt.Errorf("'OverrideQuery' is %s, but must be one of ['' | all | match]", string(ruleQuery.OverridesToInclude))
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (ruleQuery *RuleQuery) EncodeJSON() ([]byte, error) {
	return json.Marshal(ruleQuery)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse RuleQuery from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (ruleQuery *RuleQuery) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, ruleQuery); err != nil {
		return err
	}

	return nil
}
