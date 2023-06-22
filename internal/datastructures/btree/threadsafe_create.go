package btree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Inserts the keyValue into the tree if the key does not already exist:
// In this case, the keyValue returned from 'onCreate()' will be saved in the tree iff the return keyValue != nil.
//
// If the key already exists:
// the key's associated keyValue will be passed to the 'onFind' callback.
//
// PARAMS:
// * value - value to insert into the BTree
// * onCreate - required function to create the value if it does not yet exist in the tree
// * onFind - required callbaack function that will ass the value as the param to datastructures.OnFind
//
// RETURNS:
// * TreeItem - value that was originally passed in for insertion, or the original value that matches
//
// PARAMS:
//   - key - key to use when comparing to other possible values
//   - onCreate - callback function to create the value if it does not exist. If the create callback was to fail, its up
//     to the callback to perform any cleanup operations and return nil. In this case nothing will be saved to the tree
//   - onFind - method to call if the key already exists
//
// RETURNS:
// - error - any errors encontered. I.E. key is not valid
func (btree *threadSafeBTree) CreateOrFind(key datatypes.EncapsulatedData, onCreate datastructures.OnCreate, onFind datastructures.OnFind) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if onCreate == nil {
		return fmt.Errorf("onCreate callback is nil, but a keyValue is required")
	}

	if onFind == nil {
		return fmt.Errorf("onFind callback is nil, but a keyValue is required")
	}

	// always attempt a full read find for an value first. This way multiple
	// reads can happen at once, but an Insert or Delete will then lock the tree structure
	// down to the nodes that need an update
	found := false
	findAlreadyCreated := func(item any) {
		found = true
		onFind(item)
	}
	// won't get an error here since params already checked
	_ = btree.Find(key, findAlreadyCreated)
	if found {
		return nil
	}

	// TODO: Improve me
	// value was not found. now we need to create it, so need a tree path lock
	// There has to be a better way of inserting, but not sure atm how to do that properly
	btree.lock.Lock()
	defer btree.lock.Unlock()

	if btree.root == nil {
		btree.root = newBTreeNode(btree.nodeSize)
	}

	if newRoot := btree.root.createOrFind(btree.nodeSize, key, onFind, onCreate); newRoot != nil {
		btree.root = newRoot
	}

	return nil
}

// create a new value in the BTree
//
// PARAMS:
// * value - value to be inserted into the tree
// * onFind - optional callbaack function that will ass the value as the param to datastructures.OnFind
// * onCreate - required function to create the value if it does not yet exist in the tree
//
// RETURNS:
// * TreeItem - the keyValue inserted or value if it already existed
// * *threadSafeBNode - a new node if there was a split
func (bn *threadSafeBNode) createOrFind(nodeSize int, key datatypes.EncapsulatedData, onFind datastructures.OnFind, onCreate datastructures.OnCreate) *threadSafeBNode {
	switch bn.numberOfChildren {
	case 0: // leaf node
		bn.createTreeItem(key, onFind, onCreate)

		// need to split node
		if bn.numberOfValues > nodeSize {
			return bn.splitLeaf(nodeSize)
		}

		return nil
	default: // internal node
		var index int
		for index = 0; index < bn.numberOfValues; index++ {
			keyValue := bn.keyValues[index]

			if !keyValue.key.Less(key) {
				// value already exists, return the original keyValue
				if !key.Less(keyValue.key) {
					onFind(bn.keyValues[index].value)
					return nil
				}

				// found index
				break
			}
		}

		//  will be the found or created on the last child index (right child)
		if node := bn.children[index].createOrFind(nodeSize, key, onFind, onCreate); node != nil {
			bn.insertNode(node)
			if bn.numberOfValues > nodeSize {
				return bn.splitNode(nodeSize)
			}
		}

		return nil
	}
}

// createTreeItem is called only on "leaf" nodes who have space for a new keyValue
//
// PARAMS:
// * key - tree key keyValue
// * value - value to be saved and returned on a Find
func (bn *threadSafeBNode) createTreeItem(key datatypes.EncapsulatedData, onFind datastructures.OnFind, onCreate datastructures.OnCreate) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if !keyValue.key.Less(key) {
			// value already exists, return the original keyValue
			if !key.Less(keyValue.key) {
				onFind(keyValue.value)
				return
			}

			// shift current values all 1 position
			for i := bn.numberOfValues; i > index; i-- {
				bn.keyValues[i] = bn.keyValues[i-1]
			}

			break
		}
	}

	if value := onCreate(); value != nil {
		bn.numberOfValues++
		bn.keyValues[index] = &keyValue{key: key, value: value}
	}
}

// appendTreeItem is called only on "leaf" nodes when splitting a node
//
// PARAMS:
// * key - tree key keyValue
// * value - value to be saved and returned on a Find
func (bn *threadSafeBNode) appendTreeItem(key datatypes.EncapsulatedData, value any) {
	bn.keyValues[bn.numberOfValues] = &keyValue{key: key, value: value}
	bn.numberOfValues++
}

// splitLeaf is called only on "leaf" nodes and reurns a new node with 1 keyValue and 2 children
//
// PARAMS:
// * value - value to insert
//
// RETURNS:
// * TreeItem - tree value to be inserted or original value if found
// * threadSafeBNode - new "root" node of the split nodes. Will be nil if original value is found
func (bn *threadSafeBNode) splitLeaf(nodeSize int) *threadSafeBNode {
	pivotIndex := nodeSize / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(nodeSize)
	parentNode.appendTreeItem(bn.keyValues[pivotIndex].key, bn.keyValues[pivotIndex].value)
	parentNode.numberOfChildren = 2

	// 2. create left node
	parentNode.children[0] = newBTreeNode(nodeSize)
	for i := 0; i < pivotIndex; i++ {
		parentNode.children[0].appendTreeItem(bn.keyValues[i].key, bn.keyValues[i].value)
	}

	// 3. create right node
	parentNode.children[1] = newBTreeNode(nodeSize)
	for i := pivotIndex + 1; i <= nodeSize; i++ {
		parentNode.children[1].appendTreeItem(bn.keyValues[i].key, bn.keyValues[i].value)
	}

	return parentNode
}

// insertNode is called only on "internal" nodes who have space for a promoted node keyValue
func (bn *threadSafeBNode) insertNode(node *threadSafeBNode) {
	//if node.numberOfChildren == 0 && node.numberOfValues == 0 {
	//	return
	//}

	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if node.keyValues[0].key.Less(keyValue.key) {
			// shift current values all 1 position
			for i := bn.numberOfValues; i > index; i-- {
				bn.keyValues[i] = bn.keyValues[i-1]
			}

			for i := bn.numberOfChildren; i > index; i-- {
				bn.children[i] = bn.children[i-1]
			}

			break
		}
	}

	// adding a node to the end
	bn.keyValues[index] = node.keyValues[0]
	bn.children[index] = node.children[0]
	bn.children[index+1] = node.children[1]

	if bn.keyValues[index] != nil {
		bn.numberOfValues++
	}

	if bn.children[index] != nil {
		bn.numberOfChildren++
	}
}

// splitNode is called only on "internal" nodes and reurns a new node with 1 keyValue and 2 children
//
// PARAMS:
// * node - additional node that needs to be added to current node causing the split
// * insertIndex - index where the node would be addeded if there was space
//
// RETURNS:
// * TreeItem - tree value to be inserted or original value if found
// * threadSafeBNode - new "root" node of the split nodes. Will be nil if original value is found
func (bn *threadSafeBNode) splitNode(nodeSize int) *threadSafeBNode {
	pivotIndex := nodeSize / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(nodeSize)
	parentNode.appendTreeItem(bn.keyValues[pivotIndex].key, bn.keyValues[pivotIndex].value)
	parentNode.numberOfChildren = 2

	// 2. create left nodes
	parentNode.children[0] = newBTreeNode(nodeSize)
	var index int
	for index = 0; index < pivotIndex; index++ {
		parentNode.children[0].appendTreeItem(bn.keyValues[index].key, bn.keyValues[index].value)
		parentNode.children[0].children[index] = bn.children[index]
		parentNode.children[0].numberOfChildren++
	}
	parentNode.children[0].children[index] = bn.children[index]
	parentNode.children[0].numberOfChildren++

	// 2. create right nodes
	parentNode.children[1] = newBTreeNode(nodeSize)
	for index = pivotIndex + 1; index <= nodeSize; index++ {
		parentNode.children[1].appendTreeItem(bn.keyValues[index].key, bn.keyValues[index].value)
		parentNode.children[1].children[index-pivotIndex-1] = bn.children[index]
		parentNode.children[1].numberOfChildren++
	}
	parentNode.children[1].numberOfChildren++
	parentNode.children[1].children[index-pivotIndex-1] = bn.children[index]

	return parentNode
}
