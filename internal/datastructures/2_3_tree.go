package datastructures

import (
	"fmt"
	"sync"
)

type BTree interface {
}

// need to be able to pull the key from any items on a splut
type TreeItem interface {
	Less(compareItem TreeItem) bool
}

// root node never changes for the caller of this package and allows us to update
// the root node on any splits that might occur
type TwoThreeRoot struct {
	lock *sync.RWMutex
	root *twoThreeNode
}

type twoThreeNode struct {
	lock  *sync.RWMutex
	order uint

	// each node can have either keys, or values, but never both
	values   []TreeItem
	children []*twoThreeNode
	// number of keys or items currently being used
	count uint
}

// create a new B+Tree
func NewBTree(order uint) (*TwoThreeRoot, error) {
	if order <= 1 {
		return nil, fmt.Errorf("order must be greater than 1")
	}

	return &TwoThreeRoot{
		lock: new(sync.RWMutex),
		root: newBTreeNode(order, true),
	}, nil
}

func newBTreeNode(order uint, leaf bool) *twoThreeNode {
	return &twoThreeNode{
		lock:     new(sync.RWMutex),
		order:    order,
		values:   make([]TreeItem, order),        // should be fine n memory. They are nil
		children: make([]*twoThreeNode, order+1), // should be fine n memory. They are nil
		count:    0,
	}
}

func (ttr *TwoThreeRoot) FindOrCreate(item TreeItem) TreeItem {
	// always attempt a full read find for an item first so we don't need to lock the entire tree
	ttr.lock.RLock()
	returnItem := ttr.root.find(item)
	ttr.lock.RUnlock()

	if returnItem != nil {
		return returnItem
	}

	// item was not found. now we need to create it, so need a tree lock
	ttr.lock.Lock()
	defer ttr.lock.Unlock()

	returnItem, newRoot := ttr.root.findOrCreate(item)
	if newRoot != nil {
		ttr.root = newRoot
	}

	return returnItem
}

// find an item in the 2-3 Tree. This is a thread safe read only operation
func (ttn *twoThreeNode) find(item TreeItem) TreeItem {
	ttn.lock.RLock()
	defer ttn.lock.RUnlock()

	//// we are on the leaf node, try and find the value
	//for i := 0; i < count; i++ {
	//	// when both less than are true, we know we found the item
	//	if !ttn.values[i].Less(item) && item.Less(ttn.values[i]) {
	//		return ttn.values[i]
	//	}
	//}

	//// no item found
	//return nil

	//// we are on an internal node, so try and find the proper leaf
	//for i := 0; i < count; i++ {
	//}
	return nil
}

// create a new item in the 2-3 tree
//
// PARAMS:
// * item - item to be inserted into the 2-3 tree
//
// RETURNS:
// * TreeItem - the value inserted or already at the desired position
// * *twoThreeNode - a new node if there was a split
func (ttn *twoThreeNode) findOrCreate(item TreeItem) (TreeItem, *twoThreeNode) {
	ttn.lock.Lock()
	defer ttn.lock.Unlock()

	// 1. Check to see where the item fits on this node
	for index, value := range ttn.values {
		// a. found an empty value. can possibly insert it here
		if value == nil {
			// i. check the child. on a first pass this will always be empty.
			//    on 2...n passes this is the right check (greater than) check of the previous value
			if ttn.children[index] != nil {
				item, node := ttn.children[index].findOrCreate(item)
				if node != nil {
					// set new node into current node
				}

				return item, nil
			}

			// ii. no chid node, this must mean we are on a leaf node and can insert the value at our current index
			ttn.insertTreeItem(item, uint(index))
			return item, nil
		}

		// b. check to see if the current item is less than the item in the current list
		if item.Less(value) {
			// i. check to see if there is a less than child node that this value needs to be inserted into
			if ttn.children[index] != nil {
				item, node := ttn.children[index].findOrCreate(item)
				if node != nil {
					if ttn.count < ttn.order {
						// set new node into current node
						ttn.insertNode(uint(index), node)
					} else {
						// need to now split this node and propigate
					}
				}

				return item, nil
			}

			// ii. if there is no less than node, see if we can insert on this node
			if ttn.count < ttn.order {
				ttn.insertTreeItem(item, uint(index))
				return item, nil
			}

			// iii. there is no more room on this node. need to split this node
			return ttn.splitNode(item, uint(index))
		}

		// c. check to see if it is the current item
		if !value.Less(item) {
			// i. this is the value we are trying to insert. so return the original item
			return value, nil
		}

		// d. Start again from the next key value. Don't need to check the right node. that is the same
		//    as a left check (less than) on the next value if it exists
	}

	// 2. check to see if the item can fight on the rightmost child
	if ttn.children[ttn.order-1] != nil {

	}

	// 3. Need to split this node and propigate a new node up
	return ttn.splitNode(item, ttn.count)
}

// split node, splits the current node into a parent containing the "middle" item and a left + right node
// with the 2 subsets of all values
func (ttn *twoThreeNode) splitNode(item TreeItem, insertIndex uint) (TreeItem, *twoThreeNode) {
	leftNode := newBTreeNode(ttn.order, true)
	rightNode := newBTreeNode(ttn.order, true)
	parentNode := newBTreeNode(ttn.order, false)

	node := leftNode           // set to left, parent, right nodes based off of place in order
	nodeInsertIndex := uint(0) // reset each time the node changes
	lookupIndex := uint(0)     // iterator for the known ttn.values

	for i := uint(0); i <= ttn.order; i++ {
		// we reached the middle index, need to set parent index
		if i == (ttn.order+1)/2 {
			nodeInsertIndex = uint(0)
			node = parentNode
		}

		// we reached the right child inde
		if i > (ttn.order+1)/2 {
			nodeInsertIndex = uint(0)
			node = rightNode
		}

		// we are at the insert index
		if i == insertIndex {
			node.values[nodeInsertIndex] = item
			node.count++
			nodeInsertIndex++
			continue
		}

		node.values[nodeInsertIndex] = ttn.values[lookupIndex]
		node.count++
		nodeInsertIndex++
		lookupIndex++
	}

	parentNode.children[0] = leftNode
	parentNode.children[1] = rightNode

	return item, parentNode
}

func (ttn *twoThreeNode) insertNode(insertIndex uint, node *twoThreeNode) {
	for end := ttn.order; end > insertIndex+1; end-- {
		ttn.values[end-1] = ttn.values[end-2]   // shift the values by 1
		ttn.children[end] = ttn.children[end-1] // shift the children by 1
	}

	ttn.count++
	ttn.values[insertIndex] = node.values[0]
	ttn.children[insertIndex] = node.children[0]
	ttn.children[insertIndex+1] = node.children[1]
}

// insertTreeItem is called only on "leaf" nodes who have space for a new value
func (ttn *twoThreeNode) insertTreeItem(item TreeItem, insertIndex uint) {
	// shift all items over by one
	for end := ttn.order; end > insertIndex+1; end-- {
		ttn.values[end-1] = ttn.values[end-2]   // shift the values by 1
		ttn.children[end] = ttn.children[end-1] // shift the children by 1
	}

	ttn.count++
	ttn.values[insertIndex] = item
}
