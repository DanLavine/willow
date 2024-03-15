package queryassociatedaction

import (
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type ValueQuery struct {
	// Value associated with the key to lookup
	Value datatypes.EncapsulatedValue `json:"Value"`

	// how to compare the value
	Comparison v1.DataComparison `json:"Comparison"`

	// restrictions on possible key value combinations we want to find. These indicate the allowed
	// KeyValue's type to be searchable in the tree.
	//
	// NOTE: The special case that can be a bit confusing is for the `Comparison = '!='`. In this case, the
	// TypeRestrictions identify the specific key + Any() to not include if they allow for it. Otherwise, the
	// != Comparison will just ignore the specific provided key
	TypeRestrictions v1.TypeRestrictions `json:"TypeRestrictions"`
}

func (valueQuery ValueQuery) Validate() error {
	if err := valueQuery.Comparison.Validate(); err != nil {
		return &errors.ModelError{Field: "Comparison", Child: err.(*errors.ModelError)}
	}

	switch valueQuery.Comparison {
	case v1.Equals, v1.NotEquals:
		if err := valueQuery.Value.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
			return &errors.ModelError{Field: "Value", Child: err.(*errors.ModelError)}
		}
	case v1.LessThan, v1.LessThanOrEqual, v1.GreaterThan, v1.GreaterThanOrEqual:
		if err := valueQuery.Value.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
			return &errors.ModelError{Field: "Value", Child: err.(*errors.ModelError)}
		}
	}

	if err := valueQuery.TypeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "TypeRestrictions", Child: err.(*errors.ModelError)}
	}

	return nil
}
