package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// used with list
type CounterResponse struct {
	KeyValues datatypes.KeyValues

	Counters uint64
}

type CountersResponse []CounterResponse

func (cr CountersResponse) ToBytes() []byte {
	data, _ := json.Marshal(cr)
	return data
}

// used with increment and decrement
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

func (c Counter) ToBytes() []byte {
	data, _ := json.Marshal(c)
	return data
}

type CounterSet struct {
	// Specific key values to add or remove a counter from
	KeyValues datatypes.KeyValues

	// specify the specific value to set
	Count uint64
}

func ParseCounterSetRequest(reader io.ReadCloser) (*CounterSet, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := CounterSet{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return &obj, nil
}

func (cs CounterSet) Validate() *api.Error {
	if err := cs.KeyValues.Validate(); err != nil {
		return api.InvalidRequestBody.With("Key values to be valid", err.Error())
	}

	return nil
}

func (cs CounterSet) ToBytes() []byte {
	data, _ := json.Marshal(cs)
	return data
}
