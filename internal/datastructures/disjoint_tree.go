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
	Find(keys []TreeKey, onFind string) (any, error)

	// Find the provided tree item if it already exists. Or return the newly inserted tree item
	// TODO rename this to UpdateOrCreate?
	FindOrCreate(keys []TreeKey, onFind string, onCreate func() (any, error)) (any, error)
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

func (dt *disjointNode) OnUpdate() {
	dt.lock.Lock()
}
func (dt *disjointNode) OnFind() {
	dt.lock.RLock()
}

func (dt *disjointTree) Find(keys []TreeKey, onFind string) (any, error) {
	switch len(keys) {
	case 0:
		return nil, fmt.Errorf("Received an invalid keys length. Needs to be at least 1")
	case 1:
		// create the item or find the item if it does not currently exist
		treeItem := dt.tree.Find(keys[0], "OnFind")
		if treeItem == nil {
			return nil, fmt.Errorf("item not found")
		}

		disjointNode := treeItem.(*disjointNode)
		defer disjointNode.lock.RUnlock()

		if onFind != "" {
			_ = reflect.ValueOf(disjointNode.value).MethodByName(onFind).Call(nil)
		}

		return disjointNode.value, nil
	default:
		treeItem := dt.tree.Find(keys[0], "OnFind")
		if treeItem == nil {
			return nil, fmt.Errorf("item not found")
		}

		disjointNode := treeItem.(*disjointNode)
		defer disjointNode.lock.RUnlock()

		return disjointNode.children.Find(keys[1:], onFind)
	}
}

// FindORCreate is a thread safe way to create, find or update items in the DisjointTree
// All Operations are currently guarded by an exclusive lock to save on memory... which might want to change
// in the future
func (dt *disjointTree) FindOrCreate(keys []TreeKey, onFind string, onCreate func() (any, error)) (any, error) {
	if onCreate == nil {
		return nil, fmt.Errorf("Received a nil onCreate callback. Needs to not be nil")
	}

	switch len(keys) {
	case 0:
		return nil, fmt.Errorf("Received an invalid keys length. Needs to be at least 1")
	case 1:
		// create the item or find the item if it does not currently exist
		treeItem, err := dt.tree.FindOrCreate(keys[0], "OnUpdate", dt.newDisjointNodeWithValue(onCreate))
		if err != nil {
			return nil, err
		}

		disjointNode := treeItem.(*disjointNode)
		defer disjointNode.lock.Unlock()

		if disjointNode.value == nil {
			// this is an update for a node that was added that didn't have an original value
			// similar to an OnCreate since the value was not set
			value, err := onCreate()
			if err != nil {
				return nil, err
			}

			disjointNode.value = value
		} else {
			// This is an on find call
			if onFind != "" {
				_ = reflect.ValueOf(disjointNode.value).MethodByName(onFind).Call(nil)
			}
		}

		return disjointNode.value, nil
	default:
		// don't need an on find callback here since we are recursing down
		treeItem, err := dt.tree.FindOrCreate(keys[0], "OnUpdate", dt.newDisjointNode)
		if err != nil {
			return nil, err
		}

		disjointNode := treeItem.(*disjointNode)
		defer disjointNode.lock.Unlock()

		if disjointNode.children == nil {
			disjointNode.children = NewDisjointTree()
		}

		return disjointNode.children.FindOrCreate(keys[1:], onFind, onCreate)
	}
}
