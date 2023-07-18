package btreeshared

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func (tsat *threadsafeAssociatedTree) CreateOrFind(keyValuePairs datatypes.StringMap, onCreate datastructures.OnCreate, onFind datastructures.OnFind) error {
	if len(keyValuePairs) == 0 {
		return fmt.Errorf("keyValuePairs cannot be empty")
	}

	if onCreate == nil {
		return fmt.Errorf("onCreate cannot be nil")
	}

	if onFind == nil {
		return fmt.Errorf("onFind cannot be nil")
	}

	// TODO: This can use a read lock. but the other part needs a write lock I think. Need to figure that out
	//// always attempt a read first since these are read locked
	//found := false
	//wrappedOnFind := func(item any) {
	//	found = true
	//	onFind(item)
	//}
	//_ = at.Find(keyValuePairs, wrappedOnFind)
	//if found == true {
	//	return nil
	//}

	var idSet set.Set[uint64]
	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	keyValuePairsLen := len(keyValuePairs)
	sortedKeys := keyValuePairs.SoretedKeys()

	// callback for when a "value" is found
	findValue := func(item any) {
		idNode := item.(*threadsafeIDNode)
		idNodes = append(idNodes, idNode)

		// need to create the new association
		if len(idNode.ids) < keyValuePairsLen {
			for i := len(idNode.ids); i < keyValuePairsLen; i++ {
				idNode.ids = append(idNode.ids, []uint64{})
			}

			// record that we are going to need a new ID
			if idSet == nil {
				idSet = set.New[uint64]()
			} else {
				idSet.Clear()
			}

			return
		}

		// need to save the possible IDs it could be
		if idSet == nil {
			idSet = set.New[uint64](idNode.ids[keyValuePairsLen-1]...)
		} else {
			idSet.Intersection(idNode.ids[keyValuePairsLen-1])
		}
	}

	// callback for when a "value" needs to be created
	createValue := func() any {
		newIDNode := newIDNode()
		findValue(newIDNode)

		return newIDNode
	}

	// callback for when a "key" is found
	findKey := func(key string) func(item any) {
		return func(item any) {
			valuesNode := item.(*threadsafeValuesNode)

			lookupValue := keyValuePairs[key]
			if err := valuesNode.values.CreateOrFind(lookupValue, createValue, findValue); err != nil {
				panic(err)
			}
		}
	}

	// callback when creating a new value node when searching for a "key"
	createKey := func(onFind datastructures.OnFind) func() any {
		return func() any {
			newValueNode := newValuesNode()
			onFind(newValueNode)

			return newValueNode
		}
	}

	for _, key := range sortedKeys {
		tsat.keys.CreateOrFind(datatypes.String(key), createKey(findKey(key)), findKey(key))
	}

	// must have been a race where 2 requests tried to create the same object
	if idSet.Size() == 1 {
		onFind(tsat.ids.Get(idSet.Values()[0]))
		return nil
	}

	// need to create the the new object and save the ID
	if newValue := onCreate(); newValue != nil {
		newID := tsat.ids.Add(newValue)
		for _, idNode := range idNodes {
			idNode.ids[keyValuePairsLen-1] = append(idNode.ids[keyValuePairsLen-1], newID)
		}
	} else {
		// something failed. Will need to clean up

		// shrink the value nodes and possibly remove them
		deleteValue := func(item any) bool {
			idNode := item.(*threadsafeIDNode)
			for i := len(idNode.ids) - 1; i >= 0; i-- {
				if len(idNode.ids[i]) == 0 {
					idNode.ids = idNode.ids[:len(idNode.ids)-1]
				} else {
					// don't process anymore since we found a value
					break
				}
			}

			return len(idNode.ids) == 0
		}

		// remove the key if there are no more values
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

		for _, key := range sortedKeys {
			if err := tsat.keys.Delete(datatypes.String(key), deleteKey(key)); err != nil {
				panic(err)
			}
		}
	}

	return nil
}
