package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// CreateOrFind inserts or finds the value in the assoociation tree. This is thread safe to call with
// any other functions on the same object.
//
// PARAMS:
// - keyValuePair - is a map of key value pairs that compose an object's identity
// - onCreate - is the callback used to create the value if it doesn't already exist in the tree. This must return nil, if creatiing the object failed.
// - onFind - is the callback used when an item is found in the tree. It will recive the object's value saved in the tree (what was originally provided)
//
// RETURNS:
// - error - any errors encountered with the parameters
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

	// always attempt a find first so we only need read locks
	found := false
	wrappedOnFind := func(item any) {
		found = true
		onFind(item)
	}
	if err := tsat.Find(keyValuePairs, wrappedOnFind); err != nil {
		return err
	}
	if found {
		return nil
	}

	// At this point we are 99%+ going to create the values, so our IDNodes need to use a write lock

	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	keyValuePairsLen := len(keyValuePairs)
	sortedKeys := keyValuePairs.SoretedKeys()

	// callback for when a "value" is found
	findValue := func(item any) {
		idNode := item.(*threadsafeIDNode)
		idNode.lock.Lock()
		defer idNode.lock.Unlock()

		idNode.creating.Add(1)
		idNodes = append(idNodes, idNode)
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

	var idSet set.Set[string]
	for _, idNode := range idNodes {
		idNode.lock.Lock()

		if len(idNode.ids) < keyValuePairsLen {
			// need to create the new association
			for i := len(idNode.ids); i < keyValuePairsLen; i++ {
				idNode.ids = append(idNode.ids, []string{})
			}

			// record that we are going to need a new ID
			if idSet == nil {
				idSet = set.New[string]()
			} else {
				idSet.Clear()
			}
		} else {
			// need to save the possible IDs it could be
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuePairsLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuePairsLen-1])
			}
		}
	}

	// must have been a race where 2 requests tried to create the same object
	if idSet != nil && idSet.Size() == 1 {
		tsat.ids.Find(datatypes.String(idSet.Values()[0]), onFind)

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		return nil
	}

	// always save the new IDs so we can unlock the IDNodes
	newID := tsat.idGenerator.ID()
	for _, idNode := range idNodes {
		idNode.ids[keyValuePairsLen-1] = append(idNode.ids[keyValuePairsLen-1], newID)
	}

	if newValue := onCreate(); newValue != nil {
		tsat.ids.CreateOrFind(datatypes.String(newID), func() any { return newValue }, func(item any) {
			panic(fmt.Errorf("found an id that already exists. Globaly unique ID failure: %s", newID))
		})

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}
	} else {
		// failed to crete the new item to save.

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		// call delete for the item to clean everything up
		tsat.Delete(keyValuePairs, nil)
	}

	return nil
}
