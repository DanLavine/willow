package v1

import (
	"encoding/json"
	"fmt"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
)

// RuleMatch is used to find any Rules that match the provided KeyValues
type RuleMatch struct {
	// Find any rules that match the provided MatchQuery.
	RulesToMatch *v1common.MatchQuery

	// Overrrides to match for
	OverridesToMatch *v1common.MatchQuery
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (ruleMatch *RuleMatch) Validate() error {
	if ruleMatch.RulesToMatch == nil {
		return fmt.Errorf("'RulesToMatch' cannot be nil")
	}

	if err := ruleMatch.RulesToMatch.Validate(); err != nil {
		return err
	}

	// validate the overrides match
	if ruleMatch.OverridesToMatch != nil {
		if err := ruleMatch.OverridesToMatch.Validate(); err != nil {
			return err
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (ruleMatch *RuleMatch) EncodeJSON() ([]byte, error) {
	return json.Marshal(ruleMatch)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse RuleQuery from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (ruleMatch *RuleMatch) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, ruleMatch); err != nil {
		return err
	}

	return nil
}
