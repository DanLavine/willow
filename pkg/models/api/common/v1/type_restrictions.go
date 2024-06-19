package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type TypeRestrictions struct {
	// Setting this will enfore a restriction on comparing the value to a minimum type
	MinDataType datatypes.DataType `json:"MinDataType"`

	// Setting this will enforce the query to only use a max data type.
	// I.E: Where 'key1' >= Uint8(3) MAX_DATA_TYPE T_Uint // select only the Uint range of types
	MaxDataType datatypes.DataType `json:"MaxDataType"`
}

func (valueTypeRestrictions TypeRestrictions) Validate() *errors.ModelError {
	switch {
	case datatypes.GeneralDataTypes[valueTypeRestrictions.MinDataType], datatypes.AnyDataType[valueTypeRestrictions.MinDataType]:
		// these are fine
	default:
		return &errors.ModelError{Field: "MinDataType", Err: fmt.Errorf("unknown value received '%d'", valueTypeRestrictions.MinDataType)}
	}

	switch {
	case datatypes.GeneralDataTypes[valueTypeRestrictions.MaxDataType], datatypes.AnyDataType[valueTypeRestrictions.MaxDataType]:
		// these are fine
	default:
		return &errors.ModelError{Field: "MaxDataType", Err: fmt.Errorf("unknown value received '%d'", valueTypeRestrictions.MaxDataType)}
	}

	if valueTypeRestrictions.MinDataType > valueTypeRestrictions.MaxDataType {
		return &errors.ModelError{Err: fmt.Errorf("MinDataType is greater than MaxDataType")}
	}

	return nil
}

func (valueTypeRestrictions TypeRestrictions) TypeWithinRange(checkType datatypes.DataType) bool {
	if checkType < valueTypeRestrictions.MinDataType {
		return false
	}

	if checkType > valueTypeRestrictions.MaxDataType {
		return false
	}

	return true
}
