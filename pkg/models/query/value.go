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

func Equals() Comparison             { return equals }
func NotEquals() Comparison          { return notEquals }
func LessThan() Comparison           { return lessThan }
func LessThanOrEqual() Comparison    { return lessThanOrEqual }
func GreaterThan() Comparison        { return greaterThan }
func GreaterThanOrEqual() Comparison { return greaterThanOrEqual }

func EqualsPtr() *Comparison             { return &equals }
func NotEqualsPtr() *Comparison          { return &notEquals }
func LessThanPtr() *Comparison           { return &lessThan }
func LessThanOrEqualPtr() *Comparison    { return &lessThanOrEqual }
func GreaterThanPtr() *Comparison        { return &greaterThan }
func GreaterThanOrEqualPtr() *Comparison { return &greaterThanOrEqual }

type Value struct {
	Exists     *bool
	ExistsType *datatypes.DataType

	Value           *datatypes.EncapsulatedData
	ValueComparison *Comparison

	// Only in use for ranged operations [!=, <, <=, >, >=]
	// When set to true, will ensure that we select only values that match
	// the data type ov 'Value'
	//
	// TODO: change from *bool -> bool
	ValueTypeMatch *bool
}

func (v *Value) Validate() error {
	if v.Exists == nil && v.Value == nil {
		return fmt.Errorf(": Requires an Exists or Value check")
	}

	// validate existance
	if v.Exists != nil {
		if v.Value != nil {
			return fmt.Errorf(": Can only contain a single Exists or Value check, not both")
		}

		if v.ValueComparison != nil {
			return fmt.Errorf(": ValueComparison is provided, but is incompatible with Exists check")
		}

		if v.ValueTypeMatch != nil {
			return fmt.Errorf(": ValueTypeMatch is provided, but is incompatible with Exists check")
		}

		if v.ExistsType != nil {
			switch *v.ExistsType {
			case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14:
				// these are all valid
			default:
				return fmt.Errorf(": Unexpected ExistsType. Must be an int from 1-14 inclusive")
			}
		}
	}

	// validate values
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
