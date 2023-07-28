package btree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
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

// Iterate over all the values for a given type
func (btree *threadSafeBTree) IterateMatchType(dataType datatypes.DataType, callback datastructures.OnFind) error {
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	if btree.root != nil {
		btree.root.iterateMatchType(dataType, callback)
	}

	return nil
}

func (bn *threadSafeBNode) iterateMatchType(dataType datatypes.DataType, callback datastructures.OnFind) {
	var i int
	for i = 0; i < bn.numberOfValues; i++ {
		// the key in the tree is less then the value we are looking for, iterate to the next value if it exists
		if bn.keyValues[i].key.DataType.Less(dataType) {
			continue
		}

		// the key we are searching for is less than the key in the tree. Try the less than tree and return
		if dataType.Less(bn.keyValues[i].key.DataType) {
			if bn.numberOfChildren != 0 {
				bn.children[i].iterateMatchType(dataType, callback)
			}

			return
		}

		// at this point, we know that the DataType's match

		// run the callback
		callback(bn.keyValues[i].value)

		// if there is a less than child, run that
		if bn.numberOfChildren != 0 {
			bn.children[i].iterateMatchType(dataType, callback)
		}
	}

	// if we go here, run one last check on the last greater than child tree
	if bn.numberOfChildren != 0 {
		bn.children[i].iterateMatchType(dataType, callback)
	}
}
