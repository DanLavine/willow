package compositetree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
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

	query := v1.Query{
		Matches: v1.Matches{
			StrictMatches: keyValues,
		},
	}

	findResults := ct.Find(query, onFind)
	if findResults == nil {
		// nothing to do here will need to create the item
	} else if len(findResults) == 1 {
		return findResults[0], nil
	} else {
		panic("remove me, how did we get here!")
	}

	ct.lock.Lock()
	defer ct.lock.Unlock()

	idSet := set.New()
	create := false
	var idHolders []*idHolder

	// first find the "compositColumn" gropuings where our tags might reside.
	castableCompositeColumn, _ := ct.compositeColumns.CreateOrFind(datatypes.Int(len(keyValues)), nil, createCompositeColumn)
	compositeColumn := castableCompositeColumn.(*compositeColumn)

	for createKey, createValue := range keyValues {
		// create or find the key
		castableKeyValuePairs, _ := compositeColumn.keyValuePairs.CreateOrFind(createKey, nil, createKeyValuePairs)
		keyValuePairs := castableKeyValuePairs.(*keyValuePairs)

		// create or find the ID holder from the value
		castableIDHolder, _ := keyValuePairs.idHolders.CreateOrFind(createValue, findIDHolder(idSet), createIDHolder(&create))
		idHolders = append(idHolders, castableIDHolder.(*idHolder))
	}

	// need to create the new value and save the ID
	if create {
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

	// should have only 1 id at this point. so its a find
	// this happens on races where 2 item try to create the same keys simultaniously
	if idSet.Len() != 1 {
		panic("extra values in the set!")
	}

	item := ct.idTree.Get(idSet.Values()[0])
	if onFind != nil {
		onFind(item)
	}

	return item, nil
}

func (ct *compositeTree) cleanFaildCreate(keyValues map[datatypes.String]datatypes.String) {
	// first find the "compositColumn" gropuings where our tags might reside.
	castableCompositeColumn, _ := ct.compositeColumns.CreateOrFind(datatypes.Int(len(keyValues)), nil, createCompositeColumn)
	compositeColumn := castableCompositeColumn.(*compositeColumn)

	for createKey, createValue := range keyValues {
		// create or find the key
		castableKeyValuePairs, _ := compositeColumn.keyValuePairs.CreateOrFind(createKey, nil, createKeyValuePairs)
		keyValuePairs := castableKeyValuePairs.(*keyValuePairs)

		keyValuePairs.idHolders.Delete(createValue, canDeleteIDHolder)          // attempt to delete the idHolder
		compositeColumn.keyValuePairs.Delete(createKey, canDeleteKeyValuePairs) // attempt to delete the keys
	}

	ct.compositeColumns.Delete(datatypes.Int(len(keyValues)), canDeleteCompositeColumns) // attempt to delete the compositeColumns
}
