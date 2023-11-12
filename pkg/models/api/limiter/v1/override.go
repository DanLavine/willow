package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Override struct {
	// name for the override. Must be unique for all overrides attached to a rule
	Name string

	// They key value parings we are making the override for
	// NOTE: these must match the GroupBy keys in the original Rule
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

	obj := &Override{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.ValidateRequest(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (or *Override) ValidateRequest() *api.Error {
	if or.Name == "" {
		return api.InvalidRequestBody.With("Name to be provided", "received empty sting")
	}

	if len(or.KeyValues) == 0 {
		return api.InvalidRequestBody.With("KeyValues tags to be provided", "received empty set")
	}

	if err := or.KeyValues.Validate(); err != nil {
		return api.InvalidRequestBody.With("KeyValues to be valid", err.Error())
	}

	return nil
}

func (or *Override) ToBytes() []byte {
	data, _ := json.Marshal(or)
	return data
}
