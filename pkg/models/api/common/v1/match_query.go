package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Match Query is used when permutations of the KeyValues need to be accounted for
type MatchQuery struct {
	// Optional KeyValues are used with the 'permutations' if these are provided.
	// Otherwise, treated as a "select all"
	KeyValues *datatypes.KeyValues

	// #TODO: Order BY + Pagination options
}

//	RETURNS:
//	- error - error describing any possible issues with the query and the steps to rectify them
//
// Validate ensures the CreateLockRequest has all required fields set
func (query *MatchQuery) Validate() error {
	if query.KeyValues != nil {
		if err := query.KeyValues.Validate(); err != nil {
			return err
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the AssociatedQuery
//
// EncodeJSON encodes the model to a valid JSON format
func (query *MatchQuery) EncodeJSON() ([]byte, error) {
	return json.Marshal(query)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse AssociatedQuery from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (query *MatchQuery) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, query); err != nil {
		return err
	}

	return nil
}
