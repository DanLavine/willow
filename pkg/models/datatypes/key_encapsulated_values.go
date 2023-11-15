package datatypes

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type KeyValues map[string]EncapsulatedData

func (kv KeyValues) Keys() []string {
	keys := []string{}

	for key, _ := range kv {
		keys = append(keys, key)
	}

	return keys
}

func (kv KeyValues) SoretedKeys() []string {
	keys := []string{}

	for key := range kv {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func (kv KeyValues) StripKey(removeKey string) KeyValues {
	returnMap := KeyValues{}

	for key, value := range kv {
		if key != removeKey {
			returnMap[key] = value
		}
	}

	return returnMap
}

func (kv KeyValues) MarshalJSON() ([]byte, error) {

	if err := kv.Validate(); err != nil {
		return nil, err
	}

	// setup parsable string representation of data
	mapKeyValues := map[string]EncapsulatedData{}

	for key, value := range kv {
		switch value.DataType {
		case T_uint8:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatUint(uint64(value.Value.(uint8)), 10),
			}
		case T_uint16:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatUint(uint64(value.Value.(uint16)), 10),
			}
		case T_uint32:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatUint(uint64(value.Value.(uint32)), 10),
			}
		case T_uint64:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatUint(uint64(value.Value.(uint64)), 10),
			}
		case T_uint:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatUint(uint64(value.Value.(uint)), 10),
			}
		case T_int8:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatInt(int64(value.Value.(int8)), 10),
			}
		case T_int16:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatInt(int64(value.Value.(int16)), 10),
			}
		case T_int32:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatInt(int64(value.Value.(int32)), 10),
			}
		case T_int64:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatInt(int64(value.Value.(int64)), 10),
			}
		case T_int:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatInt(int64(value.Value.(int)), 10),
			}
		case T_float32:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatFloat(float64(value.Value.(float32)), 'E', -1, 32),
			}
		case T_float64:
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    strconv.FormatFloat(float64(value.Value.(float64)), 'E', -1, 64),
			}
		case T_string:
			mapKeyValues[key] = value
		case T_nil:
			fmt.Println("setting value:", value)
			mapKeyValues[key] = EncapsulatedData{
				DataType: value.DataType,
				Value:    nil,
			}
		}
	}

	return json.Marshal(mapKeyValues)
}

func (kv *KeyValues) UnmarshalJSON(b []byte) error {
	// parse the original request into a middle object
	custom := map[string]EncapsulatedData{}
	if err := json.Unmarshal(b, &custom); err != nil {
		return err
	}

	if len(custom) == 0 {
		return nil
	}

	if *kv == nil {
		*kv = KeyValues{}
	}

	for key, value := range custom {
		switch value.DataType {
		case T_uint8:
			parsedValue, err := strconv.ParseUint(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint8(uint8(parsedValue))
		case T_uint16:
			parsedValue, err := strconv.ParseUint(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint16(uint16(parsedValue))
		case T_uint32:
			parsedValue, err := strconv.ParseUint(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint32(uint32(parsedValue))
		case T_uint64:
			parsedValue, err := strconv.ParseUint(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint64(uint64(parsedValue))
		case T_uint:
			parsedValue, err := strconv.ParseUint(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint(uint(parsedValue))
		case T_int8:
			parsedValue, err := strconv.ParseInt(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int8(int8(parsedValue))
		case T_int16:
			parsedValue, err := strconv.ParseInt(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int16(int16(parsedValue))
		case T_int32:
			parsedValue, err := strconv.ParseInt(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int32(int32(parsedValue))
		case T_int64:
			parsedValue, err := strconv.ParseInt(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int64(int64(parsedValue))
		case T_int:
			parsedValue, err := strconv.ParseInt(value.Value.(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int(int(parsedValue))
		case T_float32:
			parsedValue, err := strconv.ParseFloat(value.Value.(string), 32)
			if err != nil {
				return err
			}
			(*kv)[key] = Float32(float32(parsedValue))
		case T_float64:
			parsedValue, err := strconv.ParseFloat(value.Value.(string), 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Float64(float64(parsedValue))
		case T_string:
			(*kv)[key] = value
		case T_nil:
			if value.Value != nil {
				return fmt.Errorf("key %s requires a nil value", key)
			}
			(*kv)[key] = Nil()
		default:
			return fmt.Errorf("unkown data type: %d", value.DataType)
		}
	}

	return nil
}

func (kv KeyValues) Validate() error {
	for key, value := range kv {
		if err := value.Validate(); err != nil {
			return fmt.Errorf("key %s error: %w", key, err)
		}
	}

	return nil
}

// TODO: think i will need something like this to check the GroupedBy[] in the lmiter requests.
//
//	but will also need to GenerateTagPairs
func (kv KeyValues) ExistenceQuery() AssociatedKeyValuesQuery {
	query := AssociatedKeyValuesQuery{}

	return query
}

// GenerateGroupPairs can be used to go through a list of strings and create all unique ordered groupings.
// The returned slice is sorted on length of key value pairs with the longest value always last
//
// Example:
// * Tags{"b":"2", "a":"1", "c":"3"} -> [{"a":"1"}, {"b":"2"}, {"c":"3"}, {"a":"1", "b":"2"}, {"a":"1", "c":"3"}, {"b":"2", "c":"3"}, {"a":"1", "b":"2", "c":"3"}]
func (kv KeyValues) GenerateTagPairs() []KeyValues {
	groupPairs := kv.generateTagPairs(kv.Keys())

	sort.SliceStable(groupPairs, func(i, j int) bool {
		if len(groupPairs[i]) < len(groupPairs[j]) {
			return true
		}

		if len(groupPairs[i]) == len(groupPairs[j]) {
			sortedKeysI := groupPairs[i].SoretedKeys()
			sortedKeysJ := groupPairs[j].SoretedKeys()

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
