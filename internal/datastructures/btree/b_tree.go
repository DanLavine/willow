package btree

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"math"
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// BTree is a generic 2-3-4 BTree implementation.
// See https://www.geeksforgeeks.org/2-3-4-tree/ for details on what a 2-3-4 tree is
//
// NOTE:
// I think there is lot of improvement to make things faster here since it currently uses a naive
// approach to locking values pessimistically. But it is at least safe and provides all the functional
// foundation that is needed
type BTree interface {
	// Find the item in the Tree and option to run the OnFind() function for a TreeItem. will be nil if it does not exists
	//
	// PARAMS:
	// * key - key to use when comparing to other possible items
	// * onFind - potential method to call when found by name. The methed must be public and take no parameters or return anything
	//
	// RETURNS:
	// * any - the item if found or nil
	Find(key datatypes.CompareType, onFind datastructures.OnFind) any

	// Find the provided tree item if it already exists. Or return the newly inserted tree item
	//
	// PARAMS:
	// * key - key to use when comparing to other possible items
	// * onFind - potential method to call when found by name. The methed must be public and take no parameters or return anything
	// * onCreate - callback function to create the item if it does not exist. This allows for parameter encapsulation
	//
	// RETURNS:
	// * any - the item if found or created
	// * error - any errors reported from onCreate will be returned if they occur. In that case the item will not be saved
	CreateOrFind(key datatypes.CompareType, onFind datastructures.OnFind, onCreate datastructures.OnCreate) (any, error)

	// Iterate over the tree and for each value found invoke the callback with the node's value
	//
	// PARAMS:
	// * callback - Each item in the BTree will run callback
	Iterate(callback datastructures.Iterate)

	// Delete an item in the Tree
	//
	// PARAMS:
	// * key - key to delete
	// * canDelete - optional function to check if a value can be deleted
	Delete(key datatypes.CompareType, canDelete datastructures.CanDelete)

	// Check to see if there are any elements in a Btree
	Empty() bool
}

// bTree is a shareable thread safe bTree object. It contains a number of private fields
// and needs to be created through the `New()` function
type bTree struct {
	lock *sync.RWMutex

	// number of items that can be in a node at any given time
	nodeSize int

	// this can be nil (on delete) or created as part of create and points to the constantly updating root value
	root *bNode
}

// bNode is the internal node for a BTree. Each node in the tree is a bNode
type bNode struct {
	lock *sync.RWMutex

	numberOfValues int      // number of current values set. Cannot use len(values) since nil counts in that list
	values         []*value // the items for this node

	numberOfChildren int      // number of current chidlrent set. Cannot use len(children) since nil counts in that list
	children         []*bNode // children for this node
}

type value struct {
	key  datatypes.CompareType
	item any
}

// New create a new thread safe *bTree
//
// PARAMS:
// * nodeSize - how many values to store in each node. must be at least 2
//
// RETURNS:
// * BTree - root of the BTree that is thread safe for any number of goroutines
// * error - an error if the nodeSize is not acceptable
func New(nodeSize int) (*bTree, error) {
	if nodeSize <= 1 {
		return nil, fmt.Errorf("nodeSize must be greater than 1 for BTree")
	}

	if nodeSize >= math.MaxInt-2 {
		return nil, fmt.Errorf("nodeSize must be 2 less than %d", math.MaxInt)
	}

	return &bTree{
		lock:     new(sync.RWMutex),
		nodeSize: nodeSize,
		root:     nil, // NOTE: this is nil because it can be removed on a "delete". So that case always needs to be handled
	}, nil
}

func newBTreeNode(nodeSize int) *bNode {
	return &bNode{
		lock:     new(sync.RWMutex),
		values:   make([]*value, nodeSize+1, nodeSize+1), // set len and cap calls to be constant
		children: make([]*bNode, nodeSize+2, nodeSize+2), // set len and cap calls to be constant
	}
}

// general helper functions

// check to see if a btree is empty
func (bt *bTree) Empty() bool {
	bt.lock.RLock()
	defer bt.lock.RUnlock()

	return bt.root == nil
}

// check to see if the key is greate than the current value
func (v *value) greater(compareKey datatypes.CompareType) bool {
	return !v.key.Less(compareKey)
}

// lastChild returns the last child in the bTree
func (bn *bNode) lastChild() *bNode {
	switch bn.numberOfChildren {
	case 0:
		return nil
	default:
		return bn.children[bn.numberOfChildren-1]
	}
}

// lastValue returns the last value in the bTree
func (bn *bNode) lastValue() *value {
	switch bn.numberOfValues {
	case 0:
		return nil
	default:
		return bn.values[bn.numberOfValues-1]
	}
}

// dropGreates is used to remove the rightmost value and children from a node
func (bn *bNode) dropGreatest() {
	bn.values[bn.numberOfValues-1] = nil
	bn.numberOfValues--

	if bn.numberOfChildren != 0 {
		bn.children[bn.numberOfChildren-1] = nil
		bn.numberOfChildren--
	}
}

// leafs will never have any children so just use this as the check
func (bn *bNode) isLeaf() bool {
	return bn.numberOfChildren == 0
}

// non leaf. non root
func (bn *bNode) minChildren() int {
	// same as math.Ceil(order / 2). but ceil only works for floats
	return (cap(bn.values) + 1) / 2
}

func (bn *bNode) maxChildren() int {
	return cap(bn.values)
}

// true leaf. non root
func (bn *bNode) minValues() int {
	return bn.minChildren() - 1
}

func (bn *bNode) maxValues() int {
	return cap(bn.values) - 1
}

// complete helper function for tests
func (bn *bNode) print(parentString string) {
	if parentString == "" {
		fmt.Println("tree")
	}
	passedString := parentString

	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		parentString = passedString
		parentString = fmt.Sprintf("%s[%d]", parentString, index)

		if bn.children[index] != nil {
			bn.children[index].print(fmt.Sprintf("%s.child[%d]", parentString, index))
		}

		if bn.values[index] != nil {
			if index == 0 {
				fmt.Printf("%s key: %v, number of values: %d, number of children %d\n", parentString, bn.values[index].key, bn.numberOfValues, bn.numberOfChildren)
			} else {
				fmt.Printf("%s key: %v\n", parentString, bn.values[index].key)
			}
		}
	}

	if bn.children[index] != nil {
		bn.children[index].print(fmt.Sprintf("%s.child[%d]", parentString, index))
	}
}
