package v1

// import (
// 	"fmt"

// 	"github.com/DanLavine/willow/pkg/models/api/common/errors"
// 	"github.com/DanLavine/willow/pkg/models/datatypes"
// )

// type TypedKeyValues map[string]datatypes.EncapsulatedValue

// func (typedKeyValues TypedKeyValues) Validate() error {
// 	for key, value := range typedKeyValues {
// 		if err := value.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
// 			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err.(*errors.ModelError)}
// 		}
// 	}

// 	return nil
// }

// func (typedKeyValues TypedKeyValues) ToEncapsulatedKeyValues() datatypes.KeyValues {
// 	return datatypes.KeyValues(typedKeyValues)
// }

// type AnyKeyValues map[string]datatypes.EncapsulatedValue

// func (anyKeyValues AnyKeyValues) Validate() error {
// 	for key, value := range anyKeyValues {
// 		if err := value.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
// 			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err.(*errors.ModelError)}
// 		}
// 	}

// 	return nil
// }

// func (anyKeyValues AnyKeyValues) ToEncapsulatedKeyValues() datatypes.KeyValues {
// 	return datatypes.KeyValues(anyKeyValues)
// }
