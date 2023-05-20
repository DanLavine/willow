package btree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Deletion tutorial: https://www.youtube.com/watch?v=GKa_t7fF8o0
//
// order is the number of children a tree can have
//
// min values = (order-1)/2  (Other than root. that can have 1 key)
// max values = order - 1
// min children = (order + 1)/ 2
// max children = order

// Delete a value from the BTree for the given Key. If the key does not exist
// in the tree then this performs a no-op. If the key is nil, then Delete will panic
//
// PARAMS:
// * key - the key for the item to delete from the tree
// * canDelete - optional function to check if an item can be deleted. If this is nil, the item will be deleted from the tree
func (btree *bTree) Delete(key datatypes.CompareType, canDelete datastructures.CanDelete) {
	if key == nil {
		panic("key is nil")
	}

	btree.lock.Lock()
	defer btree.lock.Unlock()

	if btree.root != nil {
		btree.root.remove(key, canDelete)

		if btree.root.numberOfValues == 0 {
			btree.root = btree.root.children[0]
		}
	}
}

// remove an item from the tree.
func (bn *bNode) remove(keyToDelete datatypes.CompareType, canDelete datastructures.CanDelete) {
	// find the currrent or child index to recurse down
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		if bn.values[index].greater(keyToDelete) {
			break
		}
	}

	if index < bn.numberOfValues && !keyToDelete.Less(bn.values[index].key) {
		// index to delete is in this node

		if bn.isLeaf() {
			// return on the leaf. there are no further actions to take other than remove
			bn.removeLeafItem(index, canDelete)
			return
		} else {
			if canDelete == nil {
				bn.removeNodeItem(index)
			} else {
				if canDelete(bn.values[index].item) {
					// need to swap and remove the item from this node
					bn.removeNodeItem(index)
				} else {
					// just return early, cannot delete the item
					return
				}
			}
		}
	} else {
		// index to delete is not in this node

		// we are on a leaf, key to remove was not found
		if bn.isLeaf() {
			return
		}

		// recurse and remove the node
		bn.children[index].remove(keyToDelete, canDelete)
	}

	bn.rebalance(index)
}

// called when removing an item from a leaf node. this is
// the only time any item will be removed from the actual tree
func (bn *bNode) removeLeafItem(index int, canDelete datastructures.CanDelete) {
	// check the optional argument for deletion
	if canDelete != nil {
		if !canDelete(bn.values[index].item) {
			return
		}
	}

	// need to shift the rest of the values to the left by 1
	for shiftIndex := index; shiftIndex < bn.numberOfValues-1; shiftIndex++ {
		bn.values[shiftIndex] = bn.values[shiftIndex+1]
	}

	// always remove the last index since everything was shifted
	bn.values[bn.numberOfValues-1] = nil
	bn.numberOfValues--
}

// called when the item to remove is on an internal node. In this case, we need to
// swap the item with a leaf node and delete from there.
//
// NOTE: we need to swap on the left side when both values are at the minimum number of values
func (bn *bNode) removeNodeItem(index int) {
	if bn.children[index+1].numberOfValues > bn.children[index].numberOfValues {
		// swap the smallest element on right sub tree
		bn.children[index+1].swap(bn, index)
	} else {
		// swap the larget element on left sub tree
		bn.children[index].swap(bn, index)
	}
}

// swap is used to recursively find the proper leaf node and swap the inner
// node's value to be removed
func (bn *bNode) swap(swapNode *bNode, swapIndex int) {
	// on a leaf node to swap and just return.
	if bn.isLeaf() {
		if bn.values[0].greater(swapNode.values[swapIndex].key) {
			// swapping on the left most value and deleting
			swapNode.values[swapIndex] = bn.values[0]
			bn.removeLeafItem(0, nil) // already checked if item can be deleted
		} else {
			// swapping on the right most value and deleting
			swapNode.values[swapIndex] = bn.values[bn.numberOfValues-1]
			bn.removeLeafItem(bn.numberOfValues-1, nil) // already checked if item can be deleted
		}

		return
	}

	// recursivly swap
	if bn.values[0].greater(swapNode.values[swapIndex].key) {
		// swapping on the left node
		bn.children[0].swap(swapNode, swapIndex)
		bn.rebalance(0)
	} else {
		// swapping on the right node
		bn.children[bn.numberOfChildren-1].swap(swapNode, swapIndex)
		bn.rebalance(bn.numberOfChildren - 1)
	}
}

// rebalance is used to chek if a node needs to be rebalanced after removing an index
func (bn *bNode) rebalance(childIndex int) {
	// no action to take if the child is still saturated
	if bn.children[childIndex].numberOfValues >= bn.minValues() {
		return
	}

	switch childIndex {
	case 0: // can only look to the right children
		if bn.children[childIndex+1].numberOfValues > bn.minValues() {
			// rotate a value from right to left
			bn.rotateLeft(childIndex)
			return
		}
	case bn.numberOfChildren - 1: // can only look to the left children
		if bn.children[childIndex-1].numberOfValues > bn.minValues() {
			// rotate a value from left to right
			bn.rotateRightNew(childIndex)
			return
		}
	default: // can look to both left and right children
		if bn.children[childIndex+1].numberOfValues > bn.minValues() {
			// rotate a value from right to left
			bn.rotateLeft(childIndex)
			return
		} else if bn.children[childIndex-1].numberOfValues > bn.minValues() {
			// rotate a value from left to right
			bn.rotateRightNew(childIndex)
			return
		}
	}

	// merge nodes down
	bn.mergeDown(childIndex)
}

// Rotate left can be used to rotate a value from a right child tree into a left child tree
//
// Example tree (order 3): each leaf or internal node can have 1 value, but no fewer
//
//	        6
//	    /         \
//	   3         9 ,  12
//	 /   \     /    |    \
//	1,2  4,5  7,8  10,11  13,14
//
// PARAMS:
// * rotateChildIndex - left index that will need a value to be rotated into it
func (bn *bNode) rotateLeft(rotateChildIndex int) {
	child := bn.children[rotateChildIndex]

	// perform rotation
	child.copyTreeItem(bn.values[rotateChildIndex])                                      // copy current value into left child tree
	child.children[child.numberOfChildren] = bn.children[rotateChildIndex+1].children[0] // copy right children to left children if they are there
	bn.values[rotateChildIndex] = bn.children[rotateChildIndex+1].values[0]              // move right child to new nodes values

	if child.numberOfChildren != 0 {
		child.numberOfChildren++
	}

	// shift right child left 1, removing the 0 indexes now they have been rotated to the left child and new values
	bn.children[rotateChildIndex+1].shiftNodeLeft(0, 1)
}

// Example tree (order 3): each leaf or internal node can have 1 value, but no fewer
//
//	        6
//	    /         \
//	   3         9 ,  12
//	 /   \     /    |    \
//	1,2  4,5  7,8  10,11  13,14
//
// rotateRight is used to move any undeer flow values from a left child to a right child.
// For example, if 4,5 was deleted, then 3 would move down to that location and 2 would move up to 3.
// Likewise, removing 3,4,5, would create an empty node on the left with
//
// PARAMS:
// * rotateChildIndex - right index that will need a value to be rotated into it
func (bn *bNode) rotateRightNew(rotateChildIndex int) {
	child := bn.children[rotateChildIndex]

	child.shiftNodeRight(0, 1) // shift all right child elements right 1

	child.values[0] = bn.values[rotateChildIndex-1]                             // copy current value into the right child
	child.children[0] = bn.children[rotateChildIndex-1].lastChild()             // copy left's greatest value children to the right
	bn.values[rotateChildIndex-1] = bn.children[rotateChildIndex-1].lastValue() // copy the left child's greates value into current value

	// drop the greatest value and children of the left child
	bn.children[rotateChildIndex-1].dropGreatest()
}

// merge down will always merge into the left child index
func (bn *bNode) mergeDown(childIndex int) {
	// deteremine the node index we need to merge on
	nodeIndex := childIndex
	if childIndex != 0 {
		// merge the right child into left
		nodeIndex = childIndex - 1
	}

	bn.children[nodeIndex].copyTreeItem(bn.values[nodeIndex])        // move index down
	bn.children[nodeIndex].appendChildNode(bn.children[nodeIndex+1]) // merge the right node into the left node

	// clear out the merged items
	bn.dropIndexByShiftLeft(nodeIndex, nodeIndex+1)
}

// shifit an entire node from the start index to the right by count.
// this includes moving all the children as well
//
// shiftNodeRight(0,2)
// [1,2,3,4,5,6,7, nil, nil] -> [nil,nil,1,2,3,4,5,6,7]
func (bn *bNode) shiftNodeRight(startIndex, count int) {
	for index := bn.numberOfValues - 1; index >= startIndex; index-- {
		bn.values[index+count] = bn.values[index]

		if startIndex+count > index {
			bn.values[index] = nil
		}
	}

	for index := bn.numberOfChildren - 1; index >= startIndex; index-- {
		bn.children[index+count] = bn.children[index]

		if startIndex+count > index {
			bn.children[index] = nil
		}

	}

	bn.numberOfValues += count
	if bn.numberOfChildren != 0 {
		bn.numberOfChildren += count
	}
}

// shifit an entire node from the start index to the left by count.
// this includes moving all the children as well
//
// shiftNodeLeft(2,2)
// [nil,nil,1,2,3,4,5,6,7] -> [1,2,3,4,5,6,7,nil,nil]
func (bn *bNode) shiftNodeLeft(startIndex, count int) {
	for index := startIndex; startIndex < bn.numberOfValues-1; index++ {
		if index+count > bn.maxValues() {
			break
		} else if index+count >= bn.numberOfValues {
			bn.values[index] = nil
			continue
		}

		bn.values[index] = bn.values[index+count]
	}

	for index := startIndex; startIndex < bn.numberOfChildren-1; index++ {
		if index+count > bn.numberOfChildren {
			break
		} else if index+count >= bn.numberOfChildren {
			bn.children[index] = nil
			continue
		}

		bn.children[index] = bn.children[index+count]
	}

	bn.numberOfValues -= count
	if bn.numberOfChildren != 0 {
		bn.numberOfChildren -= count
	}
}

// drop an index by shifting a node to the left
func (bn *bNode) dropIndexByShiftLeft(nodeIndex, childIndex int) {
	for index := nodeIndex; index < bn.numberOfValues; index++ {
		bn.values[index] = bn.values[index+1]
	}

	for index := childIndex; index < bn.numberOfChildren; index++ {
		bn.children[index] = bn.children[index+1]
	}

	bn.numberOfValues--
	if bn.numberOfChildren != 0 {
		bn.numberOfChildren--
	}
}

// appendChildNode copies the values from the provided node
func (bn *bNode) appendChildNode(node *bNode) {
	if node == nil {
		return
	}

	var index int
	for index = 0; index < node.numberOfValues; index++ {
		bn.values[bn.numberOfValues+index] = node.values[index]
		bn.children[bn.numberOfChildren+index] = node.children[index]
	}

	bn.children[bn.numberOfChildren+index] = node.lastChild()

	bn.numberOfValues += node.numberOfValues
	if bn.lastChild() != nil {
		bn.numberOfChildren += node.numberOfChildren
	}
}
