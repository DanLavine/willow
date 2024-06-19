package querymatchaction

import (
	"fmt"
	"sort"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type MatchKeyValues map[string]MatchValue

func (matchKeyValues MatchKeyValues) Keys() []string {
	keys := []string{}

	for key, _ := range matchKeyValues {
		keys = append(keys, key)
	}

	return keys
}

func (matchKeyValues MatchKeyValues) SortedKeys() []string {
	keys := matchKeyValues.Keys()
	sort.Strings(keys)

	return keys
}

func (matchValue MatchKeyValues) Validate() *errors.ModelError {
	for key, value := range matchValue {
		if err := value.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%s]", key), Child: err}
		}
	}

	return nil
}

type MatchValue struct {
	Value datatypes.EncapsulatedValue `json:"Value"`

	TypeRestrictions v1.TypeRestrictions `json:"TypeRestrictions"`
}

func (matchValue MatchValue) Validate() *errors.ModelError {
	if err := matchValue.TypeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "TypeRestrictions", Child: err}
	}

	if err := matchValue.Value.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return &errors.ModelError{Field: "Value", Child: err}
	}

	return nil
}
