package btreeassociated

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func (at *associatedTree) Delete(keyValuePairs datatypes.StringMap, canDelete datastructures.CanDelete) {
	if keyValuePairs == nil {
		panic("keyValuePairs is nil")
	}

	at.lock.Lock()
	defer at.lock.Unlock()

	var idHolders []*idHolder
	firstLoop := true
	idSet := set.New()

	// remove an ID from the ID  Holders
	idHolderFind := func(item any) {
		idHolder := item.(*idHolder)
		idHolders = append(idHolders, idHolder)

		if firstLoop {
			idSet.Add(idHolder.IDs)
			firstLoop = false
		} else {
			idSet.Keep(idHolder.IDs)
		}
	}

	// Delete the ID Holder if the number of values stored are 0
	idHolderDelete := func(item any) bool {
		idHolder := item.(*idHolder)
		return len(idHolder.IDs) == 0
	}

	// recurse through all the values for a particular key
	valueFind := func(deleteValue datatypes.String) func(item any) {
		return func(item any) {
			valuesForKey := item.(*keyValues)
			valuesForKey.values.Find(deleteValue, idHolderFind)
		}
	}

	// delete the value for for a particular key
	valueDelete := func(deleteValue datatypes.String) func(item any) bool {
		return func(item any) bool {
			valuesForKey := item.(*keyValues)
			valuesForKey.values.Delete(deleteValue, idHolderDelete)

			return valuesForKey.values.Empty()
		}
	}

	KeyDelete := func(item any) bool {
		valuesForKey := item.(*keyValues)
		return valuesForKey.values.Empty()
	}

	groupedKeyValuesDelete := func(item any) bool {
		groupedKeyValues := item.(*keyValues)
		return groupedKeyValues.values.Empty()
	}

	willDelete := false
	associatedDelete := func(item any) {
		associatedKeyValues := item.(*keyValues)
		associatedKeyValues.lock.Lock()
		defer associatedKeyValues.lock.Unlock()

		for searchKey, searchValue := range keyValuePairs {
			associatedKeyValues.values.Find(searchKey, valueFind(searchValue))
		}

		switch idSet.Len() {
		case 1:
			idToDelete := idSet.Values()[0]
			idItem := at.idTree.Get(idToDelete)

			if canDelete != nil {
				willDelete = canDelete(idItem)
			} else {
				willDelete = true
			}

			if willDelete {
				// remove the item from the ID tree
				_ = at.idTree.Remove(idToDelete)

				// remove the ID from all id holders
				for _, idHolder := range idHolders {
					idHolder.remove(idToDelete)
				}

				// attempt to remove the actual ID holders
				for searchKey, searchValue := range keyValuePairs {
					// attempt to delete the values
					associatedKeyValues.values.Delete(searchKey, valueDelete(searchValue))

					// attempt to delete the keys
					associatedKeyValues.values.Delete(searchKey, KeyDelete)
				}
			}
		default:
			// nothing to do here. Must have not found a key value pair
		}
	}

	// attempt to delete an associated key value group
	_ = at.groupedKeyValueAssociation.Find(datatypes.Int(len(keyValuePairs)), associatedDelete)

	// attempt to delete the whole associated tree
	if willDelete {
		at.groupedKeyValueAssociation.Delete(datatypes.Int(len(keyValuePairs)), groupedKeyValuesDelete)
	}
}
