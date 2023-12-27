package v1

import (
	"encoding/json"
	"fmt"
)

// RuleCreateRequest contains all the properties needed when creating a new Limiter rule
type RuleCreateRequest struct {
	// Name of the Rule to create. Must be unique for all Rules
	Name string

	// GroupBy contains the logical "key" grouping of any KeyValues for the counters.
	GroupBy []string

	// Limit dictates what value of grouped counter KeyValues to allow untill a limit is reached
	// Setting this value to -1 means unlimited
	Limit int64
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the RuleCreateRequest has all required fields set
func (req *RuleCreateRequest) Validate() error {
	if req.Name == "" {
		return errorNameIsInvalid
	}

	if len(req.GroupBy) == 0 {
		return errorGroupByInvalidLength
	}

	// ensure each key is unique
	seenKeys := map[string]struct{}{}
	for _, key := range req.GroupBy {
		if _, ok := seenKeys[key]; ok {
			return fmt.Errorf("%w: %s", errorGroupByInvalidKeys, key)
		}

		seenKeys[key] = struct{}{}
	}

	// ensure the limit is set properly
	if req.Limit < -1 {
		return fmt.Errorf("'Limit' is set below the minimum value of -1. Value must be [-1 (ulimited) | 0+ (zero or more specific limit) ]")
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the RuleCreateRequest
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (req *RuleCreateRequest) EncodeJSON() ([]byte, error) {
	return json.Marshal(req)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse RuleCreateRequest from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (req *RuleCreateRequest) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, req); err != nil {
		return err
	}

	return nil
}
