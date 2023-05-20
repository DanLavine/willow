package btree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Inserts the item if it does not already exist. Otherwise the original item already in the
// tree is returned. Will panicc if the item passed in is nil. This is also thread safe to call
// multiple times concurrently
//
// PARAMS:
// * item - item to insert into the BTree
// * onFind - optional callbaack function that will ass the item as the param to datastructures.OnFind
// * onCreate - required function to create the item if it does not yet exist in the tree
//
// RETURNS:
// * TreeItem - item that was originally passed in for insertion, or the original item that matches
func (btree *bTree) CreateOrFind(key datatypes.CompareType, onFind datastructures.OnFind, onCreate datastructures.OnCreate) (any, error) {
	if key == nil {
		panic("key is nil")
	}
	if onCreate == nil {
		panic("item is nil")
	}

	// always attempt a full read find for an item first. This way multiple
	// reads can happen at once, but an Insert or Delete will then lock the tree structure
	// down to the nodes that need an update
	item := btree.Find(key, onFind)
	if item != nil {
		return item, nil
	}

	// TODO: Improve me
	// item was not found. now we need to create it, so need a tree path lock
	// There has to be a better way of inserting, but not sure atm how to do that properly
	btree.lock.Lock()
	defer btree.lock.Unlock()

	if btree.root == nil {
		btree.root = newBTreeNode(btree.nodeSize)
	}

	returnItem, err, newRoot := btree.root.createOrFind(btree.nodeSize, key, onFind, onCreate)
	if newRoot != nil {
		btree.root = newRoot
	}

	return returnItem, err
}

// create a new item in the BTree
//
// PARAMS:
// * item - item to be inserted into the tree
// * onFind - optional callbaack function that will ass the item as the param to datastructures.OnFind
// * onCreate - required function to create the item if it does not yet exist in the tree
//
// RETURNS:
// * TreeItem - the value inserted or item if it already existed
// * *bNode - a new node if there was a split
func (bn *bNode) createOrFind(nodeSize int, key datatypes.CompareType, onFind datastructures.OnFind, onCreate datastructures.OnCreate) (any, error, *bNode) {
	switch bn.numberOfChildren {
	case 0: // leaf node
		item, err := bn.createTreeItem(key, onFind, onCreate)
		if err != nil {
			return nil, err, nil
		}

		// need to split node
		if bn.numberOfValues > nodeSize {
			return item, nil, bn.splitLeaf(nodeSize)
		}

		return item, nil, nil
	default: // internal node
		var index int
		for index = 0; index < bn.numberOfValues; index++ {
			value := bn.values[index]

			if !value.key.Less(key) {
				// item already exists, return the original value
				if !key.Less(value.key) {
					item := bn.values[index].item
					if onFind != nil {
						onFind(item)
					}

					return item, nil, nil
				}

				// found index
				break
			}
		}

		//  will be the found or created on the last child index (right child)
		item, err, node := bn.children[index].createOrFind(nodeSize, key, onFind, onCreate)
		if err != nil {
			return nil, err, nil
		}

		if node != nil {
			bn.insertNode(node)
			if bn.numberOfValues > nodeSize {
				return item, nil, bn.splitNode(nodeSize)
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
func (bn *bNode) createTreeItem(key datatypes.CompareType, onFind datastructures.OnFind, onCreate datastructures.OnCreate) (any, error) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		currentValue := bn.values[index]

		if !currentValue.key.Less(key) {
			// item already exists, return the original value
			if !key.Less(currentValue.key) {
				if onFind != nil {
					onFind(currentValue.item)
				}

				return currentValue.item, nil
			}

			// shift current items all 1 position
			for i := bn.numberOfValues; i > index; i-- {
				bn.values[i] = bn.values[i-1]
			}

			break
		}
	}

	item, err := onCreate()
	if err != nil {
		return nil, err
	}

	bn.numberOfValues++
	bn.values[index] = &value{key: key, item: item}
	return item, nil
}

// insertTreeItem is called only on "leaf" nodes when splitting a node
//
// PARAMS:
// * key - tree key value
// * item - item to be saved and returned on a Find
func (bn *bNode) insertTreeItem(key datatypes.CompareType, item any) any {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		currentValue := bn.values[index]

		if !currentValue.key.Less(key) {
			// shift current items all 1 position
			for i := bn.numberOfValues + 1; i > index; i-- {
				bn.values[i] = bn.values[i-1]
			}

			// overwrite value
			bn.values[index] = &value{key: key, item: item}
			bn.numberOfValues++

			return item
		}
	}

	bn.values[index] = &value{key: key, item: item}
	bn.numberOfValues++
	return item
}

// copyTreeItem is called when rebalancing nodes
//
// PARAMS:
// * treeItem - tree item to copy
func (bn *bNode) copyTreeItem(treeItem *value) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		currentValue := bn.values[index]

		if !currentValue.key.Less(treeItem.key) {
			// shift current items all 1 position
			for i := bn.numberOfValues + 1; i > index; i-- {
				bn.values[i] = bn.values[i-1]
			}

			break
		}
	}

	// save value
	bn.values[index] = treeItem
	bn.numberOfValues++
}

// splitLeaf is called only on "leaf" nodes and reurns a new node with 1 value and 2 children
//
// PARAMS:
// * item - item to insert
//
// RETURNS:
// * TreeItem - tree item to be inserted or original item if found
// * bNode - new "root" node of the split nodes. Will be nil if original item is found
func (bn *bNode) splitLeaf(nodeSize int) *bNode {
	pivotIndex := nodeSize / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(nodeSize)
	parentNode.insertTreeItem(bn.values[pivotIndex].key, bn.values[pivotIndex].item)
	parentNode.numberOfChildren = 2

	// 2. create left node
	parentNode.children[0] = newBTreeNode(nodeSize)
	for i := 0; i < pivotIndex; i++ {
		_ = parentNode.children[0].insertTreeItem(bn.values[i].key, bn.values[i].item)
	}

	// 3. create right node
	parentNode.children[1] = newBTreeNode(nodeSize)
	for i := pivotIndex + 1; i <= nodeSize; i++ {
		_ = parentNode.children[1].insertTreeItem(bn.values[i].key, bn.values[i].item)
	}

	return parentNode
}

// insertNode is called only on "internal" nodes who have space for a promoted node value
func (bn *bNode) insertNode(node *bNode) {
	if node.numberOfChildren == 0 && node.numberOfValues == 0 {
		return
	}

	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		value := bn.values[index]

		if node.values[0].key.Less(value.key) {
			// shift current items all 1 position
			for i := bn.numberOfValues; i > index; i-- {
				bn.values[i] = bn.values[i-1]
			}

			for i := bn.numberOfChildren; i > index; i-- {
				bn.children[i] = bn.children[i-1]
			}

			break
		}
	}

	// adding a node to the end
	bn.values[index] = node.values[0]
	bn.children[index] = node.children[0]
	bn.children[index+1] = node.children[1]

	if bn.values[index] != nil {
		bn.numberOfValues++
	}

	if bn.children[index] != nil {
		bn.numberOfChildren++
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
func (bn *bNode) splitNode(nodeSize int) *bNode {
	pivotIndex := nodeSize / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(nodeSize)
	parentNode.insertTreeItem(bn.values[pivotIndex].key, bn.values[pivotIndex].item)
	parentNode.numberOfChildren = 2

	// 2. create left nodes
	parentNode.children[0] = newBTreeNode(nodeSize)
	var index int
	for index = 0; index < pivotIndex; index++ {
		_ = parentNode.children[0].insertTreeItem(bn.values[index].key, bn.values[index].item)
		parentNode.children[0].children[index] = bn.children[index]
		parentNode.children[0].numberOfChildren++
	}
	parentNode.children[0].children[index] = bn.children[index]
	parentNode.children[0].numberOfChildren++

	// 2. create right nodes
	parentNode.children[1] = newBTreeNode(nodeSize)
	for index = pivotIndex + 1; index <= nodeSize; index++ {
		_ = parentNode.children[1].insertTreeItem(bn.values[index].key, bn.values[index].item)
		parentNode.children[1].children[index-pivotIndex-1] = bn.children[index]
		parentNode.children[1].numberOfChildren++
	}
	parentNode.children[1].numberOfChildren++
	parentNode.children[1].children[index-pivotIndex-1] = bn.children[index]

	return parentNode
}
