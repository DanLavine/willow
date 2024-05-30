package queryassociatedaction

import (
	"fmt"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// AssociatedActionQuery is used to query a number of various apis for the
// Willow services. Mos data is saved throught the common KeyValues for values like:
// Willow's queues, Limiter's Rule Overrides, Limiter's Counters and Locker's Locks.

// Query to find specific key values.
// NOTE: if all fields are empty, this is equevelent to a select all
type AssociatedActionQuery struct {
	Selection *Selection `json:"Selection,omitempty"`

	Or  []*AssociatedActionQuery `json:"Or,omitempty"`
	And []*AssociatedActionQuery `json:"And,omitempty"`
}

//	RETURNS:
//	- error - error describing any possible issues with the query and the steps to rectify them
//
// Validate ensures the CreateLockRequest has all required fields set
func (query *AssociatedActionQuery) Validate() *errors.ModelError {
	if query == nil {
		// this is the select all case
		return nil
	}

	// validate single query
	if query.Selection != nil {
		if err := query.Selection.Validate(); err != nil {
			return &errors.ModelError{Field: "Selection", Child: err.(*errors.ModelError)}
		}
	}

	// validate all OR joins
	for index, or := range query.Or {
		if or == nil {
			return &errors.ModelError{Field: fmt.Sprintf("Or[%d]", index), Err: fmt.Errorf("is invalid because it is nil")}
		}

		if err := or.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("Or[%d]", index), Child: err}
		}
	}

	// validate all ns joins
	for index, and := range query.And {
		if and == nil {
			return &errors.ModelError{Field: fmt.Sprintf("And[%d]", index), Err: fmt.Errorf("is invalid because it is nil")}
		}

		if err := and.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("And[%d]", index), Child: err}
		}
	}

	return nil
}

func StringToAssociatedActionQuery(id string) *AssociatedActionQuery {
	return &AssociatedActionQuery{
		Selection: &Selection{
			IDs: []string{id},
		},
	}
}

func KeyValuesToExactAssociatedActionQuery(keyValues datatypes.KeyValues) *AssociatedActionQuery {
	query := &AssociatedActionQuery{
		Selection: &Selection{
			MinNumberOfKeyValues: helpers.PointerOf(len(keyValues)),
			MaxNumberOfKeyValues: helpers.PointerOf(len(keyValues)),
			KeyValues:            SelectionKeyValues{},
		},
	}

	for key, value := range keyValues {
		query.Selection.KeyValues[key] = ValueQuery{
			Value:      value,
			Comparison: v1.Equals,
			TypeRestrictions: v1.TypeRestrictions{
				MinDataType: value.Type,
				MaxDataType: value.Type,
			},
		}
	}

	return query
}

// used to know if arbitrary tags, would be found from a query
func (q *AssociatedActionQuery) MatchTags(tags datatypes.KeyValues) bool {
	matched := true

	// if Query is not nil, check the tags
	if q.Selection != nil {
		// for each item in the tag, check to see if there is a key value guard
		// need to check the query as the source of truth...
		matched = q.Selection.MatchKeyValues(tags)
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
