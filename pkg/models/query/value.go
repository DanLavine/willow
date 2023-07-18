package query

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Comparison string

var (
	equals             Comparison = "="
	notEquals          Comparison = "!="
	lessThan           Comparison = "<"
	lessThanOrEqual    Comparison = "<="
	greaterThan        Comparison = ">"
	greaterThanOrEqual Comparison = ">="
)

func Equals() *Comparison             { return &equals }
func NotEquals() *Comparison          { return &notEquals }
func LessThan() *Comparison           { return &lessThan }
func LessThanOrEqual() *Comparison    { return &lessThanOrEqual }
func GreaterThan() *Comparison        { return &greaterThan }
func GreaterThanOrEqual() *Comparison { return &greaterThanOrEqual }

type Value struct {
	Exists *bool

	Value           *datatypes.EncapsulatedData
	ValueComparison *Comparison
}

func (v *Value) Validate() error {
	if v.Exists == nil && v.Value == nil {
		return fmt.Errorf(": Requires an Exists or Value check")
	}

	if v.Exists != nil && v.Value != nil {
		return fmt.Errorf(": Can only contain a single Exists or Value check, not both")
	}

	if v.Value != nil {
		if v.ValueComparison == nil {
			return fmt.Errorf(".ValueComparison: is required for a Value")
		}

		switch *v.ValueComparison {
		case equals, notEquals, lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual:
			// nothing to do here. These are all the valid cases
		default:
			return fmt.Errorf(".ValueComparison: received an unexpected key")
		}
	}

	return nil
}
