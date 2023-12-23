package v1

import (
	"encoding/json"
	"fmt"
)

// Locks show any locks that match an AssociatedQuery.
// In the future, this will be an ADMIN API as the SessionID should be hiden from
// any malicious actors. But for now, it stands a useful debug API to inspect the
// state of the world.
type Locks []*Lock

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the LockQueryResponse has all required fields set
func (locks *Locks) Validate() error {
	if len(*locks) == 0 {
		return nil
	}

	for index, lock := range *locks {
		if lock == nil {
			return fmt.Errorf("invalid Lock at index: %d: lock cannot be nil", index)
		}

		if err := lock.Validate(); err != nil {
			return fmt.Errorf("invalid Lock at index: %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the LockQueryResponse
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (locks *Locks) EncodeJSON() ([]byte, error) {
	return json.Marshal(locks)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Locks from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (locks *Locks) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, locks); err != nil {
		return err
	}

	return nil
}
