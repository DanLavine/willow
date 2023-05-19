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
	if len(keyValues) == 0 {
		return nil, fmt.Errorf("keyValues cannot be empty")
	}

	if onCreate == nil {
		return nil, fmt.Errorf("onCreate cannot be empty")
	}

	//query := v1.Query{
	//	Matches: v1.Matches{
	//		Type: v1.Strict,
	//		Where: []v1.WhereClause{
	//			{KeyValuePairs: keyValues},
	//		},
	//	},
	//}

	//findResults := ct.Find(query, onFind)
	//if findResults == nil {
	//	// nothing to do here will need to create the item
	//} else if len(findResults) == 1 {
	//	// found the item
	//	return findResults[0], nil
	//}

	// first find the "compositColumn" gropuings where our tags might reside.
	castableKeyValues, err := ct.compositeColumns.CreateOrFind(datatypes.Int(len(keyValues)), KeyValuesLock, NewKeyValues)
	if err != nil {
		return nil, err
	}
	recordedKeyValues := castableKeyValues.(*KeyValues)
	defer recordedKeyValues.lock.Unlock() //lock here since the map of keyValues is in a random order

	// items needed to keep track of either a create or find process
	firstLoop := true
	idSet := set.New()
	var idHolders []*idHolder

	for searchKey, searchValue := range keyValues {
		// create or find the values for a particular key
		fmt.Println("finidng key:", searchKey)
		castableValues, err := recordedKeyValues.Values.CreateOrFind(searchKey, nil, NewKeyValues)
		if err != nil {
			return nil, err
		}
		valuesForKey := castableValues.(*KeyValues)

		// filter the IDs associated with the particualr value
		var castableIDHolder any
		if firstLoop {
			castableIDHolder, err = valuesForKey.Values.CreateOrFind(searchValue, onFindIDHolderAdd(idSet), newIDHolder)
			firstLoop = false
		} else {
			castableIDHolder, err = valuesForKey.Values.CreateOrFind(searchValue, onFindIDHolderKeep(idSet), newIDHolderClearSet(idSet))
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

	fmt.Println(`////////////////////////////////////////////////////////////`)

	return newValue, nil
}

func (ct *compositeTree) cleanFaildCreate(keyValues map[datatypes.String]datatypes.String) {
	// first find the "compositColumn" gropuings where our tags might reside.
	castableKeyValues := ct.compositeColumns.Find(datatypes.Int(len(keyValues)), nil)
	recordedKeyValues := castableKeyValues.(*KeyValues)

	for createKey, createValue := range keyValues {
		// find any values associated with the key
		castableValues := recordedKeyValues.Values.Find(createKey, nil)
		valuesForKey := castableValues.(*KeyValues)

		valuesForKey.Values.Delete(createValue, canRemoveIDHolder)     // attempt to delete value + idHolder
		recordedKeyValues.Values.Delete(createKey, CanRemoveKeyValues) // attempt to delete the key if there are no more values
	}

	ct.compositeColumns.Delete(datatypes.Int(len(keyValues)), canDeleteCompositeColumns) // attempt to delete the compositeColumns
}
