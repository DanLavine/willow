package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Overrides []Override

func (ovr Overrides) ToBytes() []byte {
	data, _ := json.Marshal(ovr)
	return data
}

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

// Server side call to parse the override request to know if it is valid
func ParseOverrideRequest(reader io.ReadCloser) (*Override, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := Override{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return &obj, nil
}

func (or Override) Validate() *api.Error {
	if or.Name == "" {
		return api.InvalidRequestBody.With("Name to be provided", "received empty sting")
	}

	if err := or.KeyValues.Validate(); err != nil {
		return api.InvalidRequestBody.With("KeyValues to be valid", err.Error())
	}

	return nil
}

func (or Override) ToBytes() []byte {
	data, _ := json.Marshal(or)
	return data
}
