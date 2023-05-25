package datatypes

// when trees require a slice of tree keys
type EnumerableCompareType interface {
	// pop the first index off of the Enumerable list and return it.
	// When there are no more values in the list, return nil
	Pop() (CompareType, EnumerableCompareType)

	// return the len(...) call of underlying type
	Len() int
}

// Generic type that can be used for tree keys
type CompareType interface {
	// Less will always pass in the compareKey == the original value's type
	Less(compareKey any) bool
}
