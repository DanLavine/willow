package queryassociatedaction

import (
	"fmt"
	"sort"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type SelectionKeyValues map[string]ValueQuery

func (associatedSelectionKeyValues SelectionKeyValues) Keys() []string {
	keys := []string{}

	for key, _ := range associatedSelectionKeyValues {
		keys = append(keys, key)
	}

	return keys
}

func (associatedSelectionKeyValues SelectionKeyValues) SortedKeys() []string {
	keys := associatedSelectionKeyValues.Keys()

	sort.Strings(keys)
	return keys
}

type Selection struct {
	// Select the Specific IDs that we want to query against. If KeyValues field is provided as well,
	// then the DB's KeyValues will be used to ensure that these DB object's KeyValues query KeyValues
	IDs []string `json:"IDs,omitempty"`

	// can be used to specfy a collection of KeyValues that must exist
	KeyValues SelectionKeyValues `json:"KeyValues,omitempty"`
	// limits for AssociatedKeyValues that make up an item saved in the tree
	MinNumberOfKeyValues *int `json:"MinNumberOfKeyValues,omitempty"`
	MaxNumberOfKeyValues *int `json:"MaxNumberOfKeyValues,omitempty"`
}

func (associatedSelection *Selection) Validate() error {
	if len(associatedSelection.IDs) == 0 && len(associatedSelection.KeyValues) == 0 && associatedSelection.MinNumberOfKeyValues == nil && associatedSelection.MaxNumberOfKeyValues == nil {
		return &errors.ModelError{Err: fmt.Errorf("requires 'IDs', 'KeyValues', 'MinNumberOfKeyValues' or 'MaxNumberOfKeyValues' to be specified, but received nothing")}
	}

	for key, value := range associatedSelection.KeyValues {
		if err := value.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("KeyValues[%s]", key), Child: err.(*errors.ModelError)}
		}
	}

	// ensure number of KeyValues are proper
	if associatedSelection.MinNumberOfKeyValues != nil && associatedSelection.MaxNumberOfKeyValues != nil {
		if *associatedSelection.MaxNumberOfKeyValues < *associatedSelection.MinNumberOfKeyValues {
			return &errors.ModelError{Err: fmt.Errorf("MaxNumberOfKeyValues is less than MinNumberOfKeyValues")}
		}
	}

	return nil
}

func (associatedSelection Selection) Keys() []string {
	keys := []string{}

	for key, _ := range associatedSelection.KeyValues {
		keys = append(keys, key)
	}

	return keys
}

func (associatedSelection Selection) SortedKeys() []string {
	keys := associatedSelection.Keys()

	sort.Strings(keys)
	return keys
}

// MatchTags can be used to see if a collection of KeyValues would match against the query
func (associatedSelection *Selection) MatchKeyValues(matchKeyValues datatypes.KeyValues) bool {
	// quick rejection on being under the min number of keys
	if associatedSelection.MinNumberOfKeyValues != nil {
		if len(matchKeyValues) < *associatedSelection.MinNumberOfKeyValues {
			return false
		}
	}

	// quick rejection on being over the max number of keys
	if associatedSelection.MaxNumberOfKeyValues != nil {
		if len(matchKeyValues) > *associatedSelection.MaxNumberOfKeyValues {
			return false
		}
	}

	// need to validate all the keys in the query against the passed in keyValues
	for queryKey, valueQuery := range associatedSelection.KeyValues {
		switch valueQuery.Comparison {
		case v1.Equals:
			if matchValue, ok := matchKeyValues[queryKey]; ok {
				// not within the allowed range
				if !valueQuery.TypeRestrictions.TypeWithinRange(matchValue.Type) {
					return false
				}

				// not the exact specified value
				if !(!matchValue.Less(valueQuery.Value) && !valueQuery.Value.Less(matchValue)) {
					return false
				}
			} else {
				return false
			}
		case v1.NotEquals:
			if matchValue, ok := matchKeyValues[queryKey]; ok {
				if valueQuery.TypeRestrictions.TypeWithinRange(matchValue.Type) {
					// the exact key we don't want to find
					if !matchValue.Less(valueQuery.Value) && !valueQuery.Value.Less(matchValue) {
						return false
					}
				}
			}
		case v1.LessThan:
			if matchValue, ok := matchKeyValues[queryKey]; ok {
				// not within the allowed type range
				if !valueQuery.TypeRestrictions.TypeWithinRange(matchValue.Type) {
					return false
				}

				switch {
				// case when we are not allowing the T_any in the matchKeyValues
				case datatypes.GeneralDataTypes[valueQuery.TypeRestrictions.MaxDataType]:
					if !matchValue.LessMatchType(valueQuery.Value) {
						return false
					}
				// case when we are allowing the T_any in the matchKeyValues
				case datatypes.AnyDataType[valueQuery.TypeRestrictions.MaxDataType]:
					switch matchValue.Type {
					case datatypes.T_any:
						// always true
					default:
						if !matchValue.LessMatchType(valueQuery.Value) {
							return false
						}
					}
				}
			} else {
				// key not found, always false
				return false
			}
		case v1.LessThanOrEqual:
			if matchValue, ok := matchKeyValues[queryKey]; ok {
				// not within the allowed type range
				if !valueQuery.TypeRestrictions.TypeWithinRange(matchValue.Type) {
					return false
				}

				switch {
				// case when we are not allowing the T_any in the matchKeyValues
				case datatypes.GeneralDataTypes[valueQuery.TypeRestrictions.MaxDataType]:
					if !matchValue.LessMatchType(valueQuery.Value) {
						if valueQuery.Value.Less(matchValue) {
							return false
						}
					}
				// case when we are allowing the T_any in the matchKeyValues
				case datatypes.AnyDataType[valueQuery.TypeRestrictions.MaxDataType]:
					switch matchValue.Type {
					case datatypes.T_any:
						// always true
					default:
						if !matchValue.LessMatchType(valueQuery.Value) {
							if valueQuery.Value.Less(matchValue) {
								return false
							}
						}
					}
				}
			} else {
				// key not found, always false
				return false
			}
		case v1.GreaterThan:
			if matchValue, ok := matchKeyValues[queryKey]; ok {
				// not within the allowed type range
				if !valueQuery.TypeRestrictions.TypeWithinRange(matchValue.Type) {
					return false
				}

				switch {
				// case when we are not allowing the T_any in the matchKeyValues
				case datatypes.GeneralDataTypes[valueQuery.TypeRestrictions.MaxDataType]:
					if !valueQuery.Value.LessMatchType(matchValue) {
						return false
					}
				// case when we are allowing the T_any in the matchKeyValues
				case datatypes.AnyDataType[valueQuery.TypeRestrictions.MaxDataType]:
					switch matchValue.Type {
					case datatypes.T_any:
						// always true
					default:
						if !valueQuery.Value.LessMatchType(matchValue) {
							return false
						}
					}
				}
			} else {
				// key not found, always false
				return false
			}
		case v1.GreaterThanOrEqual:
			if matchValue, ok := matchKeyValues[queryKey]; ok {
				// not within the allowed type range
				if !valueQuery.TypeRestrictions.TypeWithinRange(matchValue.Type) {
					return false
				}

				switch {
				// case when we are not allowing the T_any in the matchKeyValues
				case datatypes.GeneralDataTypes[valueQuery.TypeRestrictions.MaxDataType]:
					if !valueQuery.Value.LessMatchType(matchValue) {
						if matchValue.Less(valueQuery.Value) {
							return false
						}
					}
				// case when we are allowing the T_any in the matchKeyValues
				case datatypes.AnyDataType[valueQuery.TypeRestrictions.MaxDataType]:
					switch matchValue.Type {
					case datatypes.T_any:
						// always true
					default:
						if !valueQuery.Value.LessMatchType(matchValue) {
							if matchValue.Less(valueQuery.Value) {
								return false
							}
						}
					}
				}
			} else {
				// key not found, always false
				return false
			}
		default:
			panic(fmt.Sprintf("missed a comparison somehow: %s", valueQuery.Comparison))
		}
	}

	return true
}
