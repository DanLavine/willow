package disjointtree

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/pkg/models/datatypes"
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
	Find(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind) (any, error)

	// Find the provided tree item if it already exists. Or return the newly inserted tree item
	CreateOrFind(keys datatypes.EnumerableCompareType, onFind datastructures.OnFind, onCreate datastructures.OnCreate) (any, error)

	// Iterate over the tree and for each value found invoke the callback with the node's value iff != nil
	Iterate(datastructures.Iterate)
}

// each level of a disjointTree is just a bTree. Then each value in the bTree is a disjointNode
type disjointTree struct {
	//every item in the tree is of type *disjointNode
	tree btree.BTree
}

// each value in a disjointTree's bTree is a disjointNode.
//
// NOTE:
// Each disjointNode might not have a value iff it has children. This happens
// if we create a nested key structure like ["one", "two", "three"]. Only the "three"
// value will have anything assigned where as "one" and "two" will both have children
type disjointNode struct {
	lock     *sync.RWMutex
	value    any
	children *disjointTree
}

func New() *disjointTree {
	tree, err := btree.New(2)
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
func newDisjointNodeEmpty() (any, error) {
	lock := new(sync.RWMutex)
	lock.Lock()

	return &disjointNode{
		lock:     lock,
		value:    nil,
		children: nil,
	}, nil
}

// dtLock is used when finding any disjointNode value in the tree, we need
// to make sure no other processes are using the same tree item
func dtLock(item any) {
	node := item.(*disjointNode)
	node.lock.Lock()
}

// dtReadLock is used when finding any disjointNode value in the tree, we need
// to make sure no other processes are using the same tree item
func dtReadLock(item any) {
	node := item.(*disjointNode)
	node.lock.RLock()
}
