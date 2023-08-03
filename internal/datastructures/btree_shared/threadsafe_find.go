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

	var idSet set.Set[string]
	keyValuePairsLen := len(keyValuePairs)
	sortedKeys := keyValuePairs.SoretedKeys()

	// callback for when a "value" is found
	failFast := false
	counter := 0
	findValue := func(item any) {
		if failFast {
			return
		}
		counter++

		idNode := item.(*threadsafeIDNode)
		idNode.lock.RLock()
		defer idNode.lock.RUnlock()

		if len(idNode.ids) >= keyValuePairsLen {
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuePairsLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuePairsLen-1])
			}
		} else {
			failFast = true
		}
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

	// if there is only 1 value in the set, then we know that we found the desired object
	if failFast == false && counter == keyValuePairsLen && idSet != nil && idSet.Size() == 1 {
		tsat.ids.Find(datatypes.String(idSet.Values()[0]), onFind)
	}

	return nil
}
