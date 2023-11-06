package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Find an item in the assoociation tree. This is thread safe to call with any other functions on the same object.
//
// PARAMS:
// - keyValuePair - is a map of key value pairs that compose an object's identity
// - onFind - is the callback used when an item is found in the tree. It will recive the object's value saved in the tree (what was originally provided)
//
// RETURNS:
// - string - _associated_id if the item is found
// - error - any errors encountered with the parameters
func (tsat *threadsafeAssociatedTree) Find(keyValuePairs KeyValues, onFind datastructures.OnFind) (string, error) {
	if len(keyValuePairs) == 0 {
		return "", fmt.Errorf("keyValuePairs cannot be empty")
	}

	if onFind == nil {
		return "", fmt.Errorf("onFind cannot be nil")
	}

	var idSet set.Set[string]
	keyValuePairsLen := len(keyValuePairs)
	sortedKeys := keyValuePairs.SortedKeys()

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
	findKey := func(value datatypes.EncapsulatedData) func(item any) {
		return func(item any) {
			valuesNode := item.(*threadsafeValuesNode)
			if err := valuesNode.values.Find(value, findValue); err != nil {
				panic(err)
			}
		}
	}

	// filter all the key value pairs into one specific id to lookup
	for _, key := range sortedKeys {
		tsat.keys.Find(key, findKey(keyValuePairs[key]))
		if failFast {
			break
		}
	}

	// if there is only 1 value in the set, then we know that we found the desired object
	if !failFast && counter == keyValuePairsLen && idSet != nil && idSet.Size() == 1 {
		if err := tsat.associatedIDs.Find(datatypes.String(idSet.Values()[0]), onFind); err != nil {
			panic(err)
		}

		return idSet.Values()[0], nil
	}

	return "", nil
}

// Find an item in the assoociation tree by the assocaitedID generated at creation. This is thread safe to call with any other functions on the same object.
//
// PARAMS:
// - associatedID - is the associated id generated at creation time
// - onFind - is the callback used when an item is found in the tree. It will recive the object's value saved in the tree (what was originally provided)
//
// RETURNS:
// - error - any errors encountered with the parameters
func (tsat *threadsafeAssociatedTree) FindByAssociatedID(associatedID string, onFind datastructures.OnFind) error {
	if onFind == nil {
		return fmt.Errorf("onFind cannot be nil")
	}

	return tsat.associatedIDs.Find(datatypes.String(associatedID), onFind)
}
