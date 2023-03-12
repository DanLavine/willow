package datastructures

import (
	"fmt"
	"math"
	"sync"
)

type lockStrategy int

const (
	lockExclusive lockStrategy = iota
	lockRead
)

// BTree is a 2-3-4 tree implementation of a generic BTree.
// See https://www.geeksforgeeks.org/2-3-4-tree/ for details on what a 2-3-4 tree is
type BTree interface {
	// Find the provided tree item if it already exists. Or return the newly inserted tree item
	FindOrCreate(item TreeItem) TreeItem
}

// Any items that satisfy this interface can be stored in a BTree
type TreeItem interface {
	Less(compareItem TreeItem) bool

	// Generic callback functions that can be set on items stored in the BTree
	//OnFind()
	//CanDelete() bool
}

// BRoot contains the root to the btree. As part of the split operation on inserts and
// deletes we might need to reasign a node. Having this root store the location of that
// root node allows the split operation to happen cleanly and keep track of the proper tree
type BRoot struct {
	lock *sync.RWMutex

	order int
	root  *bNode
}

// bNode is the internal node for a BTree. Each node in the tree is a bNode and stores
// values of TreeItem. Any children nodes are a bNode
type bNode struct {
	lock *sync.RWMutex

	values   []TreeItem // the items for this node
	children []*bNode   // children for this node
}

// create a new BTree.
//
// PARAMS:
// * order - how many values to store in each node. order must be at leas 2
//
// RETURNS:
// * BRoot - root of the BTree that is thread safe for any number of goroutines
// * error - an error if the order is not acceptable
func NewBTree(order int) (*BRoot, error) {
	if order <= 1 {
		return nil, fmt.Errorf("order must be greater than 1 for BTree")
	}

	if order >= math.MaxInt-2 {
		return nil, fmt.Errorf("order must be 2 less than %d", math.MaxInt)
	}

	return &BRoot{
		lock:  new(sync.RWMutex),
		order: order,
		root:  nil, // note this is nil because it can be removed on a "delete". So that case always needs to be handled
	}, nil
}

func newBTreeNode(order int) *bNode {
	return &bNode{
		lock:     new(sync.RWMutex),
		values:   make([]TreeItem, 0, order+1), // set len and cap calls to be constant
		children: make([]*bNode, 0, order+2),   // set len and cap calls to be constant
	}
}

// Inserts the item if it does not already exist. Otherwise the original item already in the
// tree is returned. Will panicc if the item passed in is nil. This is also thread safe to call
// multiple times concurrently
//
// PARAMS:
// * item - item to insert into the BTree
//
// RETURNS:
// * TreeItem - item that was originally passed in for insertion, or the original item that matches
func (ttr *BRoot) FindOrCreate(item TreeItem) TreeItem {
	if item == nil {
		panic("item is nil")
	}

	var returnItem TreeItem

	// always attempt a full read find for an item first. This way multiple
	// reads can happen at once, but an Insert or Delete will then lock the tree structure
	// down to the nodes that need an update
	//ttr.lock.RLock()
	//if ttr.root == nil {
	//	ttr.lock.RUnlock()
	//} else {
	//	returnItem := ttr.root.find(item)
	//	ttr.lock.RUnlock()

	//	if returnItem != nil {
	//		return returnItem
	//	}
	//}

	// item was not found. now we need to create it, so need a tree path lock
	ttr.lock.Lock()
	defer ttr.lock.Unlock()

	if ttr.root == nil {
		ttr.root = newBTreeNode(ttr.order)
	}

	returnItem, newRoot := ttr.root.findOrCreate(ttr.order, item)
	if newRoot != nil {
		ttr.root = newRoot
	}

	return returnItem
}

// create a new item in the BTree
//
// PARAMS:
// * item - item to be inserted into the tree
//
// RETURNS:
// * TreeItem - the value inserted or item if it already existed
// * *bNode - a new node if there was a split
func (bn *bNode) findOrCreate(order int, item TreeItem) (TreeItem, *bNode) {
	bn.lock.Lock()
	defer bn.lock.Unlock()

	switch len(bn.children) {
	case 0: // leaf node
		item = bn.insertTreeItem(item)

		if len(bn.values) > order {
			return item, bn.splitLeaf(order)
		}

		return item, nil
	default: // internal node
		var index int
		for _, value := range bn.values {
			if !value.Less(item) {
				// item already exists, return the original value
				if !item.Less(value) {
					return bn.values[index], nil
				}

				// found index
				break
			}

			index++
		}

		//  will be the found index, or last child index (right child)
		item, node := bn.children[index].findOrCreate(order, item)
		if node != nil {
			bn.insertNode(node)
			if len(bn.values) > order {
				return item, bn.splitNode(order)
			}
		}

		return item, nil
	}
}

// insertTreeItem is called only on "leaf" nodes who have space for a new value
func (bn *bNode) insertTreeItem(item TreeItem) TreeItem {
	for index, value := range bn.values {
		if !value.Less(item) {
			// item already exists, return the original value
			if !item.Less(value) {
				return value
			}

			// shift current items all 1 position
			bn.values = append(bn.values[:index+1], bn.values[index:]...)
			// overwrite value
			bn.values[index] = item
			return item
		}
	}

	bn.values = append(bn.values, item)
	return item
}

// splitLeaf is called only on "leaf" nodes and reurns a new node with 1 value and 2 children
//
// PARAMS:
// * item - item to insert
//
// RETURNS:
// * TreeItem - tree item to be inserted or original item if found
// * bNode - new "root" node of the split nodes. Will be nil if original item is found
func (bn *bNode) splitLeaf(order int) *bNode {
	pivotIndex := order / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(order)
	parentNode.insertTreeItem(bn.values[pivotIndex])

	// 2. create left node
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	for i := 0; i < pivotIndex; i++ {
		_ = parentNode.children[0].insertTreeItem(bn.values[i])
	}

	// 3. create right node
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	for i := pivotIndex + 1; i <= order; i++ {
		_ = parentNode.children[1].insertTreeItem(bn.values[i])
	}

	return parentNode
}

// insertNode is called only on "internal" nodes who have space for a promoted node value
func (bn *bNode) insertNode(node *bNode) {
	for index, value := range bn.values {
		if node.values[0].Less(value) {
			// shift current items all 1 position
			bn.values = append(bn.values[:index+1], bn.values[index:]...)
			bn.children = append(bn.children[:index+1], bn.children[index:]...)
			// overwrite value and children
			bn.values[index] = node.values[0]
			bn.children[index] = node.children[0]
			bn.children[index+1] = node.children[1]

			return
		}
	}

	if len(bn.values) == 0 {
		bn.values = append(bn.values, node.values[0])
		bn.children = append(bn.children, node.children...)
	} else {
		bn.values = append(bn.values, node.values[0])
		bn.children = bn.children[:len(bn.children)-1]
		bn.children = append(bn.children, node.children...)
	}
}

// splitNode is called only on "internal" nodes and reurns a new node with 1 value and 2 children
//
// PARAMS:
// * node - additional node that needs to be added to current node causing the split
// * insertIndex - index where the node would be addeded if there was space
//
// RETURNS:
// * TreeItem - tree item to be inserted or original item if found
// * bNode - new "root" node of the split nodes. Will be nil if original item is found
func (bn *bNode) splitNode(order int) *bNode {
	pivotIndex := order / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(order)
	parentNode.insertTreeItem(bn.values[pivotIndex])

	// 2. create left nodes
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	var index int
	for index = 0; index < pivotIndex; index++ {
		_ = parentNode.children[0].insertTreeItem(bn.values[index])
		parentNode.children[0].children = append(parentNode.children[0].children, bn.children[index])
	}
	parentNode.children[0].children = append(parentNode.children[0].children, bn.children[index])

	// 2. create right nodes
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	for index = pivotIndex + 1; index <= order; index++ {
		_ = parentNode.children[1].insertTreeItem(bn.values[index])
		parentNode.children[1].children = append(parentNode.children[1].children, bn.children[index])
	}
	parentNode.children[1].children = append(parentNode.children[1].children, bn.children[index])

	return parentNode
}

// used to print node during tests. quite helpful
func (bn *bNode) print(key string) {
	if key == "" {
		key = "node."
	}

	if len(bn.children) == 0 {
		for i, value := range bn.values {
			fmt.Printf("%svalues[%d]: %v\n", key, i, value)
		}
		return
	}

	for i := 0; i < len(bn.children); i++ {
		bn.children[i].print(fmt.Sprintf("%schild[%d].", key, i))

		if i != len(bn.children)-1 {
			fmt.Printf("%svalues[%d]: %v\n", key, i, bn.values[i])
		}
	}
}
