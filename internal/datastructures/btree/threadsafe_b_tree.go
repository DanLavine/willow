package btree

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// As part of the "any" work. I believe that this would need a field to capture the
// "any" operation. Or we have to know to search for the greatest value (max right side)
// This will break the associated logic though...
//
// threadSafeBTree is a shareable thread safe BTree object.
type threadSafeBTree struct {
	// number of items that can be in a node at any given time
	nodeSize int

	// destroying specific key operations
	destroyingKeysLock *sync.RWMutex
	destroyingKeys     []datatypes.EncapsulatedValue

	// destroying whole tree operations
	readWriteWG *sync.WaitGroup // operations to take place before destroy
	destroying  *atomic.Bool    // destroying the whole tree

	// lock for managing the root of the btree
	lock *sync.RWMutex

	// this can be nil (on delete) or created as part of create and points to the constantly updating root value
	root *threadSafeBNode
}

// threadSafeBNode is the internal node for a threadSafeBTree
type threadSafeBNode struct {
	lock *sync.RWMutex

	numberOfValues int         // number of current values set. Cannot use len(values) since nil value counts towards total items
	keyValues      []*keyValue // the items for this node

	numberOfChildren int                // number of current chidlrent set. Cannot use len(children) since nil values count towards total items
	children         []*threadSafeBNode // children for this node
}

// keyValue is the information stored in the threadSafeBTree provided by the end user
type keyValue struct {
	// lookup key for the provided key
	key datatypes.EncapsulatedValue

	// value client saves and performs operations on in the tree
	value any
}

// NewThreadSafe create a new thread safe BTree
//
// PARAMS:
// - nodeSize - how many values to store in each node. must be at least 2
//
// RETURNS:
// - threadSafeBTree - root of the BTree that is thread safe for any number of goroutines
// - error - an error if the nodeSize is not acceptable
func NewThreadSafe(nodeSize int) (*threadSafeBTree, error) {
	if nodeSize <= 1 {
		return nil, fmt.Errorf("nodeSize must be greater than 1 for BTree")
	}

	if nodeSize >= math.MaxInt-2 {
		return nil, fmt.Errorf("nodeSize must be 2 less than %d", math.MaxInt)
	}

	return &threadSafeBTree{
		nodeSize: nodeSize,

		destroyingKeysLock: new(sync.RWMutex),
		destroyingKeys:     []datatypes.EncapsulatedValue{},

		readWriteWG: new(sync.WaitGroup),
		destroying:  new(atomic.Bool),

		lock: new(sync.RWMutex),
		root: nil, // NOTE: this is nil because it can be removed on a "delete". So that case always needs to be handled
	}, nil
}

// Empty checks to see if the threadSafeBTree is empty
//
// RETURNS:
// - bool - true iff there no items in the tree
func (bt *threadSafeBTree) Empty() bool {
	bt.lock.RLock()
	defer bt.lock.RUnlock()

	return bt.root == nil
}

// general helper functions

func (btree *threadSafeBTree) checkDestroying() error {
	if btree.destroying.Load() {
		return ErrorTreeDestroying
	}

	return nil
}

func (btree *threadSafeBTree) checkDestroyingWithKey(key datatypes.EncapsulatedValue) error {
	if err := btree.checkDestroying(); err != nil {
		return err
	}

	btree.destroyingKeysLock.RLock()
	defer btree.destroyingKeysLock.RUnlock()

	for _, value := range btree.destroyingKeys {
		if value == key {
			return ErrorKeyDestroying
		}
	}

	return nil
}

// newBTreeNode creates a new threadSfeBNode (child) object for a btree
func newBTreeNode(nodeSize int) *threadSafeBNode {
	return &threadSafeBNode{
		lock: new(sync.RWMutex),

		// NOTE: both of these sizes are the intended max size + 1. This is a quick and easy
		//       way to check if the node needs to be split without having to create a bunch of temp variables
		keyValues: make([]*keyValue, nodeSize+1, nodeSize+1),        // set len and cap calls to be constant
		children:  make([]*threadSafeBNode, nodeSize+2, nodeSize+2), // set len and cap calls to be constant
	}
}

// lastChild returns the last child in the threadSafeBTree
func (bn *threadSafeBNode) lastChild() *threadSafeBNode {
	switch bn.numberOfChildren {
	case 0:
		return nil
	default:
		return bn.children[bn.numberOfChildren-1]
	}
}

// lastValue returns the last value in the threadSafeBTree
func (bn *threadSafeBNode) lastValue() *keyValue {
	switch bn.numberOfValues {
	case 0:
		return nil
	default:
		return bn.keyValues[bn.numberOfValues-1]
	}
}

// dropGreates is used to remove the rightmost value and children from a node
func (bn *threadSafeBNode) dropGreatest() {
	bn.keyValues[bn.numberOfValues-1] = nil
	bn.numberOfValues--

	if bn.numberOfChildren != 0 {
		bn.children[bn.numberOfChildren-1].lock.Lock()
		defer bn.children[bn.numberOfChildren-1].lock.Unlock()

		bn.children[bn.numberOfChildren-1] = nil
		bn.numberOfChildren--
	}
}

// leafs will never have any children so just use this as the check
func (bn *threadSafeBNode) isLeaf() bool {
	return bn.numberOfChildren == 0
}

// non leaf. non root
func (bn *threadSafeBNode) minChildren() int {
	// same as math.Ceil(order / 2). but ceil only works for floats
	return (cap(bn.keyValues) + 1) / 2
}

// max number of children a threadSafeBNode can hold
func (bn *threadSafeBNode) maxChildren() int {
	return cap(bn.keyValues)
}

// true leaf. non root
// min value a threadSafeBNode can hold
func (bn *threadSafeBNode) minValues() int {
	return bn.minChildren() - 1
}

// max value a threadSafeBNode can hold
func (bn *threadSafeBNode) maxValues() int {
	return cap(bn.keyValues) - 1
}

// printer helper function for tests to see what the tree actually looks like if there is an error
func (bn *threadSafeBNode) print(parentString string) {
	if parentString == "" {
		parentString = "tree"
	}
	passedString := parentString

	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		parentString = passedString
		parentString = fmt.Sprintf("%s[%d]", parentString, index)

		if bn.children[index] != nil {
			bn.children[index].print(fmt.Sprintf("%s.child[%d]", parentString, index))
		}

		if bn.keyValues[index] != nil {
			if index == 0 {
				fmt.Printf("%s key: %v, number of values: %d, number of children %d\n", parentString, bn.keyValues[index].key, bn.numberOfValues, bn.numberOfChildren)
			} else {
				fmt.Printf("%s key: %v\n", parentString, bn.keyValues[index].key)
			}
		}
	}

	if bn.children[index] != nil {
		bn.children[index].print(fmt.Sprintf("%s.child[%d]", parentString, index))
	}
}
