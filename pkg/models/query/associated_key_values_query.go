package query

import (
	"encoding/json"
	"fmt"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type AssociatedKeyValuesQuery struct {
	KeyValueSelection *KeyValueSelection

	Or  []AssociatedKeyValuesQuery
	And []AssociatedKeyValuesQuery
}

func ParseAssociatedKeyValuesQuery(data []byte) (*AssociatedKeyValuesQuery, error) {
	Queryion := &AssociatedKeyValuesQuery{}
	if err := json.Unmarshal(data, Queryion); err != nil {
		return nil, err
	}

	return Queryion, nil
}

func (q *AssociatedKeyValuesQuery) Validate() error {
	// validate single query
	if q.KeyValueSelection != nil {
		if err := q.KeyValueSelection.Validate(); err != nil {
			return fmt.Errorf("KeyValueSelection%s", err.Error())
		}
	}

	// validate all or joins
	for index, or := range q.Or {
		if err := or.Validate(); err != nil {
			return fmt.Errorf("Or[%d].%s", index, err.Error())
		}
	}

	// validate all and joins
	for index, and := range q.And {
		if err := and.Validate(); err != nil {
			return fmt.Errorf("And[%d].%s", index, err.Error())
		}
	}

	return nil
}

// used to know if arbitrary tags, would be found from a query
func (q *AssociatedKeyValuesQuery) MatchTags(tags datatypes.KeyValues) bool {
	matched := true

	// if Query is not nil, check the tags
	if q.KeyValueSelection != nil {
		// for each item in the tag, check to see if there is a key value guard
		// need to check the query as the source of truth...
		matched = q.KeyValueSelection.MatchTags(tags)
	}

	// for each and, need to intersect that all those values match as well
	if matched && q.And != nil {
		for _, andCheck := range q.And {
			if !andCheck.MatchTags(tags) {
				matched = false
				break
			}

		}
	}

	// can bail out early here and don't even need to check the OR values
	if matched {
		return true
	}

	for _, orValue := range q.Or {
		// if any of these match, can return true
		if orValue.MatchTags(tags) {
			return true
		}
	}

	return false
}
