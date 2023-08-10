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
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.iterate(callback)
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) iterate(callback datastructures.OnFind) {
	for i := 0; i < bn.numberOfValues; i++ {
		callback(bn.keyValues[i].value)
	}

	for i := 0; i < bn.numberOfChildren; i++ {
		bn.children[i].lock.RLock()
	}

	// have a lock on all the children at this point, can release the lock at this level
	bn.lock.RUnlock()

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
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.iterateMatchType(dataType, callback)
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) iterateMatchType(dataType datatypes.DataType, callback datastructures.OnFind) {
	startIndex := -1
	var i int
	for i = 0; i < bn.numberOfValues; i++ {
		// the key in the tree is less then the value we are looking for, iterate to the next value if it exists
		if bn.keyValues[i].key.DataType.Less(dataType) {
			continue
		}

		// the key we are searching for is less than the key in the tree. Try the less than tree and return
		if dataType.Less(bn.keyValues[i].key.DataType) {
			if startIndex == -1 {
				startIndex = i
			}

			break
		}

		// at this point, we know that the DataType's match
		if startIndex == -1 {
			startIndex = i
		}

		// run the callback
		callback(bn.keyValues[i].value)

		// always attempt a recurse on the less than nodes to find all values
		if bn.numberOfChildren != 0 {
			bn.children[i].lock.RLock()
		}
	}

	// always lock the index we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[i].lock.RLock()
	}

	// have a lock on all the children at this point, can release the lock at this level
	bn.lock.RUnlock()

	// if we go here, run one last check on the last greater than child tree
	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// key must be grater than all values we checked, must be on the greater than side
			bn.children[i].iterateMatchType(dataType, callback)
		} else {
			// need to recurse to all potential children from the start index
			for index := startIndex; index <= i; index++ {
				bn.children[index].iterateMatchType(dataType, callback)
			}
		}
	}
}
