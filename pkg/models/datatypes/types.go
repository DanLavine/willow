package datatypes

import (
	"fmt"
	"reflect"
)

// when trees require a slice of tree keys
type EnumerableCompareType interface {
	// pop the first index off of the Enumerable list and return it.
	// When there are no more Values in the list, return nil
	Pop() (ComparableDataType, EnumerableCompareType)

	// return the len(...) call of underlying type
	Len() int
}

// user defined interface that needs to implemented for each custom data type
type ComparableDataType interface {
	Less(item ComparableDataType) bool
}

type DataType int

const (
	T_uint8 DataType = iota
	T_uint16
	T_uint32
	T_uint64
	T_uint
	T_int8
	T_int16
	T_int32
	T_int64
	T_int
	T_float32
	T_float64
	T_string
	T_nil // there is no "value". Used when we only care about keys all pointing to the same thing
	T_any // the value is a complex user defined struct

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
	case T_nil:
		// NOTE: This is important to always return false. This way on a btree when doing the lookup, we will always
		// find 1 copy of this item. The check is !item1.Less(item2) && !item2.Less(item1) -> returns true
		return false
	default: // T_any
		// In this case the user defined both Values and should know how to compare their own data types
		//
		// NOTE: we know these casts are safe since the end user created these from the `Any(...)` function
		return edt.Value.(ComparableDataType).Less(comparableObj.Value.(ComparableDataType))
	}
}

// Validate that Encapsulated data is correct
//
// NOTE: call this on the BTree, to check the params. not eveery time less is called
func (edt EncapsulatedData) Validate() error {
	if edt.Value == nil {
		return fmt.Errorf("EncapsulatedData has a nil data Value")
	}

	kind := reflect.ValueOf(edt.Value).Kind()

	switch edt.DataType {
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
		if edt.Value != nil {
			return fmt.Errorf("EncapsulatedData has a 'nil' data type and requires the Value to be nil")
		}
	case T_any:
		if _, ok := edt.Value.(ComparableDataType); !ok {
			return fmt.Errorf("EncapsulatedData has an 'any' data type and requires the Value to be a ComparableDataType interface")
		}
	default: // T_any
		return fmt.Errorf("EncapsulatedData has an unkown data type")
	}

	return nil
}
