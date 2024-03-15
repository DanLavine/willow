package btree

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Destroy is used to remove a key and set the Tree's configuration so any other calls to the key
// returns an error.
//
//	RETURNS:
//	- error - any errors when trying to destroy he key
//	        - 1. datatypes.EncapsulatedValueErr // error with the key
//	        - 2. datastructures.ErrorKeyDestroying
//	        - 3. datastructures.ErrorTreeDestroying
func (btree *threadSafeBTree) Destroy(key datatypes.EncapsulatedValue, canDelete BTreeRemove) error {
	// parameter checks
	if err := key.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	// check for destroying in progress
	btree.readWriteWG.Add(1)
	defer btree.readWriteWG.Add(-1)

	if btree.destroying.Load() {
		return ErrorTreeDestroying
	}

	// 1. record the key being destroyed
	btree.destroyingKeysLock.Lock()
	for _, destroyingKey := range btree.destroyingKeys {
		if destroyingKey == key {
			btree.destroyingKeysLock.Unlock()
			return ErrorKeyDestroying
		}
	}
	btree.destroyingKeys = append(btree.destroyingKeys, key)
	btree.destroyingKeysLock.Unlock()

	// ensure the destroying key is removed from the end of the operations
	defer func() {
		btree.destroyingKeysLock.Lock()
		for index, destroyingKey := range btree.destroyingKeys {
			if destroyingKey == key {
				btree.destroyingKeys[index] = btree.destroyingKeys[len(btree.destroyingKeys)-1] // swap the index with the last
				btree.destroyingKeys = btree.destroyingKeys[:len(btree.destroyingKeys)-1]       // pop the last index
				break
			}
		}
		btree.destroyingKeysLock.Unlock()
	}()

	// 2. delete the key normally
	return btree.delete(key, canDelete)
}

// DestroyAll values in the BTree. This could be improved slightly, but the fact that we want to
// ensusre the tree is always balanced is kind of a massive headache to do all in one go as each
// deletion can propigate the entire tree.
//   - A delete from an internal node needs to swap with a leaf node
//   - A delete from the leaf node needs to ensure all internal nodes have proper children
//
// So, with that in mind. I think the easiest implementation is to grab an exlusive lock for the root
// of the tree and then Iterate over all values to find every single key. Then Simple call Delete for
// each fund key.
//
// In terms of performance, there is the extra call to find all keys + the memory it consumees, but in the
// grand scheme of things. It is really the massive delete operation this is going to be the main bottle neck
//
//	RETURNS:
//	- error - any errors when trying to destroy he key
//	        - 1. ErrorTreeDestroying
func (btree *threadSafeBTree) DestroyAll(canDelete BTreeRemove) error {
	// set destroying to true. All other operations should check this upfront and exit early if that is the case
	if !btree.destroying.CompareAndSwap(false, true) {
		return ErrorTreeDestroying
	}
	defer btree.destroying.Store(false)

	// wait for all other operations to finish processing
	btree.readWriteWG.Wait()

	// This should be exlusive the whole way through since we need to delete the whole tree.
	// any other in flight create or reads have to wait for this operation to finish
	btree.lock.Lock()
	defer btree.lock.Unlock()

	if btree.root != nil {
		// grab all keys that exist in the tree
		keys := []datatypes.EncapsulatedValue{}
		onIterate := func(key datatypes.EncapsulatedValue) {
			keys = append(keys, key)
		}
		btree.root.iterateDestroy(onIterate)

		// delete everything
		deleted := true
		deleteWrapper := func(key datatypes.EncapsulatedValue, item any) bool {
			if canDelete != nil {
				deleted = canDelete(key, item)
			}
			return deleted
		}

		for _, key := range keys {
			btree.root.lock.Lock() // need this lock as remove always calls unlock for this
			if btree.root.remove(func() {}, btree.nodeSize, key, deleteWrapper) {
				// need to re-set the root node
				if btree.root.numberOfValues == 0 {
					btree.root = btree.root.children[0]
				}
			}

			// one of the items wasn't deleted. need to cancel the delete operation
			if !deleted {
				return nil
			}
		}
	}

	return nil
}

func (bn *threadSafeBNode) iterateDestroy(callback func(key datatypes.EncapsulatedValue)) {
	bn.lock.Lock()
	defer bn.lock.Unlock()

	// add all the values at this level
	for i := 0; i < bn.numberOfValues; i++ {
		callback(bn.keyValues[i].key)
	}

	// iterate through all the children
	for i := 0; i < bn.numberOfChildren; i++ {
		bn.children[i].iterateDestroy(callback)
	}
}
