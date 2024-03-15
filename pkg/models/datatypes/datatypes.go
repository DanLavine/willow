package datatypes

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

/*
 * DataType defines all the possible values that an encapsulated value or query can be
 */

type DataType int

var (
	T_uint8   DataType = 1
	T_uint16  DataType = 2
	T_uint32  DataType = 3
	T_uint64  DataType = 4
	T_uint    DataType = 5
	T_int8    DataType = 6
	T_int16   DataType = 7
	T_int32   DataType = 8
	T_int64   DataType = 9
	T_int     DataType = 10
	T_float32 DataType = 11
	T_float64 DataType = 12
	T_string  DataType = 13
	T_any     DataType = 1024

	AnyDataType      = map[DataType]bool{T_any: true}
	GeneralDataTypes = map[DataType]bool{T_uint: true, T_uint8: true, T_uint16: true, T_uint32: true, T_uint64: true, T_int: true, T_int8: true, T_int16: true, T_int32: true, T_int64: true, T_float32: true, T_float64: true, T_string: true}

	GeneralDatatypesSlice = []DataType{T_uint, T_uint8, T_uint16, T_uint32, T_uint64, T_int, T_int8, T_int16, T_int32, T_int64, T_float32, T_float64, T_string}
	AllDatatypesSlice     = []DataType{T_uint, T_uint8, T_uint16, T_uint32, T_uint64, T_int, T_int8, T_int16, T_int32, T_int64, T_float32, T_float64, T_string, T_any}

	MinDataType           DataType = T_uint8
	MaxDataType           DataType = T_any
	MaxWithoutAnyDataType DataType = T_string
)

// Less can be used to account for the T_Any special case that always returns false
func (dt DataType) Less(compare DataType) bool {
	if dt == T_any || compare == T_any {
		return false
	}

	if dt < compare {
		return true
	}

	return false
}

// LessMatchType can be used to strictly match the types being compared
func (dt DataType) LessMatchType(compare DataType) bool {
	return dt < compare
}

// ToString converst the DataType to a human readable format
func (dt DataType) ToString() string {
	switch dt {
	case T_uint8:
		return "uint8"
	case T_uint16:
		return "uint16"
	case T_uint32:
		return "uint32"
	case T_uint64:
		return "uint64"
	case T_uint:
		return "uint"
	case T_int8:
		return "int8"
	case T_int16:
		return "int16"
	case T_int32:
		return "int32"
	case T_int64:
		return "int64"
	case T_int:
		return "int"
	case T_float32:
		return "float32"
	case T_float64:
		return "float64"
	case T_string:
		return "string"
	case T_any:
		return "any"
	default:
		return "unknown"
	}
}

func (dt DataType) Validate(min, max DataType) error {
	if dt < min || dt > max {
		return &errors.ModelError{Err: fmt.Errorf("is outside the range for for valid types [%d:%d], but received '%d'", min, max, dt)}
	}

	switch {
	case AnyDataType[dt], GeneralDataTypes[dt]:
		// these are all fine
	default:
		return &errors.ModelError{Err: fmt.Errorf("uknown value recieved '%d'", dt)}
	}

	return nil
}
