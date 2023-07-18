package btree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
)

// Iterate over each node with a thread safe read lock and call the iterate function when the value != nil
//
// PARAMS:
// - callback - function is called when a Tree's Node value != nil. The Iterate callback is passed the Node's value
func (btree *threadSafeBTree) Iterate(callback datastructures.OnFind) error {
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	if btree.root != nil {
		btree.root.iterate(callback)
	}

	return nil
}

func (bn *threadSafeBNode) iterate(callback datastructures.OnFind) {
	for i := 0; i < bn.numberOfValues; i++ {
		callback(bn.keyValues[i].value)
	}

	for i := 0; i < bn.numberOfChildren; i++ {
		bn.children[i].iterate(callback)
	}
}
