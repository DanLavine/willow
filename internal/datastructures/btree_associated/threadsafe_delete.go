package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Delete an item from the associated tree. This is thread safe to call with any other functions
// for this object. This will block any other create or find operation for the same key values
//
//	PARAMS:
//	- keyValuePair - is a map of key value pairs that compose an object to be deleted if found
//	- canDalete - optional parameter to check if an item is found, whether or not the item can be deleted
//
//	RETURNS:
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. datatypes.KeyValuesErr // error with the keyValues
//	          2. ErrorKeyValuesHasAssociatedID
//	          3. ErrorTreeItemDestroying
//	          4. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) Delete(keyValues datatypes.KeyValues, canDelete BTreeAssociatedRemove) error {
	// parameters check
	if err := keyValues.Validate(); err != nil {
		return err
	}
	if hasAssociatedID(keyValues) {
		return ErrorKeyValuesHasAssociatedID
	}

	// destroying check
	tsat.readWriteWG.Add(1)
	defer tsat.readWriteWG.Add(-1)

	if tsat.destroying.Load() {
		return ErrorTreeDestroying
	}

	var idSet set.Set[string]
	idNodes := []*threadsafeIDNode{}

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	keyValuesLen := len(keyValues)
	sortedKeys := keyValues.SortedKeys()

	// callback for when a "value" is found
	findValue := func(item any) {
		idNode := item.(*threadsafeIDNode)
		idNodes = append(idNodes, idNode)
	}

	// callback for when a "key" is found
	findKey := func(value datatypes.EncapsulatedValue) func(item any) {
		return func(item any) {
			valuesNode := item.(*threadsafeValuesNode)
			if err := valuesNode.values.Find(value, findValue); err != nil {
				panic(err)
			}
		}
	}

	// filter all the key value pairs into one specific id to lookup
	for _, key := range sortedKeys {
		if err := tsat.keys.Find(datatypes.String(key), findKey(keyValues[key])); err != nil {
			panic(err)
		}
	}

	// we hit all the key value pairs so we can obtain an associatedID to delete
	if len(idNodes) == keyValuesLen {
		for index, idNode := range idNodes {
			idNode.lock.Lock()

			// for any of the ID Nodes, if they don't have the desired key value length, we know there isn't an ID to remove
			// NOTE: This is super important to return early here. In a race with create, there won't be any
			// entries untill they are properly created. So this is a guard to bail early if create is in the process
			// of adding entries
			if len(idNode.ids) < keyValuesLen {
				// unlock any nodes that were locked
				for i := 0; i <= index; i++ {
					idNodes[i].lock.Unlock()
				}

				// just break early.
				return nil
			}

			// need to save the possible IDs it could be
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuesLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuesLen-1])
			}
		}

		// if there is only 1 value in the set, then we know that we found the desired object id
		if idSet != nil && idSet.Size() == 1 {
			deleted := true
			idToDelete := idSet.Values()[0]

			btreeDelete := func(_ datatypes.EncapsulatedValue, item any) bool {
				if canDelete != nil {
					deleted = canDelete(item.(AssociatedKeyValues))
				}

				return deleted
			}

			// try and delte the ID's value from the tree
			if err := tsat.associatedIDs.Delete(datatypes.String(idToDelete), btreeDelete); err != nil {
				switch err {
				case btree.ErrorKeyDestroying:
					// this is fine since the key is already being desstroyed.
					return ErrorTreeItemDestroying
				default:
					panic(err)
				}
			}

			// try to cleanup all ID nodes
			if deleted {
				for _, idNode := range idNodes {
					// remove the id from the id node
					for index, value := range idNode.ids[keyValuesLen-1] {
						if value == idToDelete {
							// swap with last element
							idNode.ids[keyValuesLen-1][index] = idNode.ids[keyValuesLen-1][len(idNode.ids[keyValuesLen-1])-1]
							// pop last element
							idNode.ids[keyValuesLen-1] = idNode.ids[keyValuesLen-1][:len(idNode.ids[keyValuesLen-1])-1]
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
				}
			}
		}

		// ensure we unlock the nodes at the end
		for _, idNode := range idNodes {
			idNode.lock.Unlock()
		}
	}

	// remove the KEYs and VALUEs if there are no more saved IDs
	deleteValue := func(_ datatypes.EncapsulatedValue, item any) bool {
		// don't need to lock here. BTree has exclusive lock access already
		idNode := item.(*threadsafeIDNode)
		idNode.lock.Lock()
		defer idNode.lock.Unlock()

		return len(idNode.ids) == 0 && idNode.creating.Load() == 0
	}

	deleteKey := func(key string) func(_ datatypes.EncapsulatedValue, item any) bool {
		return func(_ datatypes.EncapsulatedValue, item any) bool {
			valuesForKey := item.(*threadsafeValuesNode)

			// try to remove the value if it was created
			if err := valuesForKey.values.Delete(keyValues[key], deleteValue); err != nil {
				panic(err)
			}

			return valuesForKey.values.Empty()
		}
	}

	// cleanup any indexes that are now empty
	for _, key := range sortedKeys {
		if err := tsat.keys.Delete(datatypes.String(key), deleteKey(key)); err != nil {
			return err
		}
	}

	return nil
}

// DeleteByAssociatedID removes an item from the associated tree by the associatedID.
// This is thread safe to call with any other functions for this object
//
//	PARAMS:
//	- associatedID - string value for the associatedID to remove
//	- canDalete - optional parameter to check if an item is found, whether or not the item can be deleted
//
//	RETURNS:
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. ErrorAssociatedIDEmpty
//	          2. ErrorTreeItemDestroying
//	          3. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) DeleteByAssociatedID(assocaitedID string, canDelete BTreeAssociatedRemove) error {
	// parameterChecks
	if assocaitedID == "" {
		return ErrorAssociatedIDEmpty
	}

	// destroying check
	tsat.readWriteWG.Add(1)
	defer tsat.readWriteWG.Add(-1)

	if tsat.destroying.Load() {
		return ErrorTreeDestroying
	}

	var keyValues datatypes.KeyValues

	onFind := func(item AssociatedKeyValues) {
		keyValues = item.KeyValues()
	}

	if err := tsat.FindByAssociatedID(assocaitedID, onFind); err != nil {
		return err
	}

	if keyValues != nil {
		return tsat.Delete(keyValues, canDelete)
	}

	return nil
}

// delete KeyValues from te tree, but not the associatedID. This is used as prt of the 'create' operation
// where we created a bunch of tree items, but the create of the item to save failed and we need to clean up
//
// PARAMS:
// - keyValuePair - is a map of key value pairs that compose an object to be deleted if found
//
// RETURNS:
// - error - any errors encountered with parameters
func (tsat *threadsafeAssociatedTree) deleteKeyValues(associatedID string, keyValues datatypes.KeyValues) error {
	if len(keyValues) == 0 {
		return fmt.Errorf("keyValues cannot be empty")
	}

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	keyValuesLen := len(keyValues)
	sortedKeys := keyValues.SortedKeys()

	// remove the KEYs and VALUEs if there are no more saved IDs
	deleteValue := func(_ datatypes.EncapsulatedValue, item any) bool {
		// don't need to lock here. BTree has exclusive lock access already
		idNode := item.(*threadsafeIDNode)
		idNode.lock.Lock()
		defer idNode.lock.Unlock()

		// remove the id from the id node
		for index, value := range idNode.ids[keyValuesLen-1] {
			if value == associatedID {
				// swap with last element
				idNode.ids[keyValuesLen-1][index] = idNode.ids[keyValuesLen-1][len(idNode.ids[keyValuesLen-1])-1]
				// pop last element
				idNode.ids[keyValuesLen-1] = idNode.ids[keyValuesLen-1][:len(idNode.ids[keyValuesLen-1])-1]
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

	deleteKey := func(value datatypes.EncapsulatedValue) func(_ datatypes.EncapsulatedValue, item any) bool {
		return func(_ datatypes.EncapsulatedValue, item any) bool {
			valuesForKey := item.(*threadsafeValuesNode)

			// try to remove the value if it was created
			if err := valuesForKey.values.Delete(value, deleteValue); err != nil {
				panic(err)
			}

			return valuesForKey.values.Empty()
		}
	}

	// cleanup any indexes that are now empty
	for _, key := range sortedKeys {
		if err := tsat.keys.Delete(datatypes.String(key), deleteKey(keyValues[key])); err != nil {
			return err
		}
	}

	return nil
}
