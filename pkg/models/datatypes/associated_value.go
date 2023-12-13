package datatypes

import (
	"fmt"
)

type Comparison string

var (
	// I think this makes sense for a smarter query operation where a user wants to enforce the type of a specific key
	// With other NoSQL dbs like MongoDB, I only ever see a feature for type assertion by making 2 statements joined by 'and'.
	// 1 for the comparions (i.e. <)
	// 1 for the type assertion (I.E string)
	// But that seems like unecessary work and slower
	equals                      Comparison = "="
	notEquals                   Comparison = "!="
	lessThan                    Comparison = "<"
	lessThanMatchType           Comparison = "< MATCH"
	lessThanOrEqual             Comparison = "<="
	lessThanOrEqualMatchType    Comparison = "<= MATCH"
	greaterThan                 Comparison = ">"
	greaterThanMatchType        Comparison = "> MATCH"
	greaterThanOrEqual          Comparison = ">="
	greaterThanOrEqualMatchType Comparison = ">= MATCH"
)

func Equals() Comparison    { return equals }
func NotEquals() Comparison { return notEquals }

func EqualsPtr() *Comparison    { return &equals }
func NotEqualsPtr() *Comparison { return &notEquals }

func LessThan() Comparison                    { return lessThan }
func LessThanMatchType() Comparison           { return lessThanMatchType }
func LessThanOrEqual() Comparison             { return lessThanOrEqual }
func LessThanOrEqualMatchType() Comparison    { return lessThanOrEqualMatchType }
func GreaterThan() Comparison                 { return greaterThan }
func GreaterThanMatchType() Comparison        { return greaterThanMatchType }
func GreaterThanOrEqual() Comparison          { return greaterThanOrEqual }
func GreaterThanOrEqualMatchType() Comparison { return greaterThanOrEqualMatchType }

func LessThanPtr() *Comparison                    { return &lessThan }
func LessThanMatchTypePtr() *Comparison           { return &lessThanMatchType }
func LessThanOrEqualPtr() *Comparison             { return &lessThanOrEqual }
func LessThanOrEqualMatchTypePtr() *Comparison    { return &lessThanOrEqualMatchType }
func GreaterThanPtr() *Comparison                 { return &greaterThan }
func GreaterThanMatchTypePtr() *Comparison        { return &greaterThanMatchType }
func GreaterThanOrEqualPtr() *Comparison          { return &greaterThanOrEqual }
func GreaterThanOrEqualMatchTypePtr() *Comparison { return &greaterThanOrEqualMatchType }

// DSL TODO: Do these have to be pointers? Doing the convertion between api request -> query can be confusing at times.
type Value struct {
	// check to see if a key exists
	Exists *bool
	// check to use for a particualr type
	ExistsType *DataType

	// specific value for a particular key
	Value *EncapsulatedValue
	// what type of comparison to run on a particualr value [=, !=, <, < MATCH, <=, <= MATCH, >, > MATCH, >=, >= MATCH]
	ValueComparison *Comparison
}

func (v *Value) validate() error {
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
		case equals, notEquals, lessThan, lessThanMatchType, lessThanOrEqual, lessThanOrEqualMatchType, greaterThan, greaterThanMatchType, greaterThanOrEqual, greaterThanOrEqualMatchType:
			// nothing to do here. These are all the valid cases
		default:
			return fmt.Errorf(".ValueComparison: received an unexpected key")
		}
	}

	return nil
}

func (v *Value) validateReservedKey() error {
	if v.Exists != nil {
		return fmt.Errorf(": cannot be an existence check. It can only mach an exact string")
	}

	if v.ExistsType != nil {
		return fmt.Errorf(": cannot check an existence type. It can only mach an exact string")
	}

	if v.Value == nil {
		return fmt.Errorf(": requires a string Value to match against")
	}

	if v.Value.DataType() != T_string {
		return fmt.Errorf(": requires a string Value to match against")
	}

	if v.ValueComparison != EqualsPtr() {
		return fmt.Errorf(": requires an Equals ValueComparison")
	}

	return nil
}
