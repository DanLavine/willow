package querymatchaction

import (
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

type MatchValue struct {
	Value datatypes.EncapsulatedValue `json:"Value"`

	TypeRestrictions v1.TypeRestrictions `json:"TypeRestrictions"`
}

func (matchValue MatchKeyValues) Validate() *errors.ModelError {
	// TODO
	return nil
}
