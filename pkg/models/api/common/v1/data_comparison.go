package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type DataComparison string

const (
	// DataComparison
	Equals             DataComparison = "="
	NotEquals          DataComparison = "!="
	LessThan           DataComparison = "<"
	LessThanOrEqual    DataComparison = "<="
	GreaterThan        DataComparison = ">"
	GreaterThanOrEqual DataComparison = ">="
)

func (dataComparison DataComparison) Validate() error {
	switch dataComparison {
	case Equals, NotEquals, LessThan, LessThanOrEqual, GreaterThan, GreaterThanOrEqual:
		// these are all fine
	default:
		return &errors.ModelError{Err: fmt.Errorf("unknown value '%s'", dataComparison)}
	}

	return nil
}
