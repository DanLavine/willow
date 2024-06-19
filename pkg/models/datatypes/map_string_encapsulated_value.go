package datatypes

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type KeyValues map[string]EncapsulatedValue

func (kv KeyValues) Keys() []string {
	keys := []string{}

	for key, _ := range kv {
		keys = append(keys, key)
	}

	return keys
}

func (kv KeyValues) SortedKeys() []string {
	keys := []string{}

	for key := range kv {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func (kv KeyValues) Validate(minAllowedKeyType, maxAllowedKeyType DataType) *errors.ModelError {
	if len(kv) == 0 {
		return &errors.ModelError{Err: fmt.Errorf("recieved no KeyValues, but requires a length of at least 1")}
	}

	for key, value := range kv {
		if err := value.Validate(minAllowedKeyType, maxAllowedKeyType); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err}
		}
	}

	return nil
}

func (kv *KeyValues) UnmarshalJSON(b []byte) error {
	custom := map[string]struct {
		Type DataType `json:"Type"`
		Data any      `json:"Data"`
	}{}

	if err := json.Unmarshal(b, &custom); err != nil {
		return err
	}

	if *kv == nil {
		*kv = KeyValues{}
	}

	for key, value := range custom {
		decodedData, err := StringDecoding(value.Type, value.Data)
		if err != nil {
			return fmt.Errorf("'%s' has an invalid value: %w", key, err)
		}

		// set the actual value if things are proper
		(*kv)[key] = EncapsulatedValue{
			Type: value.Type,
			Data: decodedData,
		}
	}

	return nil
}
