package datatypes

import (
	"fmt"
	"reflect"
)

// user defined interface that needs to implemented for each custom data type
// type ComparableDataType interface {
// 	// compare the objects including type and value
// 	Less(item ComparableDataType) bool

// 	// check if the compared object is Less then the object's type
// 	LessType(item ComparableDataType) bool

// 	// compare the objects' Values. This will panic if not coomparing the same types. So needs a
// 	// few checks against the LessType to know if this is safe.
// 	LessValue(item ComparableDataType) bool
// }

type DataType int

func (dt DataType) Less(dataType DataType) bool {
	return dt < dataType
}

var (
	T_custom  DataType = -1 // custom is a way for callers to provide their own types
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
	T_nil     DataType = 13 // there is no "value". Used when we only care about keys all pointing to the same thing

	// T_any might make sense, but for now I am not using it, so ignore that case
	// T_bool doesn't make much sense since it can oonly ever be true or false
)

type EncapsulatedData struct {
	DataType DataType

	Value any
}

func (edt EncapsulatedData) Less(comparableObj EncapsulatedData) bool {
	// know data type is less
	if edt.DataType < comparableObj.DataType {
		return true
	}

	// know data type is greater
	if edt.DataType > comparableObj.DataType {
		return false
	}

	// when the EncapsulatedData and ComparableDataType are the same, check the actual values
	switch edt.DataType {
	case T_custom:
		return edt.Value.(CheckLess).Less(comparableObj.Value)
	case T_uint8:
		return edt.Value.(uint8) < comparableObj.Value.(uint8)
	case T_uint16:
		return edt.Value.(uint16) < comparableObj.Value.(uint16)
	case T_uint32:
		return edt.Value.(uint32) < comparableObj.Value.(uint32)
	case T_uint64:
		return edt.Value.(uint64) < comparableObj.Value.(uint64)
	case T_uint:
		return edt.Value.(uint) < comparableObj.Value.(uint)
	case T_int8:
		return edt.Value.(int8) < comparableObj.Value.(int8)
	case T_int16:
		return edt.Value.(int16) < comparableObj.Value.(int16)
	case T_int32:
		return edt.Value.(int32) < comparableObj.Value.(int32)
	case T_int64:
		return edt.Value.(int64) < comparableObj.Value.(int64)
	case T_int:
		return edt.Value.(int) < comparableObj.Value.(int)
	case T_float32:
		return edt.Value.(float32) < comparableObj.Value.(float32)
	case T_float64:
		return edt.Value.(float64) < comparableObj.Value.(float64)
	case T_string:
		return edt.Value.(string) < comparableObj.Value.(string)
	default: // T_nil:
		// NOTE: This is important to always return false. This way on a btree when doing the lookup, we will always
		// find 1 copy of this item. The check is !item1.Less(item2) && !item2.Less(item1) -> returns true
		return false
	}
}

func (edt EncapsulatedData) LessType(comparableObj EncapsulatedData) bool {
	// know data type is less
	return edt.DataType < comparableObj.DataType
}

func (edt EncapsulatedData) LessValue(comparableObj EncapsulatedData) bool {
	// when the EncapsulatedData and ComparableDataType are the same, check the actual values
	switch edt.DataType {
	case T_custom:
		return edt.Value.(CheckLess).Less(comparableObj.Value)
	case T_uint8:
		return edt.Value.(uint8) < comparableObj.Value.(uint8)
	case T_uint16:
		return edt.Value.(uint16) < comparableObj.Value.(uint16)
	case T_uint32:
		return edt.Value.(uint32) < comparableObj.Value.(uint32)
	case T_uint64:
		return edt.Value.(uint64) < comparableObj.Value.(uint64)
	case T_uint:
		return edt.Value.(uint) < comparableObj.Value.(uint)
	case T_int8:
		return edt.Value.(int8) < comparableObj.Value.(int8)
	case T_int16:
		return edt.Value.(int16) < comparableObj.Value.(int16)
	case T_int32:
		return edt.Value.(int32) < comparableObj.Value.(int32)
	case T_int64:
		return edt.Value.(int64) < comparableObj.Value.(int64)
	case T_int:
		return edt.Value.(int) < comparableObj.Value.(int)
	case T_float32:
		return edt.Value.(float32) < comparableObj.Value.(float32)
	case T_float64:
		return edt.Value.(float64) < comparableObj.Value.(float64)
	case T_string:
		return edt.Value.(string) < comparableObj.Value.(string)
	default: // T_nil
		// NOTE: This is important to always return false. This way on a btree when doing the lookup, we will always
		// find 1 copy of this item. The check is !item1.Less(item2) && !item2.Less(item1) -> returns true
		if comparableObj.DataType != T_nil {
			panic("comparable object data type is not T_nil")
		}

		return false
	}
}

// Validate that Encapsulated data is correct. This will fail if any datatypes have a T_reserved
//
// NOTE: call this on the BTree, to check the params. not eveery time less is called
func (edt EncapsulatedData) Validate() error {
	if edt.Value == nil {
		return fmt.Errorf("EncapsulatedData has a nil data Value")
	}

	kind := reflect.ValueOf(edt.Value).Kind()

	switch edt.DataType {
	case T_custom:
		if !reflect.TypeOf(edt.Value).Implements(reflect.TypeOf((*CheckLess)(nil)).Elem()) {
			return fmt.Errorf("EncapsulatedData has a custom data type which dos not implement: CheckLess")
		}
	case T_uint8:
		if kind != reflect.Uint8 {
			return fmt.Errorf("EncapsulatedData has a uint8 data type, but the Value is a: %s", kind.String())
		}
	case T_uint16:
		if kind != reflect.Uint16 {
			return fmt.Errorf("EncapsulatedData has a uint16 data type, but the Value is a: %s", kind.String())
		}
	case T_uint32:
		if kind != reflect.Uint32 {
			return fmt.Errorf("EncapsulatedData has a uint32 data type, but the Value is a: %s", kind.String())
		}
	case T_uint64:
		if kind != reflect.Uint64 {
			return fmt.Errorf("EncapsulatedData has a uint64 data type, but the Value is a: %s", kind.String())
		}
	case T_uint:
		if kind != reflect.Uint {
			return fmt.Errorf("EncapsulatedData has a uint data type, but the Value is a: %s", kind.String())
		}
	case T_int8:
		if kind != reflect.Int8 {
			return fmt.Errorf("EncapsulatedData has an int8 data type, but the Value is a: %s", kind.String())
		}
	case T_int16:
		if kind != reflect.Int16 {
			return fmt.Errorf("EncapsulatedData has an int16 data type, but the Value is a: %s", kind.String())
		}
	case T_int32:
		if kind != reflect.Int32 {
			return fmt.Errorf("EncapsulatedData has an int32 data type, but the Value is a: %s", kind.String())
		}
	case T_int64:
		if kind != reflect.Int64 {
			return fmt.Errorf("EncapsulatedData has an int64 data type, but the Value is a: %s", kind.String())
		}
	case T_int:
		if kind != reflect.Int {
			return fmt.Errorf("EncapsulatedData has an int data type, but the Value is a: %s", kind.String())
		}
	case T_float32:
		if kind != reflect.Float32 {
			return fmt.Errorf("EncapsulatedData has a float32 data type, but the Value is a: %s", kind.String())
		}
	case T_float64:
		if kind != reflect.Float64 {
			return fmt.Errorf("EncapsulatedData has a float64 data type, but the Value is a: %s", kind.String())
		}
	case T_string:
		if kind != reflect.String {
			return fmt.Errorf("EncapsulatedData has a string data type, but the Value is a: %s", kind.String())
		}
	case T_nil:
		if edt.Value != struct{}{} {
			return fmt.Errorf("EncapsulatedData has a 'nil' data type and requires the Value to be nil")
		}
	default:
		return fmt.Errorf("EncapsulatedData has an unkown data type")
	}

	return nil
}

// CUSTOM types
type CheckLess interface {
	// only return true if the value is less than the parameter
	Less(item any) bool
}

// Custom can be used by any Service when saving values to the AssociatedTree, and needs to create
// keys that are guranteed to not collide with an end user's values. So it is important that no
// APIs allow for receiving of this type.
func Custom(value CheckLess) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_custom,
		Value:    value,
	}
}

// UINT types
func Uint8(value uint8) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint8,
		Value:    value,
	}
}

func Uint16(value uint16) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint16,
		Value:    value,
	}
}

func Uint32(value uint32) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint32,
		Value:    value,
	}
}

func Uint64(value uint64) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint64,
		Value:    value,
	}
}

func Uint(value uint) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_uint,
		Value:    value,
	}
}

// INT types
func Int8(value int8) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int8,
		Value:    value,
	}
}

func Int16(value int16) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int16,
		Value:    value,
	}
}

func Int32(value int32) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int32,
		Value:    value,
	}
}

func Int64(value int64) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int64,
		Value:    value,
	}
}

func Int(value int) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_int,
		Value:    value,
	}
}

// float types
func Float32(value float32) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_float32,
		Value:    value,
	}
}

func Float64(value float64) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_float64,
		Value:    value,
	}
}

// STRING types
func String(value string) EncapsulatedData {
	return EncapsulatedData{
		DataType: T_string,
		Value:    value,
	}
}

// DSL TODO: rename to Empty()?
// NIL types
func Nil() EncapsulatedData {
	return EncapsulatedData{
		DataType: T_nil,
		Value:    struct{}{}, // NOTE: we really use the empty struct here to account for all values
	}
}
