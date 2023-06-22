package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
)

// Iterate over each value saved in the associated tree
//
// PARAMS:
// * callback - function is called when a Tree's Node value != nil. The Iterate callback is passed the Node's value
func (at *threadsafeAssociatedTree) Iterate(callback datastructures.OnFind) error {
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}

	at.lock.RLock()
	defer at.lock.RUnlock()

	// NOTE:  we just need to iterate over the ID tree since that stores the actual values
	if at.idTree != nil {
		at.idTree.Iterate(callback)
	}

	return nil
}
