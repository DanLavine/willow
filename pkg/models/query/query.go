package query

import (
	"fmt"
	"sort"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Query struct {
	// Key Values ensures that all key values exist
	KeyValues map[string]Value

	// limits for collections to look for
	Limits *KeyLimits
}

type KeyLimits struct {
	// limit how many keys can make up a collection
	NumberOfKeys *int
}

func (q *Query) SortedKeys() []string {
	keys := []string{}
	for key, _ := range q.KeyValues {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func (q *Query) Validate() error {
	if len(q.KeyValues) == 0 && q.Limits == nil {
		return fmt.Errorf(": requires KeyValues or Limits parameters")
	}

	if q.Limits != nil {
		if err := q.Limits.Validate(len(q.KeyValues)); err != nil {
			return fmt.Errorf(".Limits.%s", err.Error())
		}
	}

	for key, value := range q.KeyValues {
		if err := value.Validate(); err != nil {
			return fmt.Errorf(".KeyValues[%s]%s", key, err.Error())
		}
	}

	return nil
}

func (kl *KeyLimits) Validate(keys int) error {
	if kl.NumberOfKeys == nil {
		return fmt.Errorf("NumberOfKeys: Requires an int to be provided")
	}

	if *kl.NumberOfKeys <= 0 {
		return fmt.Errorf("NumberOfKeys: Must be larger than the provided 0")
	}

	if keys != 0 && keys > *kl.NumberOfKeys {
		return fmt.Errorf("NumberOfKeys: Is Less than the number of KeyValues to match. Will always result in 0 matches")
	}

	return nil
}

func (q *Query) MatchTags(tags datatypes.StringMap) bool {
	// quick checek to reject on the number of keys
	if q.Limits != nil && *q.Limits.NumberOfKeys < len(tags) {
		return false
	}

	for key, keysValue := range q.KeyValues {
		// existenc checks
		if keysValue.Exists != nil {
			switch *keysValue.Exists {
			case true:
				if tagValue, ok := tags[key]; ok {
					if keysValue.ExistsType != nil {
						if tagValue.DataType != *keysValue.ExistsType {
							// the required data type does not match so instantly fail
							return false
						}
					}
				} else {
					// tags doesn't contain a required key, instantly fail
					return false
				}
			default:
				// if the value exists in the provided tags, instantly fail
				if tagValue, ok := tags[key]; ok {
					if keysValue.ExistsType != nil {
						if tagValue.DataType == *keysValue.ExistsType {
							return false
						}
					} else {
						return false
					}
				}
			}
		}

		// specific value checks
		if keysValue.Value != nil {
			tagValue, ok := tags[key]

			// fail if the tag group doesn't have a required tag
			if !ok {
				return false
			}

			switch *keysValue.ValueComparison {
			case equals:
				if tagValue.Less(*keysValue.Value) || (*keysValue.Value).Less(tagValue) {
					return false
				}
			default:
				if keysValue.ValueTypeMatch != nil {
					switch *keysValue.ValueTypeMatch {
					case true:
						if tagValue.DataType != keysValue.Value.DataType {
							// data types don't match so fail
							return false
						}
					}
				}
				switch *keysValue.ValueComparison {
				case notEquals:
					if !tagValue.Less(*keysValue.Value) && !(*keysValue.Value).Less(tagValue) {
						return false
					}
				case lessThan:
					if !tagValue.Less(*keysValue.Value) {
						return false
					}
				case lessThanOrEqual:
					if !tagValue.Less(*keysValue.Value) {
						// the tag value is greater than the query value
						if (*keysValue.Value).Less(tagValue) {
							return false
						}

						// this is the equals case
					}
				case greaterThan:
					if !(*keysValue.Value).Less(tagValue) {
						return false
					}
				case greaterThanOrEqual:
					if !(*keysValue.Value).Less(tagValue) {
						// the query value is greater than the tag value
						if (tagValue).Less(*keysValue.Value) {
							return false
						}

						// this is the equals case
					}
				}
			}
		}
	}

	return true
}
