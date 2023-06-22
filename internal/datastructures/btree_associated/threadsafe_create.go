package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func (at *threadsafeAssociatedTree) CreateOrFind(keyValuePairs datatypes.StringMap, onCreate datastructures.OnCreate, onFind datastructures.OnFind) error {
	if len(keyValuePairs) == 0 {
		return fmt.Errorf("keyValuePairs cannot be empty")
	}

	if onCreate == nil {
		return fmt.Errorf("onCreate cannot be nil")
	}

	if onFind == nil {
		return fmt.Errorf("onFind cannot be nil")
	}

	// always attempt a read first since these are read locked
	found := false
	wrappedOnFind := func(item any) {
		found = true
		onFind(item)
	}
	_ = at.Find(keyValuePairs, wrappedOnFind)
	if found == true {
		return nil
	}

	// need to 'probably' create the item
	at.lock.Lock()
	defer at.lock.Unlock()

	var idHolders []*idHolder

	// record all the id holders
	findIDHolder := func(item any) {
		idHolder := item.(*idHolder)
		idHolders = append(idHolders, idHolder)
	}

	// callback when the key is found
	findKey := func(createValue datatypes.EncapsulatedData) func(item any) {
		return func(item any) {
			valuesForKey := item.(*keyValues)
			_ = valuesForKey.values.CreateOrFind(createValue, newIDHolder(findIDHolder), findIDHolder)
		}
	}

	// callback for the associated key value groups
	findAssociatedKeyValueGroup := func(item any) {
		associatedKeyValues := item.(*keyValues)

		// create or find the values for all key value pairs
		for createKey, createValue := range keyValuePairs {
			associatedKeyValues.values.CreateOrFind(datatypes.String(createKey), newKeyValues(findKey(createValue)), findKey(createValue))
		}
	}

	at.groupedKeyValueAssociation.CreateOrFind(
		datatypes.Int(len(keyValuePairs)),
		newKeyValues(findAssociatedKeyValueGroup),
		findAssociatedKeyValueGroup,
	)

	// determine if all the idholders hold a common id
	idSet := set.New[uint64](idHolders[0].IDs...)
	for i := 1; i < len(idHolders); i++ {
		idSet.Intersection(idHolders[i].IDs)
	}

	// must have been a race where 2 requests are creating the same item
	if idSet.Size() == 1 {
		onFind(at.idTree.Get(idSet.Values()[0]))
		return nil
	}

	// need to create the new value and save the ID
	if newValue := onCreate(); newValue != nil {
		id := at.idTree.Add(newValue)
		for _, idHolder := range idHolders {
			idHolder.add(id)
		}
	} else {
		at.cleanFaildCreate(keyValuePairs)
	}

	return nil
}

func (at *threadsafeAssociatedTree) cleanFaildCreate(keyValuePairs datatypes.StringMap) {
	// delete the Value
	deleteValue := func(item any) bool {
		idHolder := item.(*idHolder)
		return len(idHolder.IDs) == 0
	}

	// delete the Key
	deleteKey := func(value datatypes.EncapsulatedData) func(item any) bool {
		return func(item any) bool {
			valuesForKey := item.(*keyValues)
			valuesForKey.values.Delete(value, deleteValue)

			return valuesForKey.values.Empty()
		}
	}

	// delete the associated key values
	deleteAssociatedKeyValues := func(item any) bool {
		keyValues := item.(*keyValues)

		for createKey, createValue := range keyValuePairs {
			keyValues.values.Delete(datatypes.String(createKey), deleteKey(createValue))
		}

		return keyValues.values.Empty()
	}

	_ = at.groupedKeyValueAssociation.Delete(datatypes.Int(len(keyValuePairs)), deleteAssociatedKeyValues)
}
