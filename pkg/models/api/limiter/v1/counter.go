package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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

func (c Counter) Validate() *errors.Error {
	if err := c.KeyValues.Validate(); err != nil {
		return errors.InvalidRequestBody.With("KeyValues to be valid", err.Error())
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

func (cs CounterSet) Validate() *errors.Error {
	if err := cs.KeyValues.Validate(); err != nil {
		return errors.InvalidRequestBody.With("Key values to be valid", err.Error())
	}

	return nil
}

func (cs CounterSet) ToBytes() []byte {
	data, _ := json.Marshal(cs)
	return data
}
