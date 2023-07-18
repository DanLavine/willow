package query

import (
	"encoding/json"
	"fmt"
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
			return fmt.Errorf("Where.%s", err.Error())
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
