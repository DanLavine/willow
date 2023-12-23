package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// AssociatedQuery is used to query a number of various apis for the
// Willow services. Since most data is saved throught the common KeyValues,
// for Willow's queues, Limiter's Rule Overrides, Limiter's Counters and Locker's Locks.
type AssociatedQuery struct {
	// Query for the KeyValues that defined the various API Models
	AssociatedKeyValues datatypes.AssociatedKeyValuesQuery

	// #TODO: Order BY + Pagination options
}

//	RETURNS:
//	- error - error describing any possible issues with the query and the steps to rectify them
//
// Validate ensures the CreateLockRequest has all required fields set
func (query *AssociatedQuery) Validate() error {
	if err := query.AssociatedKeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the AssociatedQuery
//
// EncodeJSON encodes the model to a valid JSON format
func (query *AssociatedQuery) EncodeJSON() ([]byte, error) {
	return json.Marshal(query)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse AssociatedQuery from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (query *AssociatedQuery) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, query); err != nil {
		return err
	}

	return nil
}
