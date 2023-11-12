package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var ErrorAssociatedIDAlreadyExists = fmt.Errorf("associated id already exists")
var ErrorCreateFailureKeyValuesExist = fmt.Errorf("keyValues already exist with an associated item")

// callback for when a "value" is found
func findValue(idNodes *[]*threadsafeIDNode) func(item any) {
	return func(item any) {
		idNode := item.(*threadsafeIDNode)
		idNode.lock.Lock()
		defer idNode.lock.Unlock()

		idNode.creating.Add(1)
		*idNodes = append(*idNodes, idNode)
	}
}

// callback for when a "value" needs to be created
func createValue(idNodes *[]*threadsafeIDNode) func() any {
	return func() any {
		newIDNode := newIDNode()
		findValue(idNodes)(newIDNode)

		return newIDNode
	}
}

// callback for when a "key" is found
func findKey(idNodes *[]*threadsafeIDNode, value datatypes.EncapsulatedData) func(item any) {
	return func(item any) {
		valuesNode := item.(*threadsafeValuesNode)

		if err := valuesNode.values.CreateOrFind(value, createValue(idNodes), findValue(idNodes)); err != nil {
			panic(err)
		}
	}
}

// callback when creating a new value node when searching for a "key"
func createKey(onFind datastructures.OnFind) func() any {
	return func() any {
		newValueNode := newValuesNode()
		onFind(newValueNode)

		return newValueNode
	}
}

// Create a new value in the Association Tree. This is thread safe to call with any other functions on the same object.
//
// NOTE: The KeyValues cannot contain the reserve key '_associated_id'
//
//	PARAMS:
//	- KeyValues - is a map of key value pairs that compose an object's identity
//	- onCreate - is the callback used to create the value if it doesn't already exist in the tree. This must return nil, if creatiing the object failed.
//
//	RETURNS:
//	- string - the _associatted_id when an object is created or found. Will be the empty string if an error returns
//	- error - any errors encountered with the parameters
func (tsat *threadsafeAssociatedTree) Create(keyValues KeyValues, onCreate datastructures.OnCreate) (string, error) {
	keyValuesLen := len(keyValues)
	if keyValuesLen == 0 {
		return "", fmt.Errorf("keyValues cannot be empty")
	}
	if keyValues.HasAssociatedID() {
		return "", fmt.Errorf("keValues cannot contain a Key with the _associated_id reserved key word")
	}
	if onCreate == nil {
		return "", fmt.Errorf("onCreate cannot be nil")
	}

	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	for _, key := range keyValues.SortedKeys() {
		tsat.keys.CreateOrFind(key, createKey(findKey(&idNodes, keyValues[key])), findKey(&idNodes, keyValues[key]))
	}

	var idSet set.Set[string]
	for _, idNode := range idNodes {
		idNode.lock.Lock()

		if len(idNode.ids) < keyValuesLen {
			// need to create the new association
			for i := len(idNode.ids); i < keyValuesLen; i++ {
				idNode.ids = append(idNode.ids, []string{})
			}

			// record that we are going to need a new ID
			if idSet == nil {
				idSet = set.New[string]()
			} else {
				idSet.Clear()
			}
		} else {
			// need to save the possible IDs it could be
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuesLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuesLen-1])
			}
		}
	}

	// the KeyValues already exist with an item. need to return an error because we failed to create
	if idSet != nil && idSet.Size() == 1 {
		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		return "", ErrorCreateFailureKeyValuesExist
	}

	// always save the new IDs so we can unlock the IDNodes
	newID := tsat.idGenerator.ID()
	for _, idNode := range idNodes {
		idNode.ids[keyValuesLen-1] = append(idNode.ids[keyValuesLen-1], newID)
	}

	if newValue := onCreate(); newValue != nil {
		onCreate := func() any {
			newKeyValuesPair := KeyValues{}
			for key, value := range keyValues {
				newKeyValuesPair[key] = value
			}

			newKeyValuesPair[datatypes.String(ReservedID)] = datatypes.String(newID)
			return &AssociatedKeyValues{
				keyValues: newKeyValuesPair,
				value:     newValue,
			}
		}

		if err := tsat.associatedIDs.Create(datatypes.String(newID), onCreate); err != nil {
			panic(fmt.Errorf("failed to use a unique key id: %w", err))
		}

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}
	} else {
		// failed to crete the new item to save.

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		// call delete for the item to clean any newly created items
		tsat.Delete(keyValues, nil)

		// on create failed, so return empty and nil. Caller should know?
		return "", nil
	}

	return newID, nil
}

// CreateWithID adds a new value to the Association Tree with an explicit ID to search for on subsequent calls.
// This is thread safe to call with any other functions on the same object.
//
//	PARAMS:
//	- KeyValues - is a map of key value pairs that compose an object's identity. This must have a single reserved '_associated_id' as part of the KeyValues
//	- onCreate - is the callback used to create the value if it doesn't already exist in the tree. This must return nil, if creatiing the object failed.
//
//	RETURNS:
//	- error - any errors encountered with
//	          1. the parameters
//	          2. the associatedID already exists
//	          3. the keyValues already exist
func (tsat *threadsafeAssociatedTree) CreateWithID(associatedID string, keyValues KeyValues, onCreate datastructures.OnCreate) error {
	keyValuesLen := len(keyValues)
	if keyValuesLen == 0 {
		return fmt.Errorf("keyValues cannot be empty")
	}
	if keyValues.HasAssociatedID() {
		return fmt.Errorf("keValues cannot contain a Key with the _associated_id reserved key word")
	}
	if onCreate == nil {
		return fmt.Errorf("onCreate cannot be nil")
	}

	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	for _, key := range keyValues.SortedKeys() {
		tsat.keys.CreateOrFind(key, createKey(findKey(&idNodes, keyValues[key])), findKey(&idNodes, keyValues[key]))
	}

	var idSet set.Set[string]
	for _, idNode := range idNodes {
		idNode.lock.Lock()

		if len(idNode.ids) < keyValuesLen {
			// need to create the new association
			for i := len(idNode.ids); i < keyValuesLen; i++ {
				idNode.ids = append(idNode.ids, []string{})
			}

			// record that we are going to need a new ID
			if idSet == nil {
				idSet = set.New[string]()
			} else {
				idSet.Clear()
			}
		} else {
			// need to save the possible IDs it could be
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuesLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuesLen-1])
			}
		}
	}

	// the KeyValues already exist with an item. need to return an error because we failed to create
	if idSet != nil && idSet.Size() == 1 {
		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		return ErrorCreateFailureKeyValuesExist
	}

	// always save the new IDs so we can unlock the IDNodes
	for _, idNode := range idNodes {
		idNode.ids[keyValuesLen-1] = append(idNode.ids[keyValuesLen-1], associatedID)
	}

	if newValue := onCreate(); newValue != nil {
		onCreate := func() any {
			newKeyValuesPair := KeyValues{}
			for key, value := range keyValues {
				newKeyValuesPair[key] = value
			}

			newKeyValuesPair[datatypes.String(ReservedID)] = datatypes.String(associatedID)
			return &AssociatedKeyValues{
				keyValues: newKeyValuesPair,
				value:     newValue,
			}
		}

		if err := tsat.associatedIDs.Create(datatypes.String(associatedID), onCreate); err != nil {
			// unlock the IDs
			for _, idNode := range idNodes {
				idNode.creating.Add(-1)
				idNode.lock.Unlock()
			}

			// call delete for the item to clean any newly created items
			if err = tsat.deleteKeyValues(associatedID, keyValues); err != nil {
				panic(err)
			}

			return ErrorAssociatedIDAlreadyExists
		}

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}
	} else {
		// failed to crete the new item to save.

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		// call delete for the item to clean any newly created items
		tsat.Delete(keyValues, nil)

		// on create failed, so return empty and nil. Caller should know?
		return nil
	}

	return nil
}

// CreateOrFind inserts or finds the value in the assoociation tree. This is thread safe to call with
// any other functions on the same object.
//
// NOTE: The keyValuePairs cannot contain the reserve key '_associated_id'
//
//	PARAMS:
//	- keyValuePairs - is a map of key value pairs that compose an object's identity
//	- onCreate - is the callback used to create the value if it doesn't already exist in the tree. This must return nil, if creatiing the object failed.
//	- onFind - is the callback used when an item is found in the tree. It will recive the object's value saved in the tree (what was originally provided)
//
//	RETURNS:
//	- string - the _associatted_id when an object is created or found. Will be the empty string if an error returns
//	- error - any errors encountered with the parameters
func (tsat *threadsafeAssociatedTree) CreateOrFind(keyValues KeyValues, onCreate datastructures.OnCreate, onFind datastructures.OnFind) (string, error) {
	keyValuesLen := len(keyValues)
	if keyValuesLen == 0 {
		return "", fmt.Errorf("keyValuePairs cannot be empty")
	}
	if keyValues.HasAssociatedID() {
		return "", fmt.Errorf("keValues cannot contain a Key with the _associated_id reserved key word")
	}
	if onCreate == nil {
		return "", fmt.Errorf("onCreate cannot be nil")
	}
	if onFind == nil {
		return "", fmt.Errorf("onFind cannot be nil")
	}

	// always attempt a find first so we only need read locks
	found := false
	wrappedOnFind := func(item any) {
		found = true
		onFind(item)
	}
	associatedID, err := tsat.Find(keyValues, wrappedOnFind)
	if err != nil {
		return "", err
	}
	if found {
		return associatedID, nil
	}

	// At this point we are 99%+ going to create the values, so our IDNodes need to use a write lock\
	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	for _, key := range keyValues.SortedKeys() {
		tsat.keys.CreateOrFind(key, createKey(findKey(&idNodes, keyValues[key])), findKey(&idNodes, keyValues[key]))
	}

	var idSet set.Set[string]
	for _, idNode := range idNodes {
		idNode.lock.Lock()

		if len(idNode.ids) < keyValuesLen {
			// need to create the new association
			for i := len(idNode.ids); i < keyValuesLen; i++ {
				idNode.ids = append(idNode.ids, []string{})
			}

			// record that we are going to need a new ID
			if idSet == nil {
				idSet = set.New[string]()
			} else {
				idSet.Clear()
			}
		} else {
			// need to save the possible IDs it could be
			if idSet == nil {
				idSet = set.New[string](idNode.ids[keyValuesLen-1]...)
			} else {
				idSet.Intersection(idNode.ids[keyValuesLen-1])
			}
		}
	}

	// must have been a race where 2 requests tried to create the same object
	if idSet != nil && idSet.Size() == 1 {
		tsat.associatedIDs.Find(datatypes.String(idSet.Values()[0]), onFind)

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		return idSet.Values()[0], nil
	}

	// always save the new IDs so we can unlock the IDNodes
	newID := tsat.idGenerator.ID()
	for _, idNode := range idNodes {
		idNode.ids[keyValuesLen-1] = append(idNode.ids[keyValuesLen-1], newID)
	}

	if newValue := onCreate(); newValue != nil {
		onCreate := func() any {
			newKeyValuesPair := KeyValues{}
			for key, value := range keyValues {
				newKeyValuesPair[key] = value
			}

			newKeyValuesPair[datatypes.String(ReservedID)] = datatypes.String(newID)
			return &AssociatedKeyValues{
				keyValues: newKeyValuesPair,
				value:     newValue,
			}
		}

		if err := tsat.associatedIDs.Create(datatypes.String(newID), onCreate); err != nil {
			panic(fmt.Errorf("found an id that already exists. Globaly unique ID failure: %s", newID))
		}

		// unlock the idNodes after successfull creation
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}
	} else {
		// failed to crete the new item to save.

		// unlock the IDs
		for _, idNode := range idNodes {
			idNode.creating.Add(-1)
			idNode.lock.Unlock()
		}

		// call delete for the item to clean everything up
		tsat.Delete(keyValues, nil)
	}

	return newID, nil
}
