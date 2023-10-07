package query

import (
	"fmt"
	"sort"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

const ReservedID = "_associated_id"

type KeyValueSelection struct {
	// Key Values ensures that all key values exist
	KeyValues map[string]Value

	// limits for collections to look for
	Limits *KeyLimits
}

type KeyLimits struct {
	// limit how many keys can make up a collection
	NumberOfKeys *int
}

func (kvs *KeyValueSelection) SortedKeys() []string {
	keys := []string{}
	for key, _ := range kvs.KeyValues {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func (kvs *KeyValueSelection) Validate() error {
	if len(kvs.KeyValues) == 0 && kvs.Limits == nil {
		return fmt.Errorf(": requires KeyValues or Limits parameters")
	}

	if kvs.Limits != nil {
		if _, ok := kvs.KeyValues[ReservedID]; ok {
			if err := kvs.Limits.Validate(len(kvs.KeyValues) - 1); err != nil {
				return fmt.Errorf(".Limits.%s", err.Error())
			}
		} else {
			if err := kvs.Limits.Validate(len(kvs.KeyValues)); err != nil {
				return fmt.Errorf(".Limits.%s", err.Error())
			}
		}
	}

	for key, value := range kvs.KeyValues {
		if key == ReservedID {
			if err := value.validateReservedKey(); err != nil {
				return fmt.Errorf(".KeyValues[%s]%s", key, err.Error())
			}
		}

		if err := value.validate(); err != nil {
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

func (kvs *KeyValueSelection) MatchTags(tags datatypes.StringMap) bool {
	// quick checek to reject on the number of keys
	if kvs.Limits != nil && *kvs.Limits.NumberOfKeys < len(tags) {
		return false
	}

	for key, keysValue := range kvs.KeyValues {
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
				switch *keysValue.ValueComparison {
				case notEquals:
					if !tagValue.Less(*keysValue.Value) && !(*keysValue.Value).Less(tagValue) {
						return false
					}
				case lessThan:
					if !tagValue.Less(*keysValue.Value) {
						return false
					}
				case lessThanMatchType:
					if !tagValue.LessType(*keysValue.Value) && !keysValue.Value.LessType(tagValue) {
						if !tagValue.Less(*keysValue.Value) {
							return false
						}
					} else {
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
				case lessThanOrEqualMatchType:
					if !tagValue.LessType(*keysValue.Value) && !keysValue.Value.LessType(tagValue) {
						if !tagValue.Less(*keysValue.Value) {
							// the tag value is greater than the query value
							if (*keysValue.Value).Less(tagValue) {
								return false
							}

							// this is the equals case
						}
					} else {
						return false

					}
				case greaterThan:
					if !(*keysValue.Value).Less(tagValue) {
						return false
					}
				case greaterThanMatchType:
					if !tagValue.LessType(*keysValue.Value) && !keysValue.Value.LessType(tagValue) {
						if !(*keysValue.Value).Less(tagValue) {
							return false
						}
					} else {
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
				case greaterThanOrEqualMatchType:
					if !tagValue.LessType(*keysValue.Value) && !keysValue.Value.LessType(tagValue) {
						if !(*keysValue.Value).Less(tagValue) {
							// the query value is greater than the tag value
							if (tagValue).Less(*keysValue.Value) {
								return false
							}

							// this is the equals case
						}
					} else {
						return false
					}

				}
			}
		}
	}

	return true
}
