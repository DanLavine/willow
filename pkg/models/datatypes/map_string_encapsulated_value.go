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
			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err.(*errors.ModelError)}
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

// GenerateGroupPairs can be used to go through a list of strings and create all unique ordered groupings.
// The returned slice is sorted on length of key value pairs with the longest value always last
//
// Example:
// - Tags{"b":"2", "a":"1", "c":"3"} -> [{"a":"1"}, {"b":"2"}, {"c":"3"}, {"a":"1", "b":"2"}, {"a":"1", "c":"3"}, {"b":"2", "c":"3"}, {"a":"1", "b":"2", "c":"3"}]
func (kv KeyValues) GenerateTagPairs() []KeyValues {
	groupPairs := kv.generateTagPairs(kv.Keys())

	sort.SliceStable(groupPairs, func(i, j int) bool {
		if len(groupPairs[i]) < len(groupPairs[j]) {
			return true
		}

		if len(groupPairs[i]) == len(groupPairs[j]) {
			sortedKeysI := groupPairs[i].SortedKeys()
			sortedKeysJ := groupPairs[j].SortedKeys()

			for index, value := range sortedKeysI {
				if value < sortedKeysJ[index] {
					return true
				} else if sortedKeysJ[index] < value {
					return false
				}
			}

			// at this point, all values must be in the proper order
			return true
		}

		return false
	})

	return groupPairs
}

func (kv KeyValues) generateTagPairs(group []string) []KeyValues {
	var allGroupPairs []KeyValues

	switch len(group) {
	case 0:
		// nothing to do here
	case 1:
		// there is only 1 key value pair
		allGroupPairs = append(allGroupPairs, KeyValues{group[0]: kv[group[0]]})
	default:
		// add the first index each time. Will recurse through original group shrinking by 1 each time to capture all elements
		allGroupPairs = append(allGroupPairs, KeyValues{group[0]: kv[group[0]]})

		// drop a key and advance to the next subset ["a", "b", "c"] -> ["b", "c"]
		allGroupPairs = append(allGroupPairs, kv.generateTagPairs(group[1:])...)

		for i := 1; i < len(group); i++ {
			// generate all n[0,1] + n[x,x+1] groupings. I.E [{"a":"1","b":"2"}, {"a":"1","c":"3"}, ....]
			newGrouping := []string{group[0], group[i]}
			allGroupPairs = append(allGroupPairs, kv.generateTagGroups(newGrouping, group[i+1:])...)
		}
	}

	return allGroupPairs
}

func (kv KeyValues) generateTagGroups(prefix, suffix []string) []KeyValues {
	allGroupPairs := []KeyValues{}

	// add initial combined slice
	baseGrouping := KeyValues{}
	for _, prefixKey := range prefix {
		baseGrouping[prefixKey] = kv[prefixKey]
	}
	allGroupPairs = append(allGroupPairs, baseGrouping)

	// recurse building up to n size
	for i := 0; i < len(suffix); i++ {
		allGroupPairs = append(allGroupPairs, kv.generateTagGroups(append(prefix, suffix[i]), suffix[i+1:])...)
	}

	return allGroupPairs
}
