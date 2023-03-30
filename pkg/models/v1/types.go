package v1

import (
	"sort"

	"github.com/DanLavine/willow/internal/datastructures"
)

// Satisfy disjoint Set
type Tags []String

func (ts Tags) Each(callback func(int, datastructures.TreeKey) bool) {
	for index, tag := range ts {
		if !callback(index, tag) {
			break
		}
	}
}

func (ts Tags) Sort() {
	sort.SliceStable(ts, func(i, j int) bool {
		return ts[i] < ts[j]
	})
}

func (ts Tags) Len() int {
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
