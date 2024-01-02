package v1

import (
	"encoding/json"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
)

// RuleGet is used to get a Rule and optional override
type RuleGet struct {
	// overrrides to match for
	OverridesToMatch *v1common.MatchQuery
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (ruleGet *RuleGet) Validate() error {
	if ruleGet.OverridesToMatch != nil {
		if err := ruleGet.OverridesToMatch.Validate(); err != nil {
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
func (ruleGet *RuleGet) EncodeJSON() ([]byte, error) {
	return json.Marshal(ruleGet)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse RuleQuery from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (ruleGet *RuleGet) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, ruleGet); err != nil {
		return err
	}

	return nil
}
