package datatypes

import "sort"

type String string

// compare an item with the same type as itself
func (s String) Less(compare any) bool {
	return s < compare.(String)
}

func (s String) ToString() string {
	return string(s)
}

// Satisfy disjoint Set
type Strings []String

func (s Strings) Pop() (CompareType, EnumerableCompareType) {
	return s[0], s[1:]
}

// Sort a slice of strings into ascending order
func (s Strings) Sort() {
	sort.SliceStable(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

// Len returns the len() call of the wrapped type
func (ts Strings) Len() int {
	return len(ts)
}
