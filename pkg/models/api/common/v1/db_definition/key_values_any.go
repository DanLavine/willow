package dbdefinition

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// AnyKeyValues are used to allow all possible KeyValue combinations
type AnyKeyValues datatypes.KeyValues

func (anyKeyValues AnyKeyValues) Validate() *errors.ModelError {
	if len(anyKeyValues) == 0 {
		return &errors.ModelError{Err: fmt.Errorf("received a length of 0 key values")}
	}

	for key, value := range anyKeyValues {
		if err := value.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err}
		}
	}

	return nil
}

func (anyKeyValues *AnyKeyValues) ToKeyValues() datatypes.KeyValues {
	return datatypes.KeyValues(*anyKeyValues)
}

func KeyValuesToAnyKeyValues(keyValues datatypes.KeyValues) AnyKeyValues {
	anyKeyValues := AnyKeyValues{}

	for key, value := range keyValues {
		anyKeyValues[key] = value
	}

	return anyKeyValues
}
