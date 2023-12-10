package datatypes

import (
	"fmt"
	"reflect"
)

type DataType int

func (dt DataType) Less(compare DataType) bool {
	return dt < compare
}

var (
	// these are the types that can be added
	//T_custom  DataType = -1000 // custom is a way for services to provide their own types.
	T_uint8   DataType = 0
	T_uint16  DataType = 1
	T_uint32  DataType = 2
	T_uint64  DataType = 3
	T_uint    DataType = 4
	T_int8    DataType = 5
	T_int16   DataType = 6
	T_int32   DataType = 7
	T_int64   DataType = 8
	T_int     DataType = 9
	T_float32 DataType = 10
	T_float64 DataType = 11
	T_string  DataType = 12

	// T_any might make sense, but for now I am not using it, so ignore that case
	// T_bool doesn't seem to make much sense since it can oonly ever be true or false
)

// EncapsulatedValue provides validation for all datatypes from uint8 to string.
// It cacn be used to enforce that a proper single value is provided
type EncapsulatedValue struct {
	Type DataType

	Data any
}

func (edv EncapsulatedValue) DataType() DataType {
	return edv.Type
}

func (edv EncapsulatedValue) Value() any {
	return edv.Data
}

func (edv EncapsulatedValue) Less(comparableObj EncapsulatedData) bool {
	// know data type is less
	if edv.Type < comparableObj.DataType() {
		return true
	}

	// know data type is greater
	if edv.Type > comparableObj.DataType() {
		return false
	}

	return edv.LessValue(comparableObj)
}

func (edv EncapsulatedValue) LessType(comparableObj EncapsulatedData) bool {
	// know data type is less
	return edv.Type < comparableObj.DataType()
}

func (edv EncapsulatedValue) LessValue(comparableObj EncapsulatedData) bool {
	switch edv.Type {
	//case T_custom:
	//	return edv.Data.(CustomType).Less(comparableObj.Value())
	case T_uint8:
		return edv.Data.(uint8) < comparableObj.Value().(uint8)
	case T_uint16:
		return edv.Data.(uint16) < comparableObj.Value().(uint16)
	case T_uint32:
		return edv.Data.(uint32) < comparableObj.Value().(uint32)
	case T_uint64:
		return edv.Data.(uint64) < comparableObj.Value().(uint64)
	case T_uint:
		return edv.Data.(uint) < comparableObj.Value().(uint)
	case T_int8:
		return edv.Data.(int8) < comparableObj.Value().(int8)
	case T_int16:
		return edv.Data.(int16) < comparableObj.Value().(int16)
	case T_int32:
		return edv.Data.(int32) < comparableObj.Value().(int32)
	case T_int64:
		return edv.Data.(int64) < comparableObj.Value().(int64)
	case T_int:
		return edv.Data.(int) < comparableObj.Value().(int)
	case T_float32:
		return edv.Data.(float32) < comparableObj.Value().(float32)
	case T_float64:
		return edv.Data.(float64) < comparableObj.Value().(float64)
	case T_string:
		return edv.Data.(string) < comparableObj.Value().(string)
	default:
		panic(fmt.Sprintf("Unexpected type %d", edv.Type))
	}
}

// Validate all Encpasulated data types inclusing custom
func (edv EncapsulatedValue) Validate() error {
	if edv.Data == nil {
		return fmt.Errorf("EncapsulatedValue has a nil data Value")
	}

	kind := reflect.ValueOf(edv.Data).Kind()

	switch edv.Type {
	case T_uint8:
		if kind != reflect.Uint8 {
			return fmt.Errorf("EncapsulatedValue has a uint8 data type, but the Value is a: %s", kind.String())
		}
	case T_uint16:
		if kind != reflect.Uint16 {
			return fmt.Errorf("EncapsulatedValue has a uint16 data type, but the Value is a: %s", kind.String())
		}
	case T_uint32:
		if kind != reflect.Uint32 {
			return fmt.Errorf("EncapsulatedValue has a uint32 data type, but the Value is a: %s", kind.String())
		}
	case T_uint64:
		if kind != reflect.Uint64 {
			return fmt.Errorf("EncapsulatedValue has a uint64 data type, but the Value is a: %s", kind.String())
		}
	case T_uint:
		if kind != reflect.Uint {
			return fmt.Errorf("EncapsulatedValue has a uint data type, but the Value is a: %s", kind.String())
		}
	case T_int8:
		if kind != reflect.Int8 {
			return fmt.Errorf("EncapsulatedValue has an int8 data type, but the Value is a: %s", kind.String())
		}
	case T_int16:
		if kind != reflect.Int16 {
			return fmt.Errorf("EncapsulatedValue has an int16 data type, but the Value is a: %s", kind.String())
		}
	case T_int32:
		if kind != reflect.Int32 {
			return fmt.Errorf("EncapsulatedValue has an int32 data type, but the Value is a: %s", kind.String())
		}
	case T_int64:
		if kind != reflect.Int64 {
			return fmt.Errorf("EncapsulatedValue has an int64 data type, but the Value is a: %s", kind.String())
		}
	case T_int:
		if kind != reflect.Int {
			return fmt.Errorf("EncapsulatedValue has an int data type, but the Value is a: %s", kind.String())
		}
	case T_float32:
		if kind != reflect.Float32 {
			return fmt.Errorf("EncapsulatedValue has a float32 data type, but the Value is a: %s", kind.String())
		}
	case T_float64:
		if kind != reflect.Float64 {
			return fmt.Errorf("EncapsulatedValue has a float64 data type, but the Value is a: %s", kind.String())
		}
	case T_string:
		if kind != reflect.String {
			return fmt.Errorf("EncapsulatedValue has a string data type, but the Value is a: %s", kind.String())
		}
	default:
		return fmt.Errorf("EncapsulatedValue has an unkown data type: %d", edv.Type)
	}

	return nil
}

// CUSTOM types
//type CustomType interface {
//	// only return true if the value is less than the parameter
//	Less(item any) bool
//
//	// // encode a value to transfer over an api call
//	// Encode() []byte
//	//
//	// // decode the value transfered over an api call
//	// Decode(b []byte) (any, error)
//}
//
//// Custom can be used by any Service when saving values to the AssociatedTree, and needs to create
//// keys that are guranteed to not collide with an end user's values. So it is important that no
//// APIs allow for receiving of this type.
//func Custom(value CustomType) EncapsulatedValue {
//	return EncapsulatedValue{
//		Type: T_custom,
//		Data: value,
//	}
//}

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
