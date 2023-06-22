package btree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Find the item in the Tree and run the `OnFind(...)` function for the saved value. Will not be called if the
// key cannot be found
//
// PARAMS:
// - key - key to use when comparing to other possible items
// - onFind - method to call if the key is found
//
// RETURNS:
// - error - any errors encontered. I.E. key is not valid
func (btree *threadSafeBTree) Find(key datatypes.EncapsulatedData, onFind datastructures.OnFind) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if onFind == nil {
		return fmt.Errorf("onFind is nil, but a value is required")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	if btree.root != nil {
		btree.root.find(key, onFind)
	}

	return nil
}

// find an item from the tree if it exists
func (bn *threadSafeBNode) find(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	for index := 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if !keyValue.key.Less(key) {
			// this is an exact match for the key
			if !key.Less(keyValue.key) {
				onFind(keyValue.value)
			}

			// key must be on a child, so recurse down
			if bn.numberOfChildren != 0 {
				bn.children[index].find(key, onFind)
			}
		}
	}

	// if there are children, the value must be on the greater than path
	if bn.numberOfChildren != 0 {
		bn.children[bn.numberOfChildren-1].find(key, onFind)
	}
}
