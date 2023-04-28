package disjointtree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// CreateOrFind creates a value in the disjoint tree or possibly updates the value iff it is an internal node
//
// PARAMS:
// * keys - the set of keys for an item in the tree
// * onFind - (optional) callback to run when the item is found in the tree
// * onCreate - callback to run when the when creating the initial item to save in the tree
//
// RETURNS
// * any - item found in the tree. If the item is not found this will be nil
// * error - any errors related to keys provided (such as nil, or an empty set)
func (dt *disjointTree) CreateOrFind(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind, onCreate datastructures.OnCreate) (any, error) {
	if keys == nil || keys.Len() <= 0 {
		return nil, fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}
	if onCreate == nil {
		return nil, fmt.Errorf("Received a nil onCreate callback. Needs to not be nil")
	}

	return dt.createOrFind(keys, onFind, onCreate)
}

func (dt *disjointTree) createOrFind(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind, onCreate datastructures.OnCreate) (any, error) {
	key, keys := keys.Pop()

	if key == nil {
		return nil, fmt.Errorf("Received an invalid EnumberableTreeKeys. Can not have a nil value")
	}

	// we are at the last index
	if keys.Len() == 0 {
		// create the item or find the item if it does not currently exist
		treeItem, err := dt.tree.CreateOrFind(key, dtLock, dt.newDisjointNodeWithValue(onCreate))
		if err != nil {
			return nil, err
		}
		disjointNode := treeItem.(*disjointNode)
		defer disjointNode.lock.Unlock()

		if disjointNode.value == nil {
			// this is an update for a node that was created, but didn't have an original value.
			// similar to an OnCreate since the value was not set
			value, err := onCreate()
			if err != nil {
				return nil, err
			}

			disjointNode.value = value
		} else {
			// This is an on find call
			if onFind != nil {
				onFind(disjointNode.value)
			}
		}

		return disjointNode.value, nil
	}

	// recurse
	treeItem, err := dt.tree.CreateOrFind(key, dtLock, newDisjointNodeEmpty)
	if err != nil {
		return nil, err
	}
	disjointNode := treeItem.(*disjointNode)
	defer disjointNode.lock.Unlock()

	if disjointNode.children == nil {
		// create a new child if there currently are none
		disjointNode.children = New()
	}

	return disjointNode.children.createOrFind(keys, onFind, onCreate)
}
