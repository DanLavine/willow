package btreeshared

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func (tsat *threadsafeAssociatedTree) Find(keyValuePairs datatypes.StringMap, onFind datastructures.OnFind) error {
	if len(keyValuePairs) == 0 {
		return fmt.Errorf("keyValuePairs cannot be empty")
	}

	if onFind == nil {
		return fmt.Errorf("onFind cannot be nil")
	}

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	var idSet set.Set[uint64]
	keyValuePairsLen := len(keyValuePairs)
	sortedKeys := keyValuePairs.SoretedKeys()

	// callback for when a "value" is found
	failFast := false
	findValue := func(item any) {
		idNode := item.(*threadsafeIDNode)

		// if the ids[index] doesn't exist, then there are no key value pairs for the desired length
		if len(idNode.ids) < keyValuePairsLen {
			failFast = true
			return
		}

		// need to save the possible IDs it could be
		if idSet == nil {
			idSet = set.New[uint64](idNode.ids[keyValuePairsLen-1]...)
		} else {
			idSet.Intersection(idNode.ids[keyValuePairsLen-1])
		}
	}

	// callback for when a "key" is found
	findKey := func(key string) func(item any) {
		return func(item any) {
			valuesNode := item.(*threadsafeValuesNode)
			if err := valuesNode.values.Find(keyValuePairs[key], findValue); err != nil {
				panic(err)
			}

			// break early if we encounter an error
			if failFast {
				return
			}
		}
	}

	// filter all the key value pairs into one specific id to lookup
	for _, key := range sortedKeys {
		tsat.keys.Find(datatypes.String(key), findKey(key))

		// break early if we find a key that doesn't have the desired key value pairs length
		if failFast {
			return nil
		}
	}

	// if there is on'y 1 value in the set, then we know that we found the desired object
	if idSet != nil && idSet.Size() == 1 {
		onFind(tsat.ids.Get(idSet.Values()[0]))
	}

	return nil
}
