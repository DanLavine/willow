package v1

import (
	"encoding/json"
	"fmt"
)

// Overrides is used as part of a Query or Match lookup operation
type Overrides []*Override

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (overrides *Overrides) Validate() error {
	if len(*overrides) == 0 {
		return nil
	}

	for index, override := range *overrides {
		if override == nil {
			return fmt.Errorf("error at overreides index: %d: the override is nil", index)
		}

		if err := override.Validate(); err != nil {
			return fmt.Errorf("error at overrides index %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Overrides
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (overrides *Overrides) EncodeJSON() ([]byte, error) {
	return json.Marshal(overrides)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Overrides from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (overrides *Overrides) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, overrides); err != nil {
		return err
	}

	return nil
}
