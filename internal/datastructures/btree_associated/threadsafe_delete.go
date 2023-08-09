package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Delete an item from the associated tree. This is thread safe to call with any other functions
// for this object
//
// PARAMS:
// - keyValuePair - is a map of key value pairs that compose an object to be deleted if found
// - canDalete - optional parameter to check if an item is found, whether or not the item can be deleted
//
// RETURNS:
// - error - any errors encountered with parameters
func (tsat *threadsafeAssociatedTree) Delete(keyValuePairs datatypes.StringMap, canDelete datastructures.CanDelete) error {
	if len(keyValuePairs) == 0 {
		return fmt.Errorf("keyValuePairs cannot be empty")
	}

	var idSet set.Set[string]
	idNodes := []*threadsafeIDNode{}

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	keyValuePairsLen := len(keyValuePairs)
	sortedKeys := keyValuePairs.SoretedKeys()

	// callback for when a "value" is found
	findValue := func(item any) {
		idNode := item.(*threadsafeIDNode)
		idNodes = append(idNodes, idNode)
	}

	// callback for when a "key" is found
	findKey := func(key string) func(item any) {
		return func(item any) {
			valuesNode := item.(*threadsafeValuesNode)
			if err := valuesNode.values.Find(keyValuePairs[key], findValue); err != nil {
				panic(err)
			}
		}
	}

	// filter all the key value pairs into one specific id to lookup
	for _, key := range sortedKeys {
		tsat.keys.Find(datatypes.String(key), findKey(key))
	}

	// we at least hit all the key value pairs
	if len(idNodes) == keyValuePairsLen {
		for index, idNode := range idNodes {
			idNode.lock.Lock()

			// for any of the ID Nodes, if they don't have the desired key value length, we know there isn't an ID to remove
			// NOTE: This is super important to return early here. In a race with create, there won't be any
			// entries untill they are properly created. So this is a guard to bail early if create is in the process
			// of adding entries
			// NOTE: the above isn't correct. if create, [create|delete] happen, can be messed up still
			if len(idNode.ids) < keyValuePairsLen {
				// unlock any nodes that were locked
				for i := 0; i <= index; i++ {
					idNodes[i].lock.Unlock()
				}

				// just break early.
				return nil
			}

			// need to save the possible IDs it could be
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuePairsLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuePairsLen-1])
			}
		}

		// if there is only 1 value in the set, then we know that we found the desired object id
		if idSet != nil && idSet.Size() == 1 {
			deleted := true
			idToDelete := idSet.Values()[0]

			// try and delte the ID's value from the tree
			tsat.ids.Delete(datatypes.String(idToDelete), func(item any) bool {
				if canDelete != nil {
					deleted = canDelete(item)
				}

				return deleted
			})

			// try to cleanup all ID nodes
			if deleted {
				for _, idNode := range idNodes {
					// remove the id from the id node
					for index, value := range idNode.ids[keyValuePairsLen-1] {
						if value == idToDelete {
							// swap with last element
							idNode.ids[keyValuePairsLen-1][index] = idNode.ids[keyValuePairsLen-1][len(idNode.ids[keyValuePairsLen-1])-1]
							// pop last element
							idNode.ids[keyValuePairsLen-1] = idNode.ids[keyValuePairsLen-1][:len(idNode.ids[keyValuePairsLen-1])-1]
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
	deleteValue := func(item any) bool {
		// don't need to lock here. BTree has exclusive lock access already
		idNode := item.(*threadsafeIDNode)
		idNode.lock.Lock()
		defer idNode.lock.Unlock()

		return len(idNode.ids) == 0 && idNode.creating.Load() == 0
	}

	deleteKey := func(key string) func(item any) bool {
		return func(item any) bool {
			valuesForKey := item.(*threadsafeValuesNode)

			// try to remove the value if it was created
			if err := valuesForKey.values.Delete(keyValuePairs[key], deleteValue); err != nil {
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
