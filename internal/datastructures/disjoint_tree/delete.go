package disjointtree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// dtCanDelete is used when finding the disjointNode value in the tree to remove.
// There are a couple of checks to make sure internal disjointNodes are not delete
// if they have children
func dtCanDelete(canDelete datastructures.CanDelete) datastructures.CanDelete {
	return func(item any) bool {
		node := item.(*disjointNode)

		// check to remove the node's value
		if canDelete == nil {
			// can always set the value to nil if no optional can delete check
			node.value = nil
		} else {
			// if the node has a value
			if node.value != nil {
				// only delete the value if it can be
				if canDelete(node.value) {
					node.value = nil
				}
			}
		}

		// cannot ever delete the disjointNode if there are children
		if node.children != nil {
			return false
		}

		// if we are here, we can delete this if the value is nil because we know there are no children
		return node.value == nil
	}
}

// Delete removes a value from the disjoint tree iff it exists
//
// PARAMS:
// * keys - the set of keys for an item in the tree
// * onFind - (optional) callback to run when the item is found in the tree
//
// RETURNS
// * any - item found in the tree. If the item is not found this will be nil
// * error - any errors related to keys provided (such as nil, or an empty set)
func (dt *disjointTree) Delete(keys datatypes.EnumerableCompareType, canDelete datastructures.CanDelete) error {
	if keys == nil || keys.Len() <= 0 {
		return fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}

	dt.delete(keys, canDelete)
	return nil
}

func (dt *disjointTree) delete(keys datatypes.EnumerableCompareType, canDelete datastructures.CanDelete) {
	key, keys := keys.Pop()

	// we are at the final index
	if keys.Len() == 0 {
		dt.tree.Delete(key, dtCanDelete(canDelete))
		if dt.tree.Empty() {
			dt.tree = nil
		}

		return
	}

	// we are recursing
	item := dt.tree.Find(key, dtLock)
	if item == nil {
		// keys don't exist
		return
	}
	node := item.(*disjointNode)
	defer node.lock.Unlock()

	// remove the child key
	node.children.delete(keys, canDelete)

	// check to see if we can also remove this value from the current disjoint tree
	if node.children.tree == nil && node.value == nil {
		dt.tree.Delete(key, nil)
		if dt.tree.Empty() {
			dt.tree = nil
		}
	}
}
