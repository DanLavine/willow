package disjointtree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Find returns a value iff it exists within the disjoint tree
//
// PARAMS:
// * keys - the set of keys for an item in the tree
// * onFind - (optional) callback to run when the item is found in the tree
//
// RETURNS
// * any - item found in the tree. If the item is not found this will be nil
// * error - any errors related to keys provided (such as nil, or an empty set)
func (dt *disjointTree) Find(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind) (any, error) {
	if keys == nil || keys.Len() <= 0 {
		return nil, fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}

	return dt.find(keys, onFind)
}

func (dt *disjointTree) find(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind) (any, error) {
	key, keys := keys.Pop()

	if key == nil {
		return nil, fmt.Errorf("Received an invalid EnumberableTreeKeys. Can not have a nil value")
	}

	// find the tree item
	treeItem := dt.tree.Find(key, dtReadLock)
	if treeItem == nil {
		return nil, nil
	}
	disjointNode := treeItem.(*disjointNode)
	defer disjointNode.lock.RUnlock()

	// we are at the final index
	if keys.Len() == 0 {
		if onFind != nil {
			onFind(disjointNode.value)
		}

		return disjointNode.value, nil
	}

	// recurse
	if disjointNode.children == nil {
		return nil, nil
	}

	return disjointNode.children.find(keys, onFind)
}
