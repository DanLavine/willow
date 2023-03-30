package datastructures

import (
	"fmt"
	"reflect"
	"sync"
)

// Disjoint tree structure create trees with any number of child trees. Each Disjoint Tree is a BTree
// where each value in the tree is a node containing a possible value and child. The Trees are recursive in nature when
// creating multiple tags.
//
// I.E: The Root tree could be structured like so
//
//	   Cow
//	 /      \
//	Bat     Farm
//
// Then Under Farm there could be another tree structure of:
//
//	    Group
//	  /        \
//	Berries   HayStacks
//
// On any Operations (Find, Create, Delete, etc). when calling []string{"Farm", "HayStack"} we can look at
// the root level for "Farm" and then recurse into the tree at that level to find "HayStack" for example
type DisjointTree interface {
	// Find the provided tree item if it already exists. retrns nil if not found
	Find(keys EnumerableTreeKeys, onFind string) (any, error)

	// Find the provided tree item if it already exists. Or return the newly inserted tree item
	// TODO rename this to UpdateOrCreate?
	FindOrCreate(keys EnumerableTreeKeys, onFind string, onCreate func() (any, error)) (any, error)

	// Iterate over the tree and for each value found invoke the callback with the node's value iff != nil
	Iterate(callback func(value any))
}

// tree structure holding all nodes
type disjointTree struct {
	//every item in the tree is of type *disjointNode
	tree BTree
}

func NewDisjointTree() *disjointTree {
	tree, err := NewBTree(2)
	if err != nil {
		panic(err)
	}

	return &disjointTree{
		tree: tree,
	}
}

// wraper to create a new disjoint node with a value
func (dt *disjointTree) newDisjointNodeWithValue(onCreate func() (any, error)) func() (any, error) {
	return func() (any, error) {
		lock := new(sync.RWMutex)
		lock.Lock()

		value, err := onCreate()
		if err != nil {
			lock.Unlock()
			return nil, err
		}

		return &disjointNode{
			lock:     lock,
			value:    value,
			children: nil,
		}, nil
	}
}

// wrapper to create a new disjoint node without a value
func (dt *disjointTree) newDisjointNode() (any, error) {
	lock := new(sync.RWMutex)
	lock.Lock()

	return &disjointNode{
		lock:     lock,
		value:    nil,
		children: nil,
	}, nil
}

// each node in a disjoint tree
type disjointNode struct {
	lock     *sync.RWMutex
	value    any
	children *disjointTree
}

func (dt *disjointNode) OnFind() {
	dt.lock.RLock()
}

func (dt *disjointNode) OnUpdate() {
	dt.lock.Lock()
}

func (dt *disjointTree) Find(keys EnumerableTreeKeys, onFind string) (any, error) {
	if keys == nil {
		return nil, fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}

	var err error
	var value any
	tree := dt.tree
	size := keys.Len() - 1

	if size < 0 {
		return nil, fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}

	keys.Each(func(index int, key TreeKey) bool {
		if key == nil {
			err = fmt.Errorf("Received an invalid EnumberableTreeKeys. Can not have a nil value")
			return false
		}

		treeItem := tree.Find(key, "OnFind")
		if treeItem == nil {
			err = fmt.Errorf("item not found")
			return false
		}

		if index == size {
			// create the item or find the item if it does not currently exist
			disjointNode := treeItem.(*disjointNode)
			defer disjointNode.lock.RUnlock()

			if onFind != "" {
				_ = reflect.ValueOf(disjointNode.value).MethodByName(onFind).Call(nil)
			}

			value = disjointNode.value
			return false
		} else {
			disjointNode := treeItem.(*disjointNode)
			defer disjointNode.lock.RUnlock()

			tree = disjointNode.children.tree
		}

		return true
	})

	return value, err
}

// FindORCreate is a thread safe way to create, find or update items in the DisjointTree
// All Operations are currently guarded by an exclusive lock to save on memory... which might want to change
// in the future
func (dt *disjointTree) FindOrCreate(keys EnumerableTreeKeys, onFind string, onCreate func() (any, error)) (any, error) {
	if keys == nil {
		return nil, fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}

	if onCreate == nil {
		return nil, fmt.Errorf("Received a nil onCreate callback. Needs to not be nil")
	}

	var returnErr error
	var returnValue any
	size := keys.Len() - 1
	tree := dt.tree

	if size < 0 {
		return nil, fmt.Errorf("EnumberableTreeKeys must have at least 1 element")
	}

	keys.Each(func(index int, key TreeKey) bool {
		if key == nil {
			returnErr = fmt.Errorf("Received an invalid EnumberableTreeKeys. Can not have a nil value")
			return false
		}

		// last index so create case
		if index == size {
			// create the item or find the item if it does not currently exist
			treeItem, err := tree.FindOrCreate(key, "OnUpdate", dt.newDisjointNodeWithValue(onCreate))
			if err != nil {
				returnErr = err
				return false
			}

			disjointNode := treeItem.(*disjointNode)
			defer disjointNode.lock.Unlock()

			if disjointNode.value == nil {
				// this is an update for a node that was added that didn't have an original value
				// similar to an OnCreate since the value was not set
				value, err := onCreate()
				if err != nil {
					returnErr = err
					return false
				}

				disjointNode.value = value
				returnValue = value
				return false
			} else {
				// This is an on find call
				if onFind != "" {
					_ = reflect.ValueOf(disjointNode.value).MethodByName(onFind).Call(nil)
				}
			}

			returnValue = disjointNode.value
			return false
		}

		// keep recursing
		treeItem, err := tree.FindOrCreate(key, "OnUpdate", dt.newDisjointNode)
		if err != nil {
			returnErr = err
			return false
		}

		disjointNode := treeItem.(*disjointNode)
		defer disjointNode.lock.Unlock()

		if disjointNode.children == nil {
			disjointNode.children = NewDisjointTree()
		}

		tree = disjointNode.children.tree
		return true
	})

	return returnValue, returnErr
}

func (dt *disjointTree) Iterate(callback func(value any)) {
	if callback == nil {
		panic("callback is nil")
	}

	iterator := func(value any) {
		node := value.(*disjointNode)
		if node.value != nil {
			callback(node.value)
		}

		if node.children != nil {
			node.children.Iterate(callback)
		}
	}

	dt.tree.Iterate(iterator)
}
