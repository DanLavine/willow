package datatypes

/*

This has all moved to the associated_value_query.go in the v1common

import (
	"encoding/json"
	"fmt"
)

// TypeQuery is used to select specific types for keys. When using the T_any type, we can
// select all possible values for a key.
//
// These queries are equivlent to running "Iterate" opartions on the BTrees
type TypeQuery struct {
	Type DataType `json:"Type"`

	// Iff true, only data with the desire keys will be considered. When false, any values that have the keys
	// will not be considetred for the objects found in the query.
	Exists bool `json:"Exists"`

	//ExistanceTypeMatch ExistanceTypeMatch `json:"ExistanceTypeMatch"`
	DataTypeRestrictions DataTypeRestrictions `json:"DataTypeRestrictions"`
}

func (teq TypeQuery) Validate() error {
	if err := teq.Type.Validate(MinDataType, MaxDataType); err != nil {
		return &ModelError{Field: "Type", Child: err.(*ModelError)}
	}

	if err := teq.DataTypeRestrictions.Validate(); err != nil {
		return &ModelError{Field: "DataTypeRestrictions", Child: err.(*ModelError)}
	}

	return nil
}

type DataComparison string

const (
	// DataComparison
	Equals             DataComparison = "="
	NotEquals          DataComparison = "!="
	LessThan           DataComparison = "<"
	LessThanOrEqual    DataComparison = "<="
	GreaterThan        DataComparison = ">"
	GreaterThanOrEqual DataComparison = ">="
)

// EncapsulatedValueDataQuery is used to select specific key value combinations. For this reason, the T_any
// type is not valid for the encapsulated value as those are specific to existance checks, rather than value checks.
//
// These queries are equivlent to running "Find" opartions on the BTrees
type DataQuery struct {
	// These can be any valid type other than T_any
	EncapsulatedValue EncapsulatedValue `json:"EncapsulatedValue"`

	// How to compare data
	DataComparison DataComparison `json:"DataComparison"`

	// When comparing agains other object, how should the types be evaluated
	// DataTypeMatch DataTypeMatch `json:"DataTypeMatch"`
	DataTypeRestrictions DataTypeRestrictions `json:"DataTypeRestrictions"`
}

func (evdq DataQuery) Validate() error {
	if err := evdq.EncapsulatedValue.Validate(MinDataType, MaxWithoutAnyDataType); err != nil {
		return &ModelError{Field: "EncapsulatedValue", Child: err.(*ModelError)}
	}

	switch evdq.DataComparison {
	case Equals, NotEquals, LessThan, LessThanOrEqual, GreaterThan, GreaterThanOrEqual:
		// these are all fine
	default:
		return &ModelError{Field: "DataComparison", Err: fmt.Errorf("unknown value received '%s'", evdq.DataComparison)}
	}

	if err := evdq.DataTypeRestrictions.Validate(); err != nil {
		return &ModelError{Field: "DataTypeRestrictions", Child: err.(*ModelError)}
	}

	return nil
}

// DataTypeRestrictions can be used to enforce specific queries to select a range of types, or ignore the
// T_any keys that are saved service side if they make no sense.
type DataTypeRestrictions struct {
	// Setting this will enfore a restriction on comparing the value to a minimum type
	MinDataType DataType `json:"MinDataType"`

	// Setting this will enforce the query to only use a max data type.
	// I.E: Where 'key1' >= Uint8(3) MAX_DATA_TYPE T_Uint // select only the Uint range of types
	MaxDataType DataType `json:"MaxDataType"`
}

func (dtr DataTypeRestrictions) Validate() error {
	if dtr.MinDataType == 0 && dtr.MaxDataType == 0 {
		return &ModelError{Err: fmt.Errorf("MinDataType and MaxDataType are both 0")}
	}

	if dtr.MinDataType != 0 {
		switch {
		case GeneralDataTypes[dtr.MinDataType], AnyDataType[dtr.MinDataType]:
			// these are fine
		default:
			return &ModelError{Field: "MinDataType", Err: fmt.Errorf("unknown value received '%d'", dtr.MinDataType)}
		}
	}

	if dtr.MaxDataType != 0 {
		switch {
		case GeneralDataTypes[dtr.MaxDataType], AnyDataType[dtr.MaxDataType]:
			// these are fine
		default:
			return &ModelError{Field: "MaxDataType", Err: fmt.Errorf("unknown value received '%d'", dtr.MaxDataType)}
		}
	}

	if dtr.MinDataType > dtr.MaxDataType {
		return &ModelError{Err: fmt.Errorf("MinDataType is greater than MaxDataType")}
	}

	return nil
}

// UnmarshalJSON satisfies the json.Unrmashal interface and sets default values
func (dtr *DataTypeRestrictions) UnmarshalJSON(b []byte) error {
	type alias DataTypeRestrictions
	tmp := alias{}

	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	if tmp.MinDataType == 0 {
		tmp.MinDataType = T_uint8
	}

	if tmp.MaxDataType == 0 {
		tmp.MaxDataType = T_any
	}

	dtr.MinDataType = tmp.MinDataType
	dtr.MaxDataType = tmp.MaxDataType

	return nil
}
*/
