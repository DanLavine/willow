package btree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Deletion tutorial: https://www.youtube.com/watch?v=GKa_t7fF8o0
//
// order is the number of children a tree can have
//
// min keyValues = (order-1)/2  (Other than root. that can have 1 key)
// max keyValues = order - 1
// min children = (order + 1)/ 2
// max children = order

// Deletions requires a number of step to move keys around to guarantee balance is keept in the the tree:
//
// 1. Leaf:
//	Deletions are performed recusrsively and each node going up the tree to the root must perform
//  a rebalancing act that follows these rules:
//		a. Clean:
//			There are enough values on the node that a delete >= min values reached.
//		b. Merge
//			Deletion causes values < lest than min AND sibling(s) <= min values
//		c. Rotate
//			Deletion causes values < lest than min AND a sibling(s) > min values.
//
// 2. Internal Node (at index i):
//	a. Swap - ALWAYS swap with a child node and delete from the child.
//		When swapping, we need to find the next greates or smallest value closest to the key being removed.
//		To do this, we use chose from the 2 sub trees such that:
//		1.	Iff the child[i+1] has more keyValues > child[i] keyValues use child[i+1]
//    2. use child[i]
//    Then swap the least value from child[i+1] or greates value from child[i]. Then delete from the leaf node
//		and follow Leaf deletion rules

// Delete a keyValue from the BTree for the given Key. If the key does not exist
// in the tree then this performs a no-op. If the key is nil, then Delete will panic
//
// PARAMS:
// - key - the key for the item to delete from the tree
// - canDelete - optional function to check if an item can be deleted. If this is nil, the item will be deleted from the tree
// RETURNS:
// - error - errors encountered when validating params
func (btree *threadSafeBTree) Delete(key datatypes.EncapsulatedData, canDelete datastructures.CanDelete) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	btree.lock.Lock()
	defer btree.lock.Unlock()

	if btree.root != nil {
		btree.root.remove(key, canDelete)

		// set root to child tree if it exists, or to a nil value
		if btree.root.numberOfValues == 0 {
			btree.root = btree.root.children[0]
		}
	}

	return nil
}

// remove an item from the tree.
func (bn *threadSafeBNode) remove(keyToDelete datatypes.EncapsulatedData, canDelete datastructures.CanDelete) {
	// find the currrent or child index to recurse down
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		// found the index to delete
		if !bn.keyValues[index].key.Less(keyToDelete) {
			// foundthe key in this node
			if !keyToDelete.Less(bn.keyValues[index].key) {
				// return on the leaf. there are no further actions to take other than remove
				if bn.isLeaf() {
					bn.removeLeafItem(index, canDelete)
					return
				}

				// try and delete the value if we can
				if canDelete == nil || canDelete(bn.keyValues[index].value) {
					bn.removeNodeItem(index)
					bn.rebalance(index)
				}

				return
			} else {
				break
			}
		}
	}

	// index to delete is not in this node

	// we are on a leaf, key to remove was not found
	if bn.isLeaf() {
		return
	}

	// recurse and remove the node
	// NOTE: this is here to capture going down the right greater than tree
	bn.children[index].remove(keyToDelete, canDelete)

	// attempt to rebalance the node after a delete of a child
	bn.rebalance(index)
}

// called when removing an item from a leaf node. this is
// the only time any item will be removed from the actual tree
func (bn *threadSafeBNode) removeLeafItem(index int, canDelete datastructures.CanDelete) {
	// check the optional argument for deletion
	if canDelete != nil {
		if !canDelete(bn.keyValues[index].value) {
			return
		}
	}

	// need to shift the rest of the keyValues to the left by 1
	for shiftIndex := index; shiftIndex < bn.numberOfValues-1; shiftIndex++ {
		bn.keyValues[shiftIndex] = bn.keyValues[shiftIndex+1]
	}

	// always remove the last index since everything was shifted
	bn.keyValues[bn.numberOfValues-1] = nil
	bn.numberOfValues--
}

// called when the item to remove is on an internal node. In this case, we need to
// swap the item with a leaf node and delete from there.
//
// NOTE: we need to swap on the left side when both keyValues are at the minimum number of keyValues
func (bn *threadSafeBNode) removeNodeItem(index int) {
	if bn.children[index+1].numberOfValues > bn.children[index].numberOfValues {
		// swap the smallest element on right sub tree
		bn.children[index+1].swap(bn, index)
	} else {
		// swap the larget element on left sub tree
		bn.children[index].swap(bn, index)
	}
}

// swap is used to recursively find the proper leaf node and swap the inner
// node's keyValue to be removed
func (bn *threadSafeBNode) swap(swapNode *threadSafeBNode, swapIndex int) {
	// on a leaf node to swap and just return.
	if bn.isLeaf() {
		if !bn.keyValues[0].key.Less(swapNode.keyValues[swapIndex].key) {
			// swapping on the left most keyValue and deleting
			swapNode.keyValues[swapIndex] = bn.keyValues[0]
			bn.removeLeafItem(0, nil) // already checked if item can be deleted
		} else {
			// swapping on the right most keyValue and deleting
			swapNode.keyValues[swapIndex] = bn.keyValues[bn.numberOfValues-1]
			bn.removeLeafItem(bn.numberOfValues-1, nil) // already checked if item can be deleted
		}

		return
	}

	// recursivly swap
	if !bn.keyValues[0].key.Less(swapNode.keyValues[swapIndex].key) {
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
func (bn *threadSafeBNode) rebalance(childIndex int) {
	// no action to take if the child is still saturated
	if bn.children[childIndex].numberOfValues >= bn.minValues() {
		return
	}

	switch childIndex {
	case 0: // can only look to the right children
		if bn.children[childIndex+1].numberOfValues > bn.minValues() {
			// rotate a keyValue from right to left
			bn.rotateLeft(childIndex)
			return
		}
	case bn.numberOfChildren - 1: // can only look to the left children
		if bn.children[childIndex-1].numberOfValues > bn.minValues() {
			// rotate a keyValue from left to right
			bn.rotateRight(childIndex)
			return
		}
	default: // can look to both left and right children
		if bn.children[childIndex+1].numberOfValues > bn.minValues() {
			// rotate a keyValue from right to left
			bn.rotateLeft(childIndex)
			return
		} else if bn.children[childIndex-1].numberOfValues > bn.minValues() {
			// rotate a keyValue from left to right
			bn.rotateRight(childIndex)
			return
		}
	}

	// merge nodes down
	bn.mergeDown(childIndex)
}

// Rotate left can be used to rotate a keyValue from a right child tree into a left child tree
//
// Example tree (order 3): each leaf or internal node can have 1 keyValue, but no fewer
//
//	        6
//	    /         \
//	   3         9 ,  12
//	 /   \     /    |    \
//	1,2  4,5  7,8  10,11  13,14
//
// PARAMS:
// * rotateChildIndex - left index that will need a keyValue to be rotated into it
func (bn *threadSafeBNode) rotateLeft(rotateChildIndex int) {
	child := bn.children[rotateChildIndex]

	// perform rotation
	child.keyValues[child.numberOfValues] = bn.keyValues[rotateChildIndex] // copy current keyValue into left child tree
	child.numberOfValues++

	child.children[child.numberOfChildren] = bn.children[rotateChildIndex+1].children[0] // copy right children to left children if they are there
	bn.keyValues[rotateChildIndex] = bn.children[rotateChildIndex+1].keyValues[0]        // move right child to new nodes keyValues

	if child.numberOfChildren != 0 {
		child.numberOfChildren++
	}

	// shift right child left 1, removing the 0 indexes now they have been rotated to the left child and new keyValues
	bn.children[rotateChildIndex+1].shiftNodeLeft(0, 1)
}

// Example tree (order 3): each leaf or internal node can have 1 keyValue, but no fewer
//
//	        6
//	    /         \
//	   3         9 ,  12
//	 /   \     /    |    \
//	1,2  4,5  7,8  10,11  13,14
//
// rotateRight is used to move any undeer flow keyValues from a left child to a right child.
// For example, if 4,5 was deleted, then 3 would move down to that location and 2 would move up to 3.
// Likewise, removing 3,4,5, would create an empty node on the left with
//
// PARAMS:
// * rotateChildIndex - right index that will need a keyValue to be rotated into it
func (bn *threadSafeBNode) rotateRight(rotateChildIndex int) {
	child := bn.children[rotateChildIndex]

	child.shiftNodeRight(0, 1) // shift all right child elements right 1

	child.keyValues[0] = bn.keyValues[rotateChildIndex-1]                          // copy current keyValue into the right child
	child.children[0] = bn.children[rotateChildIndex-1].lastChild()                // copy left's greatest keyValue children to the right
	bn.keyValues[rotateChildIndex-1] = bn.children[rotateChildIndex-1].lastValue() // copy the left child's greates keyValue into current keyValue

	// drop the greatest keyValue and children of the left child
	bn.children[rotateChildIndex-1].dropGreatest()
}

// merge down will always merge into the left child index
func (bn *threadSafeBNode) mergeDown(childIndex int) {
	// deteremine the node index we need to merge on
	nodeIndex := childIndex
	if childIndex != 0 {
		// merge the right child into left
		nodeIndex = childIndex - 1
	}
	child := bn.children[nodeIndex]

	child.keyValues[child.numberOfValues] = bn.keyValues[nodeIndex] // move index down
	child.numberOfValues++
	child.appendChildNode(bn.children[nodeIndex+1]) // merge the right node into the left node

	// clear out the merged items
	bn.dropIndexByShiftLeft(nodeIndex, nodeIndex+1)
}

// shifit an entire node from the start index to the right by count.
// this includes moving all the children as well
//
// shiftNodeRight(0,2)
// [1,2,3,4,5,6,7, nil, nil] -> [nil,nil,1,2,3,4,5,6,7]
func (bn *threadSafeBNode) shiftNodeRight(startIndex, count int) {
	for index := bn.numberOfValues - 1; index >= startIndex; index-- {
		bn.keyValues[index+count] = bn.keyValues[index]

		if startIndex+count > index {
			bn.keyValues[index] = nil
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
func (bn *threadSafeBNode) shiftNodeLeft(startIndex, count int) {
	for index := startIndex; startIndex < bn.numberOfValues-1; index++ {
		if index+count > bn.maxValues() {
			break
		} else if index+count >= bn.numberOfValues {
			bn.keyValues[index] = nil
			continue
		}

		bn.keyValues[index] = bn.keyValues[index+count]
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
func (bn *threadSafeBNode) dropIndexByShiftLeft(nodeIndex, childIndex int) {
	for index := nodeIndex; index < bn.numberOfValues; index++ {
		bn.keyValues[index] = bn.keyValues[index+1]
	}

	for index := childIndex; index < bn.numberOfChildren; index++ {
		bn.children[index] = bn.children[index+1]
	}

	bn.numberOfValues--
	if bn.numberOfChildren != 0 {
		bn.numberOfChildren--
	}
}

// appendChildNode copies the keyValues from the provided node
func (bn *threadSafeBNode) appendChildNode(node *threadSafeBNode) {
	var index int
	for index = 0; index < node.numberOfValues; index++ {
		bn.keyValues[bn.numberOfValues+index] = node.keyValues[index]
		bn.children[bn.numberOfChildren+index] = node.children[index]
	}

	bn.children[bn.numberOfChildren+index] = node.lastChild()

	bn.numberOfValues += node.numberOfValues
	if bn.lastChild() != nil {
		bn.numberOfChildren += node.numberOfChildren
	}
}
