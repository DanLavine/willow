package btree

import (
	"fmt"
	"sync"

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
	once := new(sync.Once)
	unlock := func() { once.Do(func() { btree.lock.Unlock() }) }
	defer func() { unlock() }()

	if btree.root != nil {
		btree.root.lock.Lock()
		if btree.root.remove(unlock, btree.nodeSize, key, canDelete) {

			// if we were told to reblance, need to ensure we have the lock for all nodes we are touching
			// this is required when the node size is set to 4+
			if btree.root.numberOfValues == 0 {
				btree.root = btree.root.children[0]
			}
		}
	}

	return nil
}

// remove an item from the tree.
func (bn *threadSafeBNode) remove(releaseParentLock func(), nodeSize int, keyToDelete datatypes.EncapsulatedData, canDelete datastructures.CanDelete) bool {
	// setup unlock operation
	once := new(sync.Once)
	unlock := func() { once.Do(func() { bn.lock.Unlock() }) }
	defer func() { unlock() }()

	// special case when recursing if we can unlock the parents
	// This will happen on nodes that are guranteed to perform a rotation on the keys as they delete
	recurseUnlock := func() {
		releaseParentLock()
		unlock()
	}

	// we know the parent node does not have to take any action on this node's delete operation
	// since we won't cause a rebalance. So we can release all locks above this node
	if bn.numberOfValues-1 >= bn.minValues() {
		releaseParentLock()
	}

	// find the currrent or child index to recurse down
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		// found the index to delete
		if !bn.keyValues[index].key.Less(keyToDelete) {
			if !keyToDelete.Less(bn.keyValues[index].key) {
				// foundthe key in this node

				// return on the leaf. there are no further actions to take other than remove
				if bn.isLeaf() {
					return bn.removeLeafItem(index, canDelete)
				}

				// try and delete the value if we can
				if canDelete == nil || canDelete(bn.keyValues[index].value) {
					if bn.removeNodeItem(recurseUnlock, keyToDelete.Value, index) {
						return bn.rebalance(releaseParentLock, keyToDelete.Value, index)
					}
				}

				// nothing was removed or no need to rebalance
				return false
			} else {
				// need to recurse down to node to delete
				break
			}
		}
	}

	// index to delete is not in this node

	// we are on a leaf, key to remove was not found
	if bn.isLeaf() {
		return false
	}

	// recurse and remove the node
	// NOTE: this is here to capture going down the right greater than tree
	bn.children[index].lock.Lock()
	switch bn.children[index].remove(recurseUnlock, nodeSize, keyToDelete, canDelete) {
	case true:
		// need to rebalance
		return bn.rebalance(releaseParentLock, keyToDelete.Value, index)
	default:
		// no more need to rebalance
		return false
	}
}

// called when removing an item from a leaf node. this is
// the only time any item will be removed from the actual tree
func (bn *threadSafeBNode) removeLeafItem(index int, canDelete datastructures.CanDelete) bool {
	// check the optional argument for deletion
	if canDelete != nil {
		if !canDelete(bn.keyValues[index].value) {
			return false
		}
	}

	// need to shift the rest of the keyValues to the left by 1
	for shiftIndex := index; shiftIndex < bn.numberOfValues-1; shiftIndex++ {
		bn.keyValues[shiftIndex] = bn.keyValues[shiftIndex+1]
	}

	// always remove the last index since everything was shifted
	bn.keyValues[bn.numberOfValues-1] = nil
	bn.numberOfValues--

	return bn.numberOfValues < bn.minValues()
}

// called when the item to remove is on an internal node. In this case, we need to
// swap the item with a leaf node and delete from there.
//
// NOTE: we need to swap on the left side when both keyValues are at the minimum number of keyValues
func (bn *threadSafeBNode) removeNodeItem(releaseLock func(), key any, index int) bool {
	// must lock on the nodes that we check any values against
	bn.children[index].lock.Lock()
	bn.children[index+1].lock.Lock()

	if bn.children[index+1].numberOfValues > bn.children[index].numberOfValues {
		bn.children[index].lock.Unlock()

		// swap the smallest element on right sub tree
		return bn.children[index+1].swap(releaseLock, key, bn, index)
	} else {
		bn.children[index+1].lock.Unlock()

		// swap the larget element on left sub tree
		return bn.children[index].swap(releaseLock, key, bn, index)
	}
}

// swap is used to recursively find the proper leaf node and swap the inner
// node's keyValue to be removed
func (bn *threadSafeBNode) swap(releaseParentLock func(), key any, swapNode *threadSafeBNode, swapIndex int) bool {
	// setup unlock operation
	once := new(sync.Once)
	unlock := func() { once.Do(func() { bn.lock.Unlock() }) }
	defer func() { unlock() }()

	// NOTE: one might think that we could perform a releas lock operation here like so. But we cannot!
	// This is because the swap operation needs to mantain the lock on the original node, that contains
	// the key being removed. Any other operations that may happen if this thread pauses will have bad lookup
	// operations.
	//if bn.numberOfValues-1 >= bn.maxValues() {
	//	releaseParentLock()
	//}

	// on a leaf node to swap and just return.
	if bn.isLeaf() {
		if !bn.keyValues[0].key.Less(swapNode.keyValues[swapIndex].key) {
			// swapping on the left most keyValue and deleting
			swapNode.keyValues[swapIndex] = bn.keyValues[0]
			return bn.removeLeafItem(0, nil) // already checked if item can be deleted
		} else {
			// swapping on the right most keyValue and deleting
			swapNode.keyValues[swapIndex] = bn.keyValues[bn.numberOfValues-1]
			return bn.removeLeafItem(bn.numberOfValues-1, nil) // already checked if item can be deleted
		}
	}

	// setup special unlock for child node that is guranteed to to performa rotate operation. In that case,
	// this node's lock can alos be unlocked because the swap operation will have taken place and all restrited
	// resources are free
	recurseUnlock := func() {
		releaseParentLock()
		unlock()
	}

	// recursivly swap
	childIndex := 0
	if !bn.keyValues[0].key.Less(swapNode.keyValues[swapIndex].key) {
		// swapping on the left node
		childIndex = 0
	} else {
		// swapping on the right node
		childIndex = bn.numberOfChildren - 1
	}

	bn.children[childIndex].lock.Lock()
	switch bn.children[childIndex].swap(recurseUnlock, key, swapNode, swapIndex) {
	case true:
		// perform rebaance
		return bn.rebalance(releaseParentLock, key, childIndex)
	default:
		// nothing needs to be rebalanced, the leaf must have had space or already rebalanced
		return false
	}
}

// rebalance is used to chek if a node needs to be rebalanced after removing an index
func (bn *threadSafeBNode) rebalance(releaseParentLock func(), key any, childIndex int) bool {
	switch childIndex {
	case 0: // can only rotate a right child into the left child that needs an additional value
		bn.children[childIndex].lock.Lock()
		bn.children[childIndex+1].lock.Lock()
		defer bn.children[childIndex].lock.Unlock()
		defer bn.children[childIndex+1].lock.Unlock()

		if bn.children[childIndex+1].numberOfValues > bn.minValues() {
			// can release the parent locks since we are guranteed to not change the number of indexes on this node
			releaseParentLock()

			// rotate a keyValue from right to left
			bn.rotateLeft(childIndex)
			//return bn.numberOfValues < bn.minValues()
			return false
		}
	case bn.numberOfChildren - 1: // can only rotate a left children into the right child that needs an additional value
		bn.children[childIndex-1].lock.Lock()
		bn.children[childIndex].lock.Lock()
		defer bn.children[childIndex-1].lock.Unlock()
		defer bn.children[childIndex].lock.Unlock()

		if bn.children[childIndex-1].numberOfValues > bn.minValues() {
			// can release the parent locks since we are guranteed to not change the number of indexes on this node
			releaseParentLock()

			// rotate a keyValue from left to right
			bn.rotateRight(childIndex)
			//return bn.numberOfValues < bn.minValues()
			return false
		}
	default: // can look to both left and right children into child index that needs an additional value
		bn.children[childIndex-1].lock.Lock()
		bn.children[childIndex].lock.Lock()
		bn.children[childIndex+1].lock.Lock()
		defer bn.children[childIndex-1].lock.Unlock()
		defer bn.children[childIndex].lock.Unlock()
		defer bn.children[childIndex+1].lock.Unlock()

		// rotate
		if bn.children[childIndex+1].numberOfValues > bn.minValues() {
			// can release the parent locks since we are guranteed to not change the number of indexes on this node
			releaseParentLock()

			// rotate a keyValue from right to left
			bn.rotateLeft(childIndex)
			//return bn.numberOfValues < bn.minValues()
			return false
		} else if bn.children[childIndex-1].numberOfValues > bn.minValues() {
			// can release the parent locks since we are guranteed to not change the number of indexes on this node
			releaseParentLock()

			// rotate a keyValue from left to right
			bn.rotateRight(childIndex)
			//return bn.numberOfValues < bn.minValues()
			return false
		}
	}

	// merge nodes down
	bn.mergeDown(childIndex)

	return bn.numberOfValues < bn.minValues()
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
	// already have locks for:
	// bn
	// bn.children[rotateChildIndex]
	// bn.children[rotateChildIndex + 1]

	// already have locks for guranteed children
	child := bn.children[rotateChildIndex]

	// perform rotation
	child.keyValues[child.numberOfValues] = bn.keyValues[rotateChildIndex] // copy current keyValue into left child tree
	child.numberOfValues++

	// rotate children iff they exist
	if bn.children[rotateChildIndex+1].numberOfChildren != 0 {
		// copy right children to left children if they are there
		bn.children[rotateChildIndex+1].children[0].lock.Lock()
		child.children[child.numberOfChildren] = bn.children[rotateChildIndex+1].children[0]
		bn.children[rotateChildIndex+1].children[0].lock.Unlock()
	}

	// move right child to new nodes keyValues
	bn.keyValues[rotateChildIndex] = bn.children[rotateChildIndex+1].keyValues[0]

	// this is needed since we have already delete the value and decremented the child. we are inserting into a nil spot
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
	// already have guranteed locks
	// rotateChildIndex and rotateChildIndex - 1
	child := bn.children[rotateChildIndex]

	// perform roation
	child.shiftNodeRight(0, 1)                            // shift all right child elements right 1
	child.keyValues[0] = bn.keyValues[rotateChildIndex-1] // copy current keyValue into the right child

	// rotate children
	if bn.children[rotateChildIndex-1].numberOfChildren != 0 {
		// copy left child to current node
		child.children[0] = bn.children[rotateChildIndex-1].lastChild() // copy left's greatest keyValue children to the right
	}

	// copy the left child's greates keyValue into current keyValue
	bn.keyValues[rotateChildIndex-1] = bn.children[rotateChildIndex-1].lastValue()

	// drop the greatest keyValue and children of the left child
	bn.children[rotateChildIndex-1].dropGreatest()
}

// merge down will always merge into the left child index
func (bn *threadSafeBNode) mergeDown(childIndex int) {
	// already have a lock for:
	// bn
	// bn.children[childIndex-1] (iff we are not at the 0 index)
	// bn.children[childIndex]
	// bn.children[childIndex+1] (iff we are not at the last index)

	// deteremine the node index we need to merge on
	nodeIndex := childIndex
	if childIndex != 0 {
		// merge the right child into left
		nodeIndex = childIndex - 1
	}

	child := bn.children[nodeIndex]
	rightChild := bn.children[nodeIndex+1]

	// move this node's value down to the child
	child.keyValues[child.numberOfValues] = bn.keyValues[nodeIndex] // move index down to the greatest value
	child.numberOfValues++

	// move the right child's values and children into the left node
	var index int
	for index = 0; index < rightChild.numberOfValues; index++ { // iterate over values
		child.keyValues[child.numberOfValues+index] = rightChild.keyValues[index]

		if rightChild.numberOfChildren != 0 {
			rightChild.children[index].lock.Lock()
			child.children[child.numberOfChildren+index] = rightChild.children[index]
			rightChild.children[index].lock.Unlock()
		}
	}

	if rightChild.numberOfChildren != 0 { // copy last child if it exists
		rightChild.lastChild().lock.Lock()
		child.children[child.numberOfChildren+index] = rightChild.lastChild()
		rightChild.lastChild().lock.Unlock()
	}

	child.numberOfValues += rightChild.numberOfValues
	if child.numberOfChildren != 0 {
		child.numberOfChildren += rightChild.numberOfChildren
	}

	// clear out the merged items bys shifting all values to the left by 1
	for index := nodeIndex; index < bn.numberOfValues; index++ { //values
		bn.keyValues[index] = bn.keyValues[index+1]
	}

	for clearChildIndex := nodeIndex + 1; clearChildIndex < bn.numberOfChildren; clearChildIndex++ { //children
		// when shifting the children, we need to account for the possible merge options. In the case
		// where we can look at the right side, we need to ensure we take the n+2 lock
		switch childIndex {
		case 0:
			if bn.children[clearChildIndex+1] != nil {
				bn.children[clearChildIndex+1].lock.Lock()
				bn.children[clearChildIndex] = bn.children[clearChildIndex+1]
				bn.children[clearChildIndex].lock.Unlock()
			} else {
				bn.children[clearChildIndex] = nil
			}
		default:
			if bn.children[clearChildIndex+1] != nil {
				if clearChildIndex >= nodeIndex+2 {
					bn.children[clearChildIndex+1].lock.Lock()
					bn.children[clearChildIndex] = bn.children[clearChildIndex+1]
					bn.children[clearChildIndex].lock.Unlock()
				} else {
					bn.children[clearChildIndex] = bn.children[clearChildIndex+1]
				}
			} else {
				bn.children[clearChildIndex] = nil
			}
		}

	}

	bn.numberOfValues--
	if bn.numberOfChildren != 0 {
		bn.numberOfChildren--
	}
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
	// already have a lock for:
	// bn

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

		// ensure no operations are running on the node we are removing
		bn.children[index].lock.Lock()
		bn.children[index+count].lock.Lock()
		bn.children[index].lock.Unlock()

		// lock the moved child node and unlock it
		bn.children[index] = bn.children[index+count]
		bn.children[index].lock.Unlock()
	}

	bn.numberOfValues -= count
	if bn.numberOfChildren != 0 {
		bn.numberOfChildren -= count
	}
}
