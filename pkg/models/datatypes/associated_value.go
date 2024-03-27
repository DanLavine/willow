package datatypes

import (
	"fmt"
)

type Comparison string

var (
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

/*
	'key1' = Any() && 'key2' EXISTS // theses I think are the same
	'key1' != Any() && 'key2' NOT EXISTS // theses I think are the same
	'key2' = MATCH Type(int) // so in this case, we only care about the type of data to find
	'key2' > MATCH String("foo") // find all values that are greate than the string foo where the types match
	'key2' > String("foo") // find all values that are greate than the string foo including any of the types
*/

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

// Value to check for with the data types
type AssociatedValue struct {
	// value that is used for the comparison
	Value EncapsulatedValue

	// what type of comparison to run on a particualr Type and Data [=, !=, <, < MATCH, <=, <= MATCH, >, > MATCH, >=, >= MATCH]
	ValueComparison Comparison
}

func (v *AssociatedValue) validate() error {
	if err := v.Value.Validate(); err != nil {
		return fmt.Errorf(".Value: %w", err)
	}

	if v.Value.DataType() == T_any {
		switch v.ValueComparison {
		case equals, notEquals:
			// these are the only alowed values for the any type
		default:
			return fmt.Errorf(": Can only have [= | !=] for an Any value")
		}
	}

	switch v.ValueComparison {
	case equals, notEquals, lessThan, lessThanMatchType, lessThanOrEqual, lessThanOrEqualMatchType, greaterThan, greaterThanMatchType, greaterThanOrEqual, greaterThanOrEqualMatchType:
		// these are all valid
	default:
		return fmt.Errorf(".ValueComparison: received an unexpected value")
	}

	return nil
}

// func (v *Value) validateReservedKey() error {
// 	if v.Exists != nil {
// 		return fmt.Errorf(": cannot be an existence check. It can only mach an exact string")
// 	}

// 	if v.ExistsType != nil {
// 		return fmt.Errorf(": cannot check an existence type. It can only mach an exact string")
// 	}

// 	if v.Value == nil {
// 		return fmt.Errorf(": requires a string Value to match against")
// 	}

// 	if v.Value.DataType() != T_string {
// 		return fmt.Errorf(": requires a string Value to match against")
// 	}

// 	if v.ValueComparison != EqualsPtr() {
// 		return fmt.Errorf(": requires an Equals ValueComparison")
// 	}

// 	return nil
// }
