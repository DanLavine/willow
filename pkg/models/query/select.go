package query

import (
	"encoding/json"
	"fmt"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Select struct {
	Where *Query

	Or  []Select
	And []Select
}

func ParseSelect(data []byte) (*Select, error) {
	selection := &Select{}
	if err := json.Unmarshal(data, selection); err != nil {
		return nil, err
	}

	return selection, nil
}

func (s *Select) Validate() error {
	// validate single query
	if s.Where != nil {
		if err := s.Where.Validate(); err != nil {
			return fmt.Errorf("Where%s", err.Error())
		}
	}

	// validate all or joins
	for index, or := range s.Or {
		if err := or.Validate(); err != nil {
			return fmt.Errorf("Or[%d].%s", index, err.Error())
		}
	}

	// validate all and joins
	for index, and := range s.And {
		if err := and.Validate(); err != nil {
			return fmt.Errorf("And[%d].%s", index, err.Error())
		}
	}

	return nil
}

// used to know if arbitrary tags, match a query
func (s *Select) MatchTags(tags datatypes.StringMap) bool {
	matched := true

	// if Where is not nil, check the tags
	if s.Where != nil {
		// for each item in the tag, check to see if there is a key value guard
		// need to check the query as the source of truth...
		matched = s.Where.MatchTags(tags)
	}

	// for each and, need to intersect that all those values match as well
	if matched && s.And != nil {
		for _, andCheck := range s.And {
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

	for _, orValue := range s.Or {
		// if any of these match, can return true
		if orValue.MatchTags(tags) {
			return true
		}
	}

	return false
}
