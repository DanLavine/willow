package btree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Find a tree item with a thread safe read locks/
// PARAMS:
// * key - key to use when searching for the associated value
// * onFind - function to call when finding the item. Use the empty string to not call anything
func (btree *bTree) Find(key datatypes.CompareType, onFind datastructures.OnFind) any {
	if key == nil {
		panic("key is nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	if btree.root == nil {
		return nil
	}

	return btree.root.find(key, onFind)
}

// find an item from the tree if it exists
func (bn *bNode) find(key datatypes.CompareType, onFind datastructures.OnFind) any {
	for index := 0; index < bn.numberOfValues; index++ {
		value := bn.values[index]

		if !value.key.Less(key) {
			if !key.Less(value.key) {
				if onFind != nil {
					onFind(value.item)
				}

				return value.item
			}

			if bn.numberOfChildren == 0 {
				return nil
			}

			return bn.children[index].find(key, onFind)
		}
	}

	if bn.numberOfChildren == 0 {
		return nil
	}

	return bn.children[bn.numberOfChildren-1].find(key, onFind)
}
