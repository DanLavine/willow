package datatypes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// EncapsulatedValue provides validation for all datatypes from uint8 to string.
// It cacn be used to enforce that a proper single value is provided
type EncapsulatedValue struct {
	// Type of the data that will be recorded and converted for the apis
	Type DataType `json:"Type"`

	// Actual data to use and parse for requests
	Data any `json:"Data"`
}

// Less has special rules that a type provided with T_Any always returns false.
func (ev EncapsulatedValue) Less(comparableObj EncapsulatedValue) bool {
	// current type is less than the compared type
	if ev.Type.Less(comparableObj.Type) {
		return true
	}

	// compared type is greater than the current type
	if comparableObj.Type.Less(ev.Type) {
		return false
	}

	// types must be the same or any
	return ev.lessValue(comparableObj)
}

func (ev EncapsulatedValue) lessValue(comparableObj EncapsulatedValue) bool {
	if ev.Type == T_any || comparableObj.Type == T_any {
		return false
	}

	switch ev.Type {
	case T_uint8:
		return ev.Data.(uint8) < comparableObj.Data.(uint8)
	case T_uint16:
		return ev.Data.(uint16) < comparableObj.Data.(uint16)
	case T_uint32:
		return ev.Data.(uint32) < comparableObj.Data.(uint32)
	case T_uint64:
		return ev.Data.(uint64) < comparableObj.Data.(uint64)
	case T_uint:
		return ev.Data.(uint) < comparableObj.Data.(uint)
	case T_int8:
		return ev.Data.(int8) < comparableObj.Data.(int8)
	case T_int16:
		return ev.Data.(int16) < comparableObj.Data.(int16)
	case T_int32:
		return ev.Data.(int32) < comparableObj.Data.(int32)
	case T_int64:
		return ev.Data.(int64) < comparableObj.Data.(int64)
	case T_int:
		return ev.Data.(int) < comparableObj.Data.(int)
	case T_float32:
		return ev.Data.(float32) < comparableObj.Data.(float32)
	case T_float64:
		return ev.Data.(float64) < comparableObj.Data.(float64)
	case T_string:
		return ev.Data.(string) < comparableObj.Data.(string)
	default:
		panic(fmt.Sprintf("Unexpected type %d", ev.Type))
	}
}

// LessMatchType strictly matches the type between the two values so T_any will always be the greates value
func (ev EncapsulatedValue) LessMatchType(comparableObj EncapsulatedValue) bool {
	// current type is less than the compared type
	if ev.Type.LessMatchType(comparableObj.Type) {
		return true
	}

	// compared type is greater than the current type
	if comparableObj.Type.LessMatchType(ev.Type) {
		return false
	}

	return ev.lessValueMatchTyoe(comparableObj)
}

func (ev EncapsulatedValue) lessValueMatchTyoe(comparableObj EncapsulatedValue) bool {
	switch ev.Type {
	case T_uint8:
		return ev.Data.(uint8) < comparableObj.Data.(uint8)
	case T_uint16:
		return ev.Data.(uint16) < comparableObj.Data.(uint16)
	case T_uint32:
		return ev.Data.(uint32) < comparableObj.Data.(uint32)
	case T_uint64:
		return ev.Data.(uint64) < comparableObj.Data.(uint64)
	case T_uint:
		return ev.Data.(uint) < comparableObj.Data.(uint)
	case T_int8:
		return ev.Data.(int8) < comparableObj.Data.(int8)
	case T_int16:
		return ev.Data.(int16) < comparableObj.Data.(int16)
	case T_int32:
		return ev.Data.(int32) < comparableObj.Data.(int32)
	case T_int64:
		return ev.Data.(int64) < comparableObj.Data.(int64)
	case T_int:
		return ev.Data.(int) < comparableObj.Data.(int)
	case T_float32:
		return ev.Data.(float32) < comparableObj.Data.(float32)
	case T_float64:
		return ev.Data.(float64) < comparableObj.Data.(float64)
	case T_string:
		return ev.Data.(string) < comparableObj.Data.(string)
	case T_any:
		return false
	default:
		panic(fmt.Sprintf("Unexpected type %d", ev.Type))
	}
}

// Marshal a valid object to encoded data. This can panic if the original object has not been validated
func (ev EncapsulatedValue) MarshalJSON() ([]byte, error) {
	custom := struct {
		Type DataType `json:"Type"`
		Data any      `json:"Data"`
	}{
		Type: ev.Type,
		Data: StringEncoding(ev.Type, ev.Data),
	}

	return json.Marshal(custom)
}

// convert a JSON blob into a valid object. Validate should still be called on the parent object
func (ev *EncapsulatedValue) UnmarshalJSON(b []byte) error {
	custom := struct {
		Type DataType `json:"Type"`
		Data any      `json:"Data"`
	}{}

	if err := json.Unmarshal(b, &custom); err != nil {
		return err
	}

	decodedData, err := StringDecoding(custom.Type, custom.Data)
	if err != nil {
		return err
	}

	ev.Type = custom.Type
	ev.Data = decodedData
	return nil
}

// Validate all Encpasulated data types inclusing custom
func (ev EncapsulatedValue) Validate(minAllowedKeyType, maxAllowedKeyType DataType) *errors.ModelError {
	return ValidateTypeAndData(ev.Type, minAllowedKeyType, maxAllowedKeyType, ev.Data)
}

func ValidateTypeAndData(dataType, minAllowedKeyType, maxAllowedKeyType DataType, data any) *errors.ModelError {
	switch {
	case AnyDataType[dataType]:
		// special case for any as we expect the data to be nil
		if data != nil {
			return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'any' requires nil Data")}
		}
	case GeneralDataTypes[dataType]:
		if data == nil {
			return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'%s' has nil Data, but requires a castable string", dataType.ToString())}
		}

		kind := reflect.ValueOf(data).Kind()

		switch dataType {
		case T_uint8:
			if kind != reflect.Uint8 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'uint8' has Data of kind: %s", kind.String())}
			}
		case T_uint16:
			if kind != reflect.Uint16 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'uint16' has Data of kind: %s", kind.String())}
			}
		case T_uint32:
			if kind != reflect.Uint32 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'uint32' has Data of kind: %s", kind.String())}
			}
		case T_uint64:
			if kind != reflect.Uint64 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'uint64' has Data of kind: %s", kind.String())}
			}
		case T_uint:
			if kind != reflect.Uint {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'uint' has Data of kind: %s", kind.String())}
			}
		case T_int8:
			if kind != reflect.Int8 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'int8' has Data of kind: %s", kind.String())}
			}
		case T_int16:
			if kind != reflect.Int16 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'int16' has Data of kind: %s", kind.String())}
			}
		case T_int32:
			if kind != reflect.Int32 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'int32' has Data of kind: %s", kind.String())}
			}
		case T_int64:
			if kind != reflect.Int64 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'int64' has Data of kind: %s", kind.String())}
			}
		case T_int:
			if kind != reflect.Int {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'int' has Data of kind: %s", kind.String())}
			}
		case T_float32:
			if kind != reflect.Float32 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'float32' has Data of kind: %s", kind.String())}
			}
		case T_float64:
			if kind != reflect.Float64 {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'float64' has Data of kind: %s", kind.String())}
			}
		case T_string:
			if kind != reflect.String {
				return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'string' has Data of kind: %s", kind.String())}
			}
		}
	default:
		return &errors.ModelError{Field: "Type", Err: fmt.Errorf("'unknown' has Data '%v'", data)}
	}

	if dataType < minAllowedKeyType || dataType > maxAllowedKeyType {
		return &errors.ModelError{Field: "Type", Err: fmt.Errorf("invalid value '%d'. The required value must be with the data types [%d:%d] inclusively", dataType, minAllowedKeyType, maxAllowedKeyType)}
	}

	return nil
}

// must call validate first if it is applicable
func StringEncoding(dataType DataType, data any) any {
	switch dataType {
	case T_any:
		return nil
	case T_string:
		return data.(string)
	case T_uint:
		return strconv.FormatUint(uint64(data.(uint)), 10)
	case T_uint8:
		return strconv.FormatUint(uint64(data.(uint8)), 10)
	case T_uint16:
		return strconv.FormatUint(uint64(data.(uint16)), 10)
	case T_uint32:
		return strconv.FormatUint(uint64(data.(uint32)), 10)
	case T_uint64:
		return strconv.FormatUint(uint64(data.(uint64)), 10)
	case T_int:
		return strconv.FormatInt(int64(data.(int)), 10)
	case T_int8:
		return strconv.FormatInt(int64(data.(int8)), 10)
	case T_int16:
		return strconv.FormatInt(int64(data.(int16)), 10)
	case T_int32:
		return strconv.FormatInt(int64(data.(int32)), 10)
	case T_int64:
		return strconv.FormatInt(int64(data.(int64)), 10)
	case T_float32:
		return strconv.FormatFloat(float64(data.(float32)), 'E', -1, 32)
	default: // float64
		return strconv.FormatFloat(float64(data.(float64)), 'E', -1, 64)
	}
}

// must call validate first if it is applicable
func StringDecoding(dataType DataType, data any) (any, error) {
	switch {
	case AnyDataType[dataType]:
		if data != nil {
			return nil, fmt.Errorf("type 'any' requires nil Data")
		}

		return nil, nil
	case GeneralDataTypes[dataType]:
		if data == nil {
			return nil, fmt.Errorf("type '%s' has nil Data. Requires a castable string to the provided type", dataType.ToString())
		}

		if reflect.ValueOf(data).Kind() != reflect.String {
			return nil, fmt.Errorf("type '%s' has invalid data. Requires a string castable to the provided Type", dataType.ToString())
		}

		switch dataType {
		case T_uint:
			parsedValue, err := strconv.ParseUint(data.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			return uint(parsedValue), nil
		case T_uint8:
			parsedValue, err := strconv.ParseUint(data.(string), 10, 8)
			if err != nil {
				return nil, err
			}
			return uint8(parsedValue), nil
		case T_uint16:
			parsedValue, err := strconv.ParseUint(data.(string), 10, 16)
			if err != nil {
				return nil, err
			}
			return uint16(parsedValue), nil
		case T_uint32:
			parsedValue, err := strconv.ParseUint(data.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return uint32(parsedValue), nil
		case T_uint64:
			parsedValue, err := strconv.ParseUint(data.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			return uint64(parsedValue), nil
		case T_int:
			parsedValue, err := strconv.ParseInt(data.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			return int(parsedValue), nil
		case T_int8:
			parsedValue, err := strconv.ParseInt(data.(string), 10, 8)
			if err != nil {
				return nil, err
			}
			return int8(parsedValue), nil
		case T_int16:
			parsedValue, err := strconv.ParseInt(data.(string), 10, 16)
			if err != nil {
				return nil, err
			}
			return int16(parsedValue), nil
		case T_int32:
			parsedValue, err := strconv.ParseInt(data.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return int32(parsedValue), nil
		case T_int64:
			parsedValue, err := strconv.ParseInt(data.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			return int64(parsedValue), nil
		case T_float32:
			parsedValue, err := strconv.ParseFloat(data.(string), 32)
			if err != nil {
				return nil, err
			}
			return float32(parsedValue), nil
		case T_float64:
			parsedValue, err := strconv.ParseFloat(data.(string), 64)
			if err != nil {
				return nil, err
			}
			return float64(parsedValue), nil
		case T_string:
			return data.(string), nil
		}
	default:
	}

	return nil, fmt.Errorf("failed to decode JSON. Unknown type '%d'", dataType)
}

// UINT types
func Uint8(value uint8) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_uint8,
		Data: value,
	}
}

func Uint16(value uint16) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_uint16,
		Data: value,
	}
}

func Uint32(value uint32) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_uint32,
		Data: value,
	}
}

func Uint64(value uint64) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_uint64,
		Data: value,
	}
}

func Uint(value uint) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_uint,
		Data: value,
	}
}

// INT types
func Int8(value int8) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_int8,
		Data: value,
	}
}

func Int16(value int16) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_int16,
		Data: value,
	}
}

func Int32(value int32) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_int32,
		Data: value,
	}
}

func Int64(value int64) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_int64,
		Data: value,
	}
}

func Int(value int) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_int,
		Data: value,
	}
}

// float types
func Float32(value float32) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_float32,
		Data: value,
	}
}

func Float64(value float64) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_float64,
		Data: value,
	}
}

// STRING types
func String(value string) EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_string,
		Data: value,
	}
}

// ANY types
func Any() EncapsulatedValue {
	return EncapsulatedValue{
		Type: T_any,
		Data: nil,
	}
}
