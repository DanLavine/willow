package btree

import (
	"fmt"
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var ErrorKeyExists = fmt.Errorf("key already exists")

// Inserts the keyValue into the tree if the key does not already exist. If the Key does exist
// then an error will be returned and 'onCreate()' will not be called
//
// PARAMS:
// - key - key to use when comparing to other possible values
// - onCreate - callback function to create the value if it does not exist. If the create callback was to fail, its up to the callback to perform any cleanup operations and return nil. In this case nothing will be saved to the tree
//
// RETURNS:
// - error - any errors encontered. I.E. key is not valid
func (btree *threadSafeBTree) Create(key datatypes.EncapsulatedData, onCreate datastructures.OnCreate) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}
	if onCreate == nil {
		return fmt.Errorf("onCreate callback is nil, but a keyValue is required")
	}

	// lock the current node. But as we progress down the tree, if we know that we can insert a new value at a specific level, we can unlock
	// the parent node. This way, we don't need an exclusive lock on the whole tree, but can find the smallest workable subset
	btree.lock.Lock()

	once := new(sync.Once)
	unlock := func() { once.Do(func() { btree.lock.Unlock() }) }
	defer func() { unlock() }()

	if btree.root == nil {
		btree.root = newBTreeNode(btree.nodeSize)
	}

	btree.root.lock.Lock()
	newRoot, err := btree.root.create(unlock, key, onCreate)
	if err != nil {
		return err
	}

	if newRoot != nil {
		btree.root = newRoot
	}

	return nil
}

// Create a new value  in the BTree.
// The threading unlock strategy follows a number of rules when determining if it can unlock a Node
// or the parent Node(s) for parallel requests:
//  1. Lock the Node with an exclusive lock since Crete can add a new value
//  2. Increment an "operation" that happening on the node
//  3. Iff there is space to insert a new value on the node, we can consider realeasiing the locks on the parent nodes.
//     3.1 If the number of parallel process that care about the node < number of free value slots, we can release the lock of the parent nodes.
//     Otherwise, the node will be unlocked when a process on the child completes.
//     3.2 Otherwise, wrap the release into a callback down to another node that eventually has free space
//  4. At the end, decrement the operations counter and ensure the lock is released
//
// PARAMS:
// * value - value to be inserted into the tree
// * onCreate - required function to create the value if it does not yet exist in the tree
//
// RETURNS:
// * TreeItem - the keyValue inserted or value if it already existed
// * *threadSafeBNode - a new node if there was a split
func (bn *threadSafeBNode) create(releaseParentLock func(), key datatypes.EncapsulatedData, onCreate datastructures.OnCreate) (*threadSafeBNode, error) {
	// always release the current lock on return
	once := new(sync.Once)
	unlock := func() { once.Do(func() { bn.lock.Unlock() }) }
	defer func() { unlock() }()

	// special case when recursing if we can unlock the parents
	// This happens when an internal node has guranteed space for a new value
	recurseUnlock := func() {
		releaseParentLock()
		unlock()
	}

	// we know we are not splitting the current node, can release the parent locks
	if bn.numberOfValues+1 <= bn.maxValues() {
		releaseParentLock()
	}

	switch bn.numberOfChildren {
	case 0: // leaf node
		if err := bn.createTreeItem(key, onCreate); err != nil {
			return nil, err
		}

		// need to check the number of values after creation because creating the item can fail and no inster operation takes place
		if bn.numberOfValues > bn.maxValues() {
			return bn.splitLeaf(), nil
		}

		return nil, nil
	default: // internal node
		var index int
		for index = 0; index < bn.numberOfValues; index++ {
			keyValue := bn.keyValues[index]

			if !keyValue.key.Less(key) {
				// value already exists, return an error
				if !key.Less(keyValue.key) {
					return nil, ErrorKeyExists
				}

				// found child index
				break
			}
		}

		//  found the index where the child node to recurse for index creation
		bn.children[index].lock.Lock()
		node, err := bn.children[index].create(recurseUnlock, key, onCreate)
		if err != nil {
			return nil, err
		}

		if node != nil {
			bn.insertNode(node)

			if bn.numberOfValues > bn.maxValues() {
				return bn.splitNode(), nil
			}
		}

		return nil, nil
	}
}

// Inserts the keyValue into the tree if the key does not already exist:
// In this case, the keyValue returned from 'onCreate()' will be saved in the tree iff the return keyValue != nil.
//
// If the key already exists:
// the key's associated keyValue will be passed to the 'onFind' callback.
//
// PARAMS:
// - key - key to use when comparing to other possible values
// - onCreate - callback function to create the value if it does not exist. If the create callback was to fail, its up to the callback to perform any cleanup operations and return nil. In this case nothing will be saved to the tree
// - onFind - method to call if the key already exists
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
	if _ = btree.Find(key, findAlreadyCreated); found {
		return nil
	}

	// lock the current node. But as we progress down the tree, if we know that we can insert a new value at a specific level, we can unlock
	// the parent node. This way, we don't need an exclusive lock on the whole tree, but can find the smallest workable subset
	btree.lock.Lock()

	once := new(sync.Once)
	unlock := func() { once.Do(func() { btree.lock.Unlock() }) }
	defer func() { unlock() }()

	if btree.root == nil {
		btree.root = newBTreeNode(btree.nodeSize)
	}

	btree.root.lock.Lock()
	if newRoot := btree.root.createOrFind(unlock, key, onCreate, onFind); newRoot != nil {
		btree.root = newRoot
	}

	return nil
}

// Create a new value or Find an existing value in the BTree.
// The threading unlock strategy follows a number of rules when determining if it can unlock a Node
// or the parent Node(s) for parallel requests:
// TODO: DSL make sure this is still valid after delete
//  1. Lock the Node with an exclusive lock since Crete can add a new value
//  2. Increment an "operation" that happening on the node
//  3. Iff there is space to insert a new value on the node, we can consider realeasiing the locks on the parent nodes.
//     3.1 If the number of parallel process that care about the node < number of free value slots, we can release the lock of the parent nodes.
//     Otherwise, the node will be unlocked when a process on the child completes.
//     3.2 Otherwise, wrap the release into a callback down to another node that eventually has free space
//  4. At the end, decrement the operations counter and ensure the lock is released
//
// PARAMS:
// * value - value to be inserted into the tree
// * onFind - optional callbaack function that will ass the value as the param to datastructures.OnFind
// * onCreate - required function to create the value if it does not yet exist in the tree
//
// RETURNS:
// * TreeItem - the keyValue inserted or value if it already existed
// * *threadSafeBNode - a new node if there was a split
func (bn *threadSafeBNode) createOrFind(releaseParentLock func(), key datatypes.EncapsulatedData, onCreate datastructures.OnCreate, onFind datastructures.OnFind) *threadSafeBNode {
	// always release the current lock on return
	once := new(sync.Once)
	unlock := func() { once.Do(func() { bn.lock.Unlock() }) }
	defer func() { unlock() }()

	// special case when recursing if we can unlock the parents
	// This happens when an internal node has guranteed space for a new value
	recurseUnlock := func() {
		releaseParentLock()
		unlock()
	}

	// we know we are not splitting the current node, can release the parent locks
	if bn.numberOfValues+1 <= bn.maxValues() {
		releaseParentLock()
	}

	switch bn.numberOfChildren {
	case 0: // leaf node
		bn.createOrFindTreeItem(key, onFind, onCreate)

		// need to check the number of values after creation because creating the item can fail and no inster operation takes place
		if bn.numberOfValues > bn.maxValues() {
			return bn.splitLeaf()
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

				// found child index
				break
			}
		}

		//  found the index where the child node to recurse for index creation
		bn.children[index].lock.Lock()
		if node := bn.children[index].createOrFind(recurseUnlock, key, onCreate, onFind); node != nil {
			bn.insertNode(node)

			if bn.numberOfValues > bn.maxValues() {
				return bn.splitNode()
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
func (bn *threadSafeBNode) createTreeItem(key datatypes.EncapsulatedData, onCreate datastructures.OnCreate) error {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if !keyValue.key.Less(key) {
			// value already exists, return an error
			if !key.Less(keyValue.key) {
				return ErrorKeyExists
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

	return nil
}

// createTreeItem is called only on "leaf" nodes who have space for a new keyValue
//
// PARAMS:
// * key - tree key keyValue
// * value - value to be saved and returned on a Find
func (bn *threadSafeBNode) createOrFindTreeItem(key datatypes.EncapsulatedData, onFind datastructures.OnFind, onCreate datastructures.OnCreate) {
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
func (bn *threadSafeBNode) splitLeaf() *threadSafeBNode {
	pivotIndex := bn.maxValues() / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(bn.maxValues())
	parentNode.appendTreeItem(bn.keyValues[pivotIndex].key, bn.keyValues[pivotIndex].value)
	parentNode.numberOfChildren = 2

	// 2. create left node
	parentNode.children[0] = newBTreeNode(bn.maxValues())
	for i := 0; i < pivotIndex; i++ {
		parentNode.children[0].appendTreeItem(bn.keyValues[i].key, bn.keyValues[i].value)
	}

	// 3. create right node
	parentNode.children[1] = newBTreeNode(bn.maxValues())
	for i := pivotIndex + 1; i <= bn.maxValues(); i++ {
		parentNode.children[1].appendTreeItem(bn.keyValues[i].key, bn.keyValues[i].value)
	}

	return parentNode
}

// insertNode is called only on "internal" nodes who have space for a promoted node keyValue
func (bn *threadSafeBNode) insertNode(node *threadSafeBNode) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if node.keyValues[0].key.Less(keyValue.key) {
			// shift current values all 1 position
			for i := bn.numberOfValues; i > index; i-- {
				bn.keyValues[i] = bn.keyValues[i-1]
			}

			for i := bn.numberOfChildren; i > index; i-- {
				// locks are required in case a read operation is taking place. We want to ensure those finish properly
				bn.children[i-1].lock.Lock()
				bn.children[i] = bn.children[i-1]
				bn.children[i].lock.Unlock()
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
func (bn *threadSafeBNode) splitNode() *threadSafeBNode {
	pivotIndex := bn.maxValues() / 2

	// 1. create the new nodes
	parentNode := newBTreeNode(bn.maxValues())
	parentNode.appendTreeItem(bn.keyValues[pivotIndex].key, bn.keyValues[pivotIndex].value)
	parentNode.numberOfChildren = 2

	// 2. create left nodes
	parentNode.children[0] = newBTreeNode(bn.maxValues())
	var index int
	for index = 0; index < pivotIndex; index++ {
		parentNode.children[0].appendTreeItem(bn.keyValues[index].key, bn.keyValues[index].value)
		parentNode.children[0].children[index] = bn.children[index]
		parentNode.children[0].numberOfChildren++
	}
	parentNode.children[0].children[index] = bn.children[index]
	parentNode.children[0].numberOfChildren++

	// 2. create right nodes
	parentNode.children[1] = newBTreeNode(bn.maxValues())
	for index = pivotIndex + 1; index <= bn.maxValues(); index++ {
		parentNode.children[1].appendTreeItem(bn.keyValues[index].key, bn.keyValues[index].value)
		parentNode.children[1].children[index-pivotIndex-1] = bn.children[index]
		parentNode.children[1].numberOfChildren++
	}
	parentNode.children[1].numberOfChildren++
	parentNode.children[1].children[index-pivotIndex-1] = bn.children[index]

	return parentNode
}
