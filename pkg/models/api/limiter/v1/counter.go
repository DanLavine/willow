package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Counter struct {
	// Specific key values to add or remove a counter from
	KeyValues datatypes.KeyValues
}

func ParseCounterRequest(reader io.ReadCloser) (*Counter, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := Counter{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return &obj, nil
}

func (c Counter) Validate() *api.Error {
	if err := c.KeyValues.Validate(); err != nil {
		return api.InvalidRequestBody.With("Key values to be valid", err.Error())
	}

	return nil
}
