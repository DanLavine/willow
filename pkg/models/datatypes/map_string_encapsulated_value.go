package datatypes

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type KeyValuesErr struct {
	err error
}

func (kve *KeyValuesErr) Error() string {
	return kve.err.Error()
}

type KeyValues map[string]EncapsulatedValue

//	RETURNS:
//	- []byte - encoded JSON byte array for the Override
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (kv *KeyValues) EncodeJSON() ([]byte, error) {
	return json.Marshal(kv)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Override from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (kv *KeyValues) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, kv); err != nil {
		return err
	}

	return nil
}

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
	// always be safe and perform validation
	if err := kv.Validate(); err != nil {
		return nil, err
	}

	// setup parsable string representation of data
	mapKeyValues := map[string]EncapsulatedValue{}

	for key, value := range kv {
		switch value.DataType() {
		case T_uint8:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_uint8,
				Data: strconv.FormatUint(uint64(value.Value().(uint8)), 10),
			}
		case T_uint16:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_uint16,
				Data: strconv.FormatUint(uint64(value.Value().(uint16)), 10),
			}
		case T_uint32:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_uint32,
				Data: strconv.FormatUint(uint64(value.Value().(uint32)), 10),
			}
		case T_uint64:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_uint64,
				Data: strconv.FormatUint(uint64(value.Value().(uint64)), 10),
			}
		case T_uint:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_uint,
				Data: strconv.FormatUint(uint64(value.Value().(uint)), 10),
			}
		case T_int8:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_int8,
				Data: strconv.FormatInt(int64(value.Value().(int8)), 10),
			}
		case T_int16:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_int16,
				Data: strconv.FormatInt(int64(value.Value().(int16)), 10),
			}
		case T_int32:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_int32,
				Data: strconv.FormatInt(int64(value.Value().(int32)), 10),
			}
		case T_int64:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_int64,
				Data: strconv.FormatInt(int64(value.Value().(int64)), 10),
			}
		case T_int:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_int,
				Data: strconv.FormatInt(int64(value.Value().(int)), 10),
			}
		case T_float32:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_float32,
				Data: strconv.FormatFloat(float64(value.Value().(float32)), 'E', -1, 32),
			}
		case T_float64:
			mapKeyValues[key] = EncapsulatedValue{
				Type: T_float64,
				Data: strconv.FormatFloat(float64(value.Value().(float64)), 'E', -1, 64),
			}
		case T_string:
			mapKeyValues[key] = value
		default:
			return nil, fmt.Errorf("unknown data type: %d", value.DataType())
		}
	}

	return json.Marshal(mapKeyValues)
}

func (kv *KeyValues) UnmarshalJSON(b []byte) error {
	// parse the original request into a middle object
	custom := map[string]EncapsulatedValue{}
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
		if value.Value() == nil {
			return fmt.Errorf("received empty Data for key '%s'", key)
		}

		switch value.DataType() {
		case T_uint8:
			parsedValue, err := strconv.ParseUint(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint8(uint8(parsedValue))
		case T_uint16:
			parsedValue, err := strconv.ParseUint(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint16(uint16(parsedValue))
		case T_uint32:
			parsedValue, err := strconv.ParseUint(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint32(uint32(parsedValue))
		case T_uint64:
			parsedValue, err := strconv.ParseUint(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint64(uint64(parsedValue))
		case T_uint:
			parsedValue, err := strconv.ParseUint(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Uint(uint(parsedValue))
		case T_int8:
			parsedValue, err := strconv.ParseInt(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int8(int8(parsedValue))
		case T_int16:
			parsedValue, err := strconv.ParseInt(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int16(int16(parsedValue))
		case T_int32:
			parsedValue, err := strconv.ParseInt(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int32(int32(parsedValue))
		case T_int64:
			parsedValue, err := strconv.ParseInt(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int64(int64(parsedValue))
		case T_int:
			parsedValue, err := strconv.ParseInt(value.Value().(string), 10, 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Int(int(parsedValue))
		case T_float32:
			parsedValue, err := strconv.ParseFloat(value.Value().(string), 32)
			if err != nil {
				return err
			}
			(*kv)[key] = Float32(float32(parsedValue))
		case T_float64:
			parsedValue, err := strconv.ParseFloat(value.Value().(string), 64)
			if err != nil {
				return err
			}
			(*kv)[key] = Float64(float64(parsedValue))
		case T_string:
			(*kv)[key] = value
		default:
			return fmt.Errorf("key '%s' has an unkown data type: %d", key, value.DataType())
		}
	}

	return nil
}

func (kv KeyValues) Validate() error {
	if len(kv) == 0 {
		return &KeyValuesErr{err: fmt.Errorf("KeyValues cannot be empty")}
	}

	for key, value := range kv {
		if err := value.Validate(); err != nil {
			return &KeyValuesErr{err: fmt.Errorf("key '%s' error: %w", key, err)}
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
