package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// TODO: remove this?
// The locks on this are not right. IF 2 requests come in at the exact same time trying to create the same keys,
// then the could both call the onCreate() function which isn't what we want... What is the best way to structure this?
//
// could go back to the generate *bool to know if things need to be created. how annoying
func (at *associatedTree) CreateOrFind(keyValuePairs datatypes.StringMap, onCreate datastructures.OnCreate, onFind datastructures.OnFind) (any, error) {
	if len(keyValuePairs) == 0 {
		return nil, fmt.Errorf("keyValuePairs cannot be empty")
	}

	if onCreate == nil {
		return nil, fmt.Errorf("onCreate cannot be empty")
	}

	if item, _ := at.Get(keyValuePairs, onFind); item != nil {
		return item, nil
	}

	// first find the "compositKeyValue" gropuings where our tags might reside.
	// note this won't return an error since its all internal known good values
	castableAssociatedKeyValues, _ := at.groupedKeyValueAssociation.CreateOrFind(datatypes.Int(len(keyValuePairs)), keyValuesLock, newKeyValues)
	associatedKeyValues := castableAssociatedKeyValues.(*keyValues)

	// NOTE: This lock is important to have when we are here. This is because maps are unordered, so we need a lock on the entire
	// associateion when inserting
	defer associatedKeyValues.lock.Unlock()

	// items needed to keep track of either a create or find process
	firstLoop := true
	idSet := set.New()
	var idHolders []*idHolder

	for createKey, createValue := range keyValuePairs {
		// create or find the values for a particular key
		castableValuesForKey, _ := associatedKeyValues.values.CreateOrFind(createKey, nil, newKeyValues)
		valuesForKey := castableValuesForKey.(*keyValues)

		// filter the IDs associated with the particualr value
		var castableIDHolder any
		if firstLoop {
			castableIDHolder, _ = valuesForKey.values.CreateOrFind(createValue, onFindIDHolderAdd(idSet), newIDHolder)
			firstLoop = false
		} else {
			castableIDHolder, _ = valuesForKey.values.CreateOrFind(createValue, onFindIDHolderKeep(idSet), newIDHolderClearSet(idSet))
		}

		idHolders = append(idHolders, castableIDHolder.(*idHolder))
	}

	// must have been a race where 2 requests are creating the same item
	if idSet.Len() == 1 {
		item := at.idTree.Get(idSet.Values()[0])
		if onFind != nil {
			onFind(item)
		}

		return item, nil
	}

	// need to create the new value and save the ID
	newValue, err := onCreate()
	if err != nil {
		at.cleanFaildCreate(keyValuePairs)
		return nil, err
	}

	id := at.idTree.Add(newValue)
	for _, idHolder := range idHolders {
		idHolder.add(id)
	}

	return newValue, nil
}

func (at *associatedTree) cleanFaildCreate(keyValuePairs datatypes.StringMap) {
	// first find the "associatedKeyValues" gropuings where our tags might reside. This will always be available so cast is fine
	associatedKeyValues := at.groupedKeyValueAssociation.Find(datatypes.Int(len(keyValuePairs)), nil).(*keyValues)

	for createKey, createValue := range keyValuePairs {
		// find any values associated with the key. If we are here the value is guranteed to be created so cast is fine
		valuesForKey := associatedKeyValues.values.Find(createKey, nil).(*keyValues)

		valuesForKey.values.Delete(createValue, canRemoveIDHolder)       // attempt to delete value + idHolder
		associatedKeyValues.values.Delete(createKey, keyValuesCanRemove) // attempt to delete the key if there are no more values
	}

	at.groupedKeyValueAssociation.Delete(datatypes.Int(len(keyValuePairs)), keyValuesCanRemove) // attempt to delete the associatedKeyValues
}
