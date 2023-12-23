package v1

import (
	"encoding/json"
	"fmt"
)

// Collection of Rules that is returned from a "Match" or "Query"
type Rules []*Rule

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the Rules has all required fields set
func (rules *Rules) Validate() error {
	if len(*rules) == 0 {
		return nil
	}

	for index, rule := range *rules {
		if rule == nil {
			return fmt.Errorf("invalid Rule at index: %d: rule cannot be nil", index)
		}

		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid Rule at index %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Rules
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (rules *Rules) EncodeJSON() ([]byte, error) {
	return json.Marshal(rules)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Rules from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (rules *Rules) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, rules); err != nil {
		return err
	}

	return nil
}
