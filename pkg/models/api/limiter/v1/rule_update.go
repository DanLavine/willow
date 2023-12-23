package v1

import (
	"encoding/json"
)

// RuleUpdateRquest defines the fields of a Rule that can be updated
type RuleUpdateRquest struct {
	// Limit for a particual Rule
	Limit uint64
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the RuleUpdateRquest has all required fields set
func (req *RuleUpdateRquest) Validate() error {
	// no-op atm
	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the RuleUpdateRquest
//	- error - error enoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (req *RuleUpdateRquest) EncodeJSON() ([]byte, error) {
	return json.Marshal(req)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse RuleUpdate from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (req *RuleUpdateRquest) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, req); err != nil {
		return err
	}

	return nil
}
