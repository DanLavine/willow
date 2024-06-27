package v1

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Definition struct {
	// Envapsulated value defines is the actual value assocaited with the key
	EncapsulatedValue datatypes.EncapsulatedValue `json:"EncapsulatedValue"`

	// Unique defines if a particlar key should be treated as unique across all possible combinations
	Unique bool `json:"Unique"`
}

type DBDefinition map[string]Definition

// This type represents all possible EncapsulatedValue types
type AnyKeyValues = DBDefinition

// This type includes all values other than the 'Any' type:
type TypedKeyValues = DBDefinition

func (kv DBDefinition) Keys() []string {
	keys := []string{}

	for key, _ := range kv {
		keys = append(keys, key)
	}

	return keys
}

func (kv DBDefinition) SortedKeys() []string {
	keys := []string{}

	for key := range kv {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func (kv DBDefinition) Validate(minAllowedKeyType, maxAllowedKeyType datatypes.DataType) *errors.ModelError {
	if len(kv) == 0 {
		return &errors.ModelError{Err: fmt.Errorf("recieved no KeyValues, but requires a length of at least 1")}
	}

	for key, value := range kv {
		if err := value.EncapsulatedValue.Validate(minAllowedKeyType, maxAllowedKeyType); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err}
		}
	}

	return nil
}

func (kv *DBDefinition) UnmarshalJSON(b []byte) error {
	custom := map[string]struct {
		EncapsulatedValue struct {
			Data any                `json:"Data"`
			Type datatypes.DataType `json:"Type"`
		} `json:"EncapsulatedValue"`

		Unique bool `json:"Unique"`
	}{}

	if err := json.Unmarshal(b, &custom); err != nil {
		return err
	}

	if *kv == nil {
		*kv = DBDefinition{}
	}

	for key, value := range custom {
		decodedData, err := datatypes.StringDecoding(value.EncapsulatedValue.Type, value.EncapsulatedValue.Data)
		if err != nil {
			return fmt.Errorf("'%s' has an invalid value: %w", key, err)
		}

		// set the actual value if things are proper
		(*kv)[key] = Definition{
			EncapsulatedValue: datatypes.EncapsulatedValue{
				Type: value.EncapsulatedValue.Type,
				Data: decodedData,
			},
			Unique: value.Unique,
		}
	}

	return nil
}
