package btreeassociated

import (
	"github.com/DanLavine/willow/internal/datastructures"
)

// Iterate over each value saved in the associated tree
//
// PARAMS:
// * callback - function is called when a Tree's Node value != nil. The Iterate callback is passed the Node's value
func (at *associatedTree) Iterate(callback datastructures.Iterate) {
	if callback == nil {
		panic("callback is nil")
	}

	// NOTE: this actually means we just need to iterate over the ID tree since that stores the actual values
	at.idTree.Iterate(callback)
}
