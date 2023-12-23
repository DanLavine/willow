package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Override can be thought of as a "sub query" for a Rule's KeyValues. Any request that matches all the
// given tags for an override will use the new override value. If multiple overrides match a particular set of tags,
// then each override will be validated for their KeyValue group.
type Override struct {
	// Name for the Override. Must be unique for all overrides attached to a rule
	Name string

	// When checking a rule, if it has these exact keys, then the limit will be applied.
	// In the case of an override matchin many key values, all Overrides will be checked
	// unless the Limit is 0. In this case all Overrides can be ignored as the request
	// is guranteed to reject adding the counter
	KeyValues datatypes.KeyValues

	// The new limit to use for the paricular KeyValues
	Limit uint64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Override has all required fields set
func (override *Override) Validate() error {
	if override.Name == "" {
		return errorNameIsInvalid
	}

	if len(override.KeyValues) == 0 {
		return errors.KeyValuesLenghtInvalid
	}

	if err := override.KeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Override
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (override *Override) EncodeJSON() ([]byte, error) {
	return json.Marshal(override)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Override from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (override *Override) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, override); err != nil {
		return err
	}

	return nil
}
