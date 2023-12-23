package v1

import (
	"encoding/json"
	"fmt"
)

// CountersQueryResponse show any locks that match an AssociatedQuery.
type Counters []*Counter

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the CountersQueryResponse has all required fields set
func (counters *Counters) Validate() error {
	if len(*counters) == 0 {
		return nil
	}

	for index, counter := range *counters {
		if counter == nil {
			return fmt.Errorf("invalid Counter at index: %d: value cannot be nil", index)
		}

		if err := counter.Validate(); err != nil {
			return fmt.Errorf("invalid Counter at index %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the CountersQueryResponse
//	- error - error encodng to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (counters *Counters) EncodeJSON() ([]byte, error) {
	return json.Marshal(counters)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Counters from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (counters *Counters) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, counters); err != nil {
		return err
	}

	return nil
}
