package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

func (at *threadsafeAssociatedTree) Find(keyValuePairs datatypes.StringMap, onFind datastructures.OnFind) error {
	if len(keyValuePairs) == 0 {
		return fmt.Errorf("keyValuePairs requires a length of at least 1")
	}

	if onFind == nil {
		return fmt.Errorf("onFind cannot be nil")
	}

	at.lock.RLock()
	defer at.lock.RUnlock()

	counter := 0
	firstLoop := true
	idSet := set.New[uint64]()

	// operation to perform when a key value pair is found
	findValue := func(item any) {
		idHolder := item.(*idHolder)
		counter++

		if firstLoop {
			for _, id := range idHolder.IDs {
				idSet.Add(id)
			}
			firstLoop = false
		} else {
			idSet.Intersection(idHolder.IDs)
		}
	}

	// operation to perform on each key
	findKeys := func(searchValue datatypes.EncapsulatedData) func(item any) {
		return func(item any) {
			valuesForKey := item.(*keyValues)
			valuesForKey.values.Find(searchValue, findValue)
		}
	}

	// operation to perform on the grouped association
	findGroupedKeyValueAssociations := func(item any) {
		associatedKeyValues := item.(*keyValues)

		// loop through all the key value pairs to know if they were processed
		for searchKey, searchValue := range keyValuePairs {
			associatedKeyValues.values.Find(datatypes.String(searchKey), findKeys(searchValue))
		}
	}

	// find the value if they keys exist
	at.groupedKeyValueAssociation.Find(datatypes.Int(len(keyValuePairs)), findGroupedKeyValueAssociations)

	// iff all keys were found, we know that we have an ID of only 1 element
	if counter == len(keyValuePairs) && idSet.Size() == 1 {
		onFind(at.idTree.Get(idSet.Values()[0]))
	}

	return nil
}

// FindQuery will find any items that match the provided query parameter
func (at *threadsafeAssociatedTree) FindQuery(selection query.Select, onFindSelection datastructures.OnFindSelection) error {
	if err := selection.Validate(); err != nil {
		return err
	}

	if onFindSelection == nil {
		return fmt.Errorf("onFindSelection cannot be nil")
	}

	at.lock.RLock()
	defer at.lock.RUnlock()

	items := []any{}
	callback := func(item any) {
		items = append(items, item)
	}

	// Select all items in the tree
	if selection.Where == nil && len(selection.And) == 0 && len(selection.Or) == 0 {
		at.idTree.Iterate(callback)
	}

	if selection.Where != nil {

	}

	onFindSelection(items)
	return nil
}

func (at *threadsafeAssociatedTree) selectAll(onFindSelection datastructures.OnFindSelection) {
	items := []any{}
	at.idTree.Iterate(func(item any) {
		items = append(items, item)
	})
}

func (at *threadsafeAssociatedTree) selectWhere(query *query.Query, callback datastructures.OnFind, eonFindSelection datastructures.OnFindSelection) {

}
