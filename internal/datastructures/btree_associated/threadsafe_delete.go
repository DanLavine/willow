package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func (at *threadsafeAssociatedTree) Delete(keyValuePairs datatypes.StringMap, canDelete datastructures.CanDelete) error {
	if keyValuePairs == nil {
		return fmt.Errorf("keyValuePairs is nil")
	}

	at.lock.Lock()
	defer at.lock.Unlock()

	var idHolders []*idHolder

	// remove an ID from the ID  Holders
	foundValue := func(item any) {
		idHolder := item.(*idHolder)
		idHolders = append(idHolders, idHolder)
	}

	// recurse through all the values for a particular key
	foundKey := func(findValue datatypes.EncapsulatedData) func(item any) {
		return func(item any) {
			valuesForKey := item.(*keyValues)
			valuesForKey.values.Find(findValue, foundValue)
		}
	}

	// delete the value if the idHolder has no more ids
	deleteValue := func(item any) bool {
		idHolder := item.(*idHolder)
		return len(idHolder.IDs) == 0
	}

	// delete delete the key if there are no more values
	deleteKey := func(findValue datatypes.EncapsulatedData) func(item any) bool {
		return func(item any) bool {
			valuesForKey := item.(*keyValues)
			valuesForKey.values.Delete(findValue, deleteValue)

			return valuesForKey.values.Empty()
		}
	}

	groupedKeyValuesDelete := func(item any) bool {
		groupedKeyValues := item.(*keyValues)

		// first find the ID we need to delete
		for findKey, findValue := range keyValuePairs {
			groupedKeyValues.values.Find(datatypes.String(findKey), foundKey(findValue))
		}

		// exit early since all key value pairs were not found
		if len(idHolders) != len(keyValuePairs) {
			return false
		}

		// filter ids
		idSet := set.New[uint64](idHolders[0].IDs...)
		for i := 1; i < len(idHolders); i++ {
			idSet.Intersection(idHolders[i].IDs)
		}

		// didn't find a shared id
		if idSet.Size() != 1 {
			return false
		}

		// check to see if the item can be deleted
		idToDelete := idSet.Values()[0]
		itemToDelete := at.idTree.Get(idToDelete)
		if canDelete != nil && !canDelete(itemToDelete) {
			return false
		}

		// item can be deleted
		// 1. remove from the item tree
		at.idTree.Remove(idToDelete)
		// 2. remove from all id holders
		for _, idHolder := range idHolders {
			idHolder.remove(idToDelete)
		}
		// 3. loop through the trees again to check their deletion
		for findKey, findValue := range keyValuePairs {
			_ = groupedKeyValues.values.Delete(datatypes.String(findKey), deleteKey(findValue))
		}

		// remove the entire grouped key values if there are no more values
		return groupedKeyValues.values.Empty()
	}

	// attempt to delete
	_ = at.groupedKeyValueAssociation.Delete(datatypes.Int(len(keyValuePairs)), groupedKeyValuesDelete)

	return nil
}
