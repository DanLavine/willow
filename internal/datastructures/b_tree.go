package datastructures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"math"
	"reflect"
	"sync"
)

// BTree is a 2-3-4 tree implementation of a generic BTree.
// See https://www.geeksforgeeks.org/2-3-4-tree/ for details on what a 2-3-4 tree is
type BTree interface {
	// Find the item in the Tree and option to run the OnFind() function for a TreeItem. will be nil if it does not exists
	//
	// PARAMS:
	// * key - key to use when comparing to other possible items
	// * onFind - potential method to call when found by name. The methed must be public and take no parameters or return anything
	//
	// RETURNS:
	// * any - the item if found or nil
	Find(key TreeKey, onFind string) any

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
	FindOrCreate(key TreeKey, onFind string, onCreate func() (any, error)) (any, error)

	// Remove an item in the Tree
	// TODO
	//Remove(key TreeKey)
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

	values   []*value // the items for this node
	children []*bNode // children for this node
}

type value struct {
	key  TreeKey
	item any
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
		root:  nil, // NOTE: this is nil because it can be removed on a "delete". So that case always needs to be handled
	}, nil
}

func newBTreeNode(order int) *bNode {
	return &bNode{
		lock:     new(sync.RWMutex),
		values:   make([]*value, 0, order+1), // set len and cap calls to be constant
		children: make([]*bNode, 0, order+2), // set len and cap calls to be constant
	}
}

// Find a tree item with a thread safe read lock
//
// PARAMS:
// * key - key to use when searching for the associated value
// * onFind - function to call when finding the item. Use the empty string to not call anything
func (ttr *BRoot) Find(key TreeKey, onFind string) any {
	if key == nil {
		panic("key is nil")
	}

	ttr.lock.RLock()
	defer ttr.lock.RUnlock()

	if ttr.root == nil {
		return nil
	}

	return ttr.root.find(key, onFind)
}

// find an item from the tree if it exists
func (bn *bNode) find(key TreeKey, onFind string) any {
	for index, value := range bn.values {
		if !value.key.Less(key) {
			if !key.Less(value.key) {
				if onFind != "" {
					_ = reflect.ValueOf(value.item).MethodByName(onFind).Call(nil)
				}
				return value.item
			}

			if len(bn.children) == 0 {
				return nil
			}

			return bn.children[index].find(key, onFind)
		}
	}

	if len(bn.children) == 0 {
		return nil
	}

	return bn.children[len(bn.children)-1].find(key, onFind)
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
func (ttr *BRoot) FindOrCreate(key TreeKey, onFind string, onCreate func() (any, error)) (any, error) {
	if key == nil {
		panic("key is nil")
	}
	if onCreate == nil {
		panic("item is nil")
	}

	// always attempt a full read find for an item first. This way multiple
	// reads can happen at once, but an Insert or Delete will then lock the tree structure
	// down to the nodes that need an update
	item := ttr.Find(key, onFind)
	if item != nil {
		return item, nil
	}

	// item was not found. now we need to create it, so need a tree path lock
	ttr.lock.Lock()
	defer ttr.lock.Unlock()

	if ttr.root == nil {
		ttr.root = newBTreeNode(ttr.order)
	}

	returnItem, err, newRoot := ttr.root.findOrCreate(ttr.order, key, onFind, onCreate)
	if newRoot != nil {
		ttr.root = newRoot
	}

	return returnItem, err
}

// create a new item in the BTree
//
// PARAMS:
// * item - item to be inserted into the tree
//
// RETURNS:
// * TreeItem - the value inserted or item if it already existed
// * *bNode - a new node if there was a split
func (bn *bNode) findOrCreate(order int, key TreeKey, onFind string, onCreate func() (any, error)) (any, error, *bNode) {
	bn.lock.Lock()
	defer bn.lock.Unlock()

	switch len(bn.children) {
	case 0: // leaf node
		item, err := bn.createTreeItem(key, onFind, onCreate)
		if err != nil {
			return nil, err, nil
		}

		// need to split node
		if len(bn.values) > order {
			return item, nil, bn.splitLeaf(order)
		}

		return item, nil, nil
	default: // internal node
		var index int
		for _, value := range bn.values {
			if !value.key.Less(key) {
				// item already exists, return the original value
				if !key.Less(value.key) {
					item := bn.values[index].item
					if onFind != "" {
						_ = reflect.ValueOf(item).MethodByName(onFind).Call(nil)
					}

					return item, nil, nil
				}

				// found index
				break
			}

			index++
		}

		//  will be the found index, or last child index (right child)
		item, err, node := bn.children[index].findOrCreate(order, key, onFind, onCreate)
		if err != nil {
			return nil, err, nil
		}

		if node != nil {
			bn.insertNode(node)
			if len(bn.values) > order {
				return item, nil, bn.splitNode(order)
			}
		}

		return item, nil, nil
	}
}

// createTreeItem is called only on "leaf" nodes who have space for a new value
//
// PARAMS:
// * key - tree key value
// * item - item to be saved and returned on a Find
func (bn *bNode) createTreeItem(key TreeKey, onFind string, onCreate func() (any, error)) (any, error) {
	for index, currentValue := range bn.values {
		if !currentValue.key.Less(key) {
			// item already exists, return the original value
			if !key.Less(currentValue.key) {
				if onFind != "" {
					_ = reflect.ValueOf(currentValue.item).MethodByName(onFind).Call(nil)
				}

				return currentValue.item, nil
			}

			item, err := onCreate()
			if err != nil {
				return nil, err
			}

			// shift current items all 1 position
			bn.values = append(bn.values[:index+1], bn.values[index:]...)
			// overwrite value
			newValue := &value{key: key, item: item}
			bn.values[index] = newValue

			return newValue.item, nil
		}
	}

	item, err := onCreate()
	if err != nil {
		return nil, err
	}

	newValue := &value{key: key, item: item}
	bn.values = append(bn.values, newValue)
	return newValue.item, nil
}

// insertTreeItem is called only on "leaf" nodes when splitting a node
//
// PARAMS:
// * key - tree key value
// * item - item to be saved and returned on a Find
func (bn *bNode) insertTreeItem(key TreeKey, item any) any {
	for index, currentValue := range bn.values {
		if !currentValue.key.Less(key) {
			// shift current items all 1 position
			bn.values = append(bn.values[:index+1], bn.values[index:]...)
			// overwrite value
			bn.values[index] = &value{key: key, item: item}

			return item
		}
	}

	bn.values = append(bn.values, &value{key: key, item: item})
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
	parentNode.insertTreeItem(bn.values[pivotIndex].key, bn.values[pivotIndex].item)

	// 2. create left node
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	for i := 0; i < pivotIndex; i++ {
		_ = parentNode.children[0].insertTreeItem(bn.values[i].key, bn.values[i].item)
	}

	// 3. create right node
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	for i := pivotIndex + 1; i <= order; i++ {
		_ = parentNode.children[1].insertTreeItem(bn.values[i].key, bn.values[i].item)
	}

	return parentNode
}

// insertNode is called only on "internal" nodes who have space for a promoted node value
func (bn *bNode) insertNode(node *bNode) {
	for index, value := range bn.values {
		if node.values[0].key.Less(value.key) {
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
	parentNode.insertTreeItem(bn.values[pivotIndex].key, bn.values[pivotIndex].item)

	// 2. create left nodes
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	var index int
	for index = 0; index < pivotIndex; index++ {
		_ = parentNode.children[0].insertTreeItem(bn.values[index].key, bn.values[index].item)
		parentNode.children[0].children = append(parentNode.children[0].children, bn.children[index])
	}
	parentNode.children[0].children = append(parentNode.children[0].children, bn.children[index])

	// 2. create right nodes
	parentNode.children = append(parentNode.children, newBTreeNode(order))
	for index = pivotIndex + 1; index <= order; index++ {
		_ = parentNode.children[1].insertTreeItem(bn.values[index].key, bn.values[index].item)
		parentNode.children[1].children = append(parentNode.children[1].children, bn.children[index])
	}
	parentNode.children[1].children = append(parentNode.children[1].children, bn.children[index])

	return parentNode
}
