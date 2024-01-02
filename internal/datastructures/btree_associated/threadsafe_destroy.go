package btreeassociated

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Destroy the item in the tree by the associatedID
//
//	RETURNS:
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. ErrorAssociatedIDEmpty
//	          2. ErrorTreeItemDestroying
//	          3. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) DestroyByAssociatedID(associatedID string, canDelete BTreeAssociatedRemove) error {
	// parameter checks
	if associatedID == "" {
		return ErrorAssociatedIDEmpty
	}

	// check for destroying whole tree in progress
	tsat.readWriteWG.Add(1)
	defer tsat.readWriteWG.Add(-1)

	if tsat.destroying.Load() {
		return ErrorTreeDestroying
	}

	// Don't need to check that the associatedID is being destroyed. That is handled in the bTree code.
	bTreeCanDestroy := func(_ datatypes.EncapsulatedValue, item any) bool {
		associatedKeyValues := item.(AssociatedKeyValues)

		// 1. destroy the item
		if canDelete != nil {
			if !canDelete(associatedKeyValues) {
				return false
			}
		}

		// 2. the item was destroyed, need to remove the KeyValues that define the item
		tsat.destroyKeyValues(associatedID, associatedKeyValues.KeyValues())

		return true
	}

	if err := tsat.associatedIDs.Destroy(datatypes.String(associatedID), bTreeCanDestroy); err != nil {
		// if the key or tree is already being destroed, that is fine, just return the values
		switch err {
		case btree.ErrorKeyDestroying:
			return ErrorTreeItemDestroying
		default:
			panic(err)
		}
	}

	return nil
}

// Destroy all values in the associated tree. Destroy will block wait for all other in flight
// requests to finish processing before starting the ddestroy operation. At which point, any other
// request that comes in will be rejected with a common destroying error.
//
// If the destroy operation fails because an item was not properly deleted, the other operations
// are unlocked
func (tsat *threadsafeAssociatedTree) DestroyTree(canDelete BTreeAssociatedRemove) error {
	// 1. set the destroying operation on the tree to true
	if !tsat.destroying.CompareAndSwap(false, true) {
		return ErrorTreeDestroying
	}
	defer tsat.destroying.Store(false)

	// 2. ensure that no other operations  are processing at the same time
	tsat.readWriteWG.Wait()

	// 3. now perform the deletion
	deleteAssociatedID := func(_ datatypes.EncapsulatedValue, item any) bool {
		associatedKeyValues := item.(AssociatedKeyValues)

		// 1. delete the item saved in associated ids
		if canDelete != nil {
			if !canDelete(associatedKeyValues) {
				// If deleting the item failed, then stop processing
				return false
			}
		}

		// 2. we need to remove the ids from the key values internally
		tsat.deleteKeyValues(associatedKeyValues.AssociatedID(), associatedKeyValues.KeyValues())

		return true
	}

	if err := tsat.associatedIDs.DestroyAll(deleteAssociatedID); err != nil {
		panic(err)
	}

	return nil
}

func (tsat *threadsafeAssociatedTree) destroyKeyValues(associatedID string, keyValues datatypes.KeyValues) {
	deleteValue := func(keyLength int, idToDelete string) func(_ datatypes.EncapsulatedValue, item any) bool {
		return func(_ datatypes.EncapsulatedValue, item any) bool {
			idNode := item.(*threadsafeIDNode)
			idNode.lock.Lock()
			defer idNode.lock.Unlock()

			// remove the id from the id node
			for index, value := range idNode.ids[keyLength-1] {
				if value == idToDelete {
					// swap with last element
					idNode.ids[keyLength-1][index] = idNode.ids[keyLength-1][len(idNode.ids[keyLength-1])-1]
					// pop last element
					idNode.ids[keyLength-1] = idNode.ids[keyLength-1][:len(idNode.ids[keyLength-1])-1]
				}
			}

			// truncate the IDNode to free memory
			for i := len(idNode.ids) - 1; i >= 0; i-- {
				if len(idNode.ids[i]) == 0 {
					idNode.ids = idNode.ids[:len(idNode.ids)-1]
				} else {
					// don't process anymore since we found a value
					break
				}
			}

			return len(idNode.ids) == 0 && idNode.creating.Load() == 0
		}
	}

	deleteKey := func(value datatypes.EncapsulatedValue, len int, idToDelete string) func(_ datatypes.EncapsulatedValue, item any) bool {
		return func(_ datatypes.EncapsulatedValue, item any) bool {
			valuesNode := item.(*threadsafeValuesNode)
			if err := valuesNode.values.Delete(value, deleteValue(len, idToDelete)); err != nil {
				panic(err)
			}

			return valuesNode.values.Empty()
		}
	}

	for key, value := range keyValues {
		if err := tsat.keys.Delete(datatypes.String(key), deleteKey(value, len(keyValues), associatedID)); err != nil {
			panic(err)
		}
	}
}
