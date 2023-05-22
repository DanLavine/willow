package compositetree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// The locks on this are not right. IF 2 requests come in at the exact same time trying to create the same keys,
// then the could both call the onCreate() function which isn't what we want... What is the best way to structure this?
//
// could go back to the generate *bool to know if things need to be created. how annoying
func (ct *compositeTree) CreateOrFind(keyValues map[datatypes.String]datatypes.String, onCreate datastructures.OnCreate, onFind datastructures.OnFind) (any, error) {
	if onCreate == nil {
		return nil, fmt.Errorf("onCreate cannot be empty")
	}

	//findResults := ct.FindStrict(keyValues, onFind)
	//if findResults == nil {
	//	// nothing to do here will need to create the item
	//} else {
	//	// found the item
	//	return findResults.value, nil
	//}

	switch len(keyValues) {
	case 0:
		// 0 is a special case as it doesn't have any associated tags. It can be used as a sort of global object if that
		// makes sense. For example when doing a Query with no 'WHERE' clause, this object could be returned, or possibly
		// each object within the tree. Think both use cases make sense. But in the case of each object, we can just use an
		// iterator. That is for the caller to decided on what the query should actually call
		castableGlobalValue, err := ct.compositeColumns.CreateOrFind(datatypes.Int(0), globalOnFind(onFind), globalOnCreate(onCreate))
		if err != nil {
			return nil, err
		}
		return castableGlobalValue.(*globalValue).value, nil
	default:
		// first find the "compositColumn" gropuings where our tags might reside.
		// NOTE: we need a lock here since the tags are in a random order for creation
		castableCompositeKeyValues, err := ct.compositeColumns.CreateOrFind(datatypes.Int(len(keyValues)), compositeKeyValuesLock, newCompositeKeyValues)
		if err != nil {
			return nil, err
		}
		knownCompositeKeyValues := castableCompositeKeyValues.(*compositeKeyValues)
		defer knownCompositeKeyValues.lock.Unlock()

		// items needed to keep track of either a create or find process
		firstLoop := true
		idSet := set.New()
		var idHolders []*idHolder

		for searchKey, searchValue := range keyValues {
			// create or find the values for a particular key
			castableValues, err := knownCompositeKeyValues.values.CreateOrFind(searchKey, nil, newCompositeKeyValues)
			if err != nil {
				return nil, err
			}
			valuesForKey := castableValues.(*compositeKeyValues)

			// filter the IDs associated with the particualr value
			var castableIDHolder any
			if firstLoop {
				castableIDHolder, err = valuesForKey.values.CreateOrFind(searchValue, onFindIDHolderAdd(idSet), newIDHolder)
				firstLoop = false
			} else {
				castableIDHolder, err = valuesForKey.values.CreateOrFind(searchValue, onFindIDHolderKeep(idSet), newIDHolderClearSet(idSet))
			}

			if err != nil {
				return nil, err
			}

			idHolders = append(idHolders, castableIDHolder.(*idHolder))
		}

		// must have been a race where 2 requests are creating the same item
		if idSet.Len() == 1 {
			item := ct.idTree.Get(idSet.Values()[0])
			if onFind != nil {
				onFind(item)
			}

			return item, nil
		}

		// need to create the new value and save the ID
		newValue, err := onCreate()
		if err != nil {
			ct.cleanFaildCreate(keyValues)
			return nil, err
		}

		id := ct.idTree.Add(newValue)
		for _, idHolder := range idHolders {
			idHolder.add(id)
		}

		return newValue, nil
	}
}

func (ct *compositeTree) cleanFaildCreate(keyValues map[datatypes.String]datatypes.String) {
	// first find the "compositColumn" gropuings where our tags might reside.
	castableKeyValues := ct.compositeColumns.Find(datatypes.Int(len(keyValues)), nil)
	knownCompositeKeyValues := castableKeyValues.(*compositeKeyValues)

	for createKey, createValue := range keyValues {
		// find any values associated with the key
		castableValues := knownCompositeKeyValues.values.Find(createKey, nil)
		valuesForKey := castableValues.(*compositeKeyValues)

		valuesForKey.values.Delete(createValue, canRemoveIDHolder)                      // attempt to delete value + idHolder
		knownCompositeKeyValues.values.Delete(createKey, cleanFailedCompositeKeyValues) // attempt to delete the key if there are no more values
	}

	ct.compositeColumns.Delete(datatypes.Int(len(keyValues)), cleanFailedCompositeKeyValues) // attempt to delete the compositeColumns
}
