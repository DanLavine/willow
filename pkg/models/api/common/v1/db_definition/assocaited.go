package dbdefinition

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type Associated struct {
	// Association required TypedKeyValues
	TypedKeyValues TypedKeyValues `json:"TypedKeyValues,omitempty"`

	// Association can use AnyKeyValues
	AnyKeyValues AnyKeyValues `json:"AnyKeyValues,omitempty"`
}

func (associated *Associated) Validate() *errors.ModelError {
	if associated.TypedKeyValues == nil && associated.AnyKeyValues == nil {
		return &errors.ModelError{Err: fmt.Errorf("TypedKeyValues and AnyKeyValues are both empty")}
	}

	if associated.TypedKeyValues != nil && associated.AnyKeyValues != nil {
		return &errors.ModelError{Err: fmt.Errorf("TypedKeyValues and AnyKeyValues are both set, but requires eaxtly 1 to be set")}
	}

	if associated.TypedKeyValues != nil {
		if err := associated.TypedKeyValues.Validate(); err != nil {
			return &errors.ModelError{Field: "TypedKeyValues", Child: err}
		}
	}

	if associated.AnyKeyValues != nil {
		if err := associated.AnyKeyValues.Validate(); err != nil {
			return &errors.ModelError{Field: "AnyKeyValues", Child: err}
		}
	}

	return nil
}
