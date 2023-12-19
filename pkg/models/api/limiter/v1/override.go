package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Override can be thought of as a "sub query" for which the rule resides. Any request that matches all the
// given tags for an override will use the new override value. If multiple overrides match a particular set of tags,
// then the override with the lowest value will be used
type Override struct {
	// name for the override. Must be unique for all overrides attached to a rule
	Name string

	// When checking a rule, if it has these exact keys, then the limit will be applied.
	// In the case of an override matchin many key values, the smallest Limit will be enforced
	KeyValues datatypes.KeyValues

	// The new limit to use for the paricular mapping
	Limit uint64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (override *Override) Validate() error {
	if override.Name == "" {
		return fmt.Errorf("'Name' is the empty string")
	}

	if len(override.KeyValues) == 0 {
		return fmt.Errorf("'KeyValues' requres at least 1 key + value pair")
	}

	if err := override.KeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (override *Override) EncodeJSON() []byte {
	data, _ := json.Marshal(override)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream. Valida values [application/json]
//	- reader - stream to read the encoded CreateLockResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// Decode can easily parse the response body from an http create request
func (override *Override) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, override); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return override.Validate()
}

type Overrides struct {
	Overrides []*Override
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (overrides *Overrides) Validate() error {
	if len(overrides.Overrides) == 0 {
		return nil
	}

	for index, override := range overrides.Overrides {
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
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (overrides *Overrides) EncodeJSON() []byte {
	data, _ := json.Marshal(overrides)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream. Valida values [application/json]
//	- reader - stream to read the encoded CreateLockResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// Decode can easily parse the response body from an http create request
func (overrides *Overrides) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, overrides); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return overrides.Validate()
}
