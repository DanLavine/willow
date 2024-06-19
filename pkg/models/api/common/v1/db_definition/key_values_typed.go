package dbdefinition

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// TypedKeyValues are used to ensure that the KeyValues all have a specific typed value and not using the `Any` type
type TypedKeyValues datatypes.KeyValues

func (typedKeyValues TypedKeyValues) Validate() *errors.ModelError {
	if len(typedKeyValues) == 0 {
		return &errors.ModelError{Err: fmt.Errorf("received a length of 0 key values")}
	}

	for key, value := range typedKeyValues {
		if err := value.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err}
		}
	}

	return nil
}

func (typedKeyValues *TypedKeyValues) ToKeyValues() datatypes.KeyValues {
	return datatypes.KeyValues(*typedKeyValues)
}

func KeyValuesToTypedKeyValues(keyValues datatypes.KeyValues) TypedKeyValues {
	typedKeyValues := TypedKeyValues{}

	for key, value := range keyValues {
		typedKeyValues[key] = value
	}

	return typedKeyValues
}
