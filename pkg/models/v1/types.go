package v1

import (
	"sort"

	"github.com/DanLavine/willow/internal/datastructures"
)

// Satisfy disjoint Set
type Strings []String

// Each can be used to call a functionfor each element in the slice where
// the paramanters are the index + element of the slice
func (s Strings) Each(callback func(int, datastructures.TreeKey) bool) {
	for index, tag := range s {
		if !callback(index, tag) {
			break
		}
	}
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

// Satisfy BTree
type String string

// compare an item with the same type as itself
func (s String) Less(compare any) bool {
	return s < compare.(String)
}

func (s String) ToString() string {
	return string(s)
}
