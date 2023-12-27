package v1

import (
	"encoding/json"
	"fmt"
)

// Single Rule that is returned from a "Match" or "Query"
type Rule struct {
	// Name of the Rule
	Name string

	// GroupBy contains the logical "key" grouping of any KeyValues for the counters.
	GroupBy []string

	// Limit dictates what value of grouped counter KeyValues to allow untill a limit is reached.
	// Setting this value to -1 means unlimited
	Limit int64

	// Overrides for the Rule that matched the lookup parameters
	Overrides []Override
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the Rule has all required fields set
func (rule *Rule) Validate() error {
	if rule.Name == "" {
		return errorNameIsInvalid
	}

	if len(rule.GroupBy) == 0 {
		return errorGroupByInvalidKeys
	}

	if rule.Limit < -1 {
		return fmt.Errorf("'Limit' is set below the minimum value of -1. Value must be [-1 (ulimited) | 0+ (zero or more specific limit) ]")
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Rule
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (rule *Rule) EncodeJSON() ([]byte, error) {
	return json.Marshal(rule)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse a Rule from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (rule *Rule) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, rule); err != nil {
		return err
	}

	return nil
}
