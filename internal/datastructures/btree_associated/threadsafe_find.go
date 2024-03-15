package btreeassociated

// This now needs to account for the 'Any' Keys saved in the DB, but I wanted to get rid of it eventually.
// so for now just comment it out as I want interactions to be driven throught the queries

/*
import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Find an item in the assoociation tree. This is thread safe to call with any other functions on the same object.
//
//	PARAMETERS:
//	- keyValuePair - is a map of key value pairs that compose an object's identity
//	- onFind - is the callback used when an item is found in the tree. It will recive the object's value saved in the tree (what was originally provided)
//
//	RETURNS:
//	- string - _associated_id if the item is found
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. datatypes.KeyValuesErr // error with the keyValues
//	          2. ErrorKeyValuesHasAssociatedID
//	          3. ErrorOnFindNil
//	          4. ErrorTreeItemDestroying
//	          5. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) Find(keyValues datatypes.KeyValues, onFind BTreeAssociatedOnFind) error {
	// parameter check
	if err := keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return err
	}
	if onFind == nil {
		return ErrorOnFindNil
	}

	// destroying check
	tsat.readWriteWG.Add(1)
	defer tsat.readWriteWG.Add(-1)

	if tsat.destroying.Load() {
		return ErrorTreeDestroying
	}

	var idSet set.Set[string]
	keyValuesLen := len(keyValues)
	sortedKeys := keyValues.SortedKeys()

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

		if len(idNode.ids) >= keyValuesLen {
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuesLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuesLen-1])
			}
		} else {
			failFast = true
		}
	}

	// callback for when a "key" is found
	findKey := func(value datatypes.EncapsulatedValue) func(item any) {
		return func(item any) {
			valuesNode := item.(*threadsafeValuesNode)
			if err := valuesNode.values.Find(value, findValue); err != nil {
				panic(err)
			}
		}
	}

	// filter all the key value pairs into one specific id to lookup
	for _, key := range sortedKeys {
		if err := tsat.keys.Find(datatypes.String(key), findKey(keyValues[key])); err != nil {
			panic(err)
		}

		if failFast {
			break
		}
	}

	wrappedFind := func(item any) {
		onFind(item.(AssociatedKeyValues))
	}

	// if there is only 1 value in the set, then we know that we found the desired object
	if !failFast && counter == keyValuesLen && idSet != nil && idSet.Size() == 1 {
		if err := tsat.associatedIDs.Find(datatypes.String(idSet.Values()[0]), wrappedFind); err != nil {
			switch err {
			case btree.ErrorKeyDestroying:
				return ErrorTreeDestroying
			default:
				panic(err)
			}
		} else {
			return nil
		}
	}

	return nil
}

// Find an item in the assoociation tree by the assocaitedID generated at creation. This is thread safe to call with any other functions on the same object.
//
//	PARAMS:
//	- associatedID - is the associated id generated at creation time
//	- onFind - is the callback used when an item is found in the tree. It will recive the object's value saved in the tree (what was originally provided)
//
//	RETURNS:
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. ErrorAssociatedIDEmpty
//	          2. ErrorOnFindNil
//	          3. ErrorKeyValuesAlreadyExists
//	          4. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) FindByAssociatedID(associatedID string, onFind BTreeAssociatedOnFind) error {
	// parameter checks
	if associatedID == "" {
		return ErrorAssociatedIDEmpty
	}
	if onFind == nil {
		return ErrorOnFindNil
	}

	// destroy checks
	tsat.readWriteWG.Add(1)
	defer tsat.readWriteWG.Add(-1)

	if tsat.destroying.Load() {
		return ErrorTreeDestroying
	}

	wrappedFind := func(item any) {
		onFind(item.(AssociatedKeyValues))
	}

	if err := tsat.associatedIDs.Find(datatypes.String(associatedID), wrappedFind); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			return ErrorTreeDestroying
		default:
			panic(err)
		}
	}

	return nil
}
*/
