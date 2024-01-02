package v1

import (
	"encoding/json"
)

// OverrideUpdate is used to update a particular override
type OverrideUpdate struct {
	// The new limit to use for the paricular Override
	// Setting this value to -1 means unlimited
	Limit int64
}

//	RETURNS:
//	- error - any errors encountered with the OverrideUpdate
//
// Validate is used to ensure that Override has all required fields set
func (overrideUpdate *OverrideUpdate) Validate() error {
	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the OverrideUpdate
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (overrideUpdate *OverrideUpdate) EncodeJSON() ([]byte, error) {
	return json.Marshal(overrideUpdate)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse OverrideUpdate from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (overrideUpdate *OverrideUpdate) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, overrideUpdate); err != nil {
		return err
	}

	return nil
}
