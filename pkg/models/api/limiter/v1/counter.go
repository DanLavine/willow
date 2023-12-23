package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Counter is the full api model that is returned as part of the query operations
type Counter struct {
	// KeyValues that define the counter
	KeyValues datatypes.KeyValues

	// Total number of counters for the particular KeyValues
	Counters int64
}

//	RETURNS:
//	- error - error describing any possible issues with the Counter and the steps to rectify them
//
// Validate ensures the Lock has all required fields set
func (counter *Counter) Validate() error {
	if len(counter.KeyValues) == 0 {
		return errors.KeyValuesLenghtInvalid
	}

	if err := counter.KeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Counter
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (counter *Counter) EncodeJSON() ([]byte, error) {
	return json.Marshal(counter)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Counter from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (counter *Counter) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, counter); err != nil {
		return err
	}

	return nil
}
