package query

import (
	"fmt"
	"sort"
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
		return fmt.Errorf("Requires KeyValues or Limits parameters")
	}

	if q.Limits != nil {
		if err := q.Limits.Validate(len(q.KeyValues)); err != nil {
			return fmt.Errorf("Limits.%s", err.Error())
		}
	}

	for key, value := range q.KeyValues {
		if err := value.Validate(); err != nil {
			return fmt.Errorf("KeyValues[%s]%s", key, err.Error())
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
