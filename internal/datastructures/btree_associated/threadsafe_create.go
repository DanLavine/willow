package btreeassociated

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
)

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
func findKey(idNodes *[]*threadsafeIDNode, value datatypes.EncapsulatedValue) func(item any) {
	return func(item any) {
		valuesNode := item.(*threadsafeValuesNode)

		if err := valuesNode.values.CreateOrFind(value, createValue(idNodes), findValue(idNodes)); err != nil {
			panic(err)
		}
	}
}

// callback when creating a new value node when searching for a "key"
func createKey(onFind btree.BTreeOnFind) func() any {
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
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. datatypes.KeyValuesErr // error with the keyValues
//	          2. ErrorKeyValuesHasAssociatedID
//	          3. ErrorOnCreateNil
//	          4. ErrorAssociatedIDAlreadyExists
//	          5. ErrorKeyValuesAlreadyExists
//	          6. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) Create(keyValues datatypes.KeyValues, onCreate BTreeAssociatedOnCreate) (string, error) {
	// check parameters
	if err := keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return "", err
	}
	if onCreate == nil {
		return "", ErrorOnCreateNil
	}

	// tree destroying check
	if !tsat.destroySyncer.GuardOperation() {
		return "", ErrorTreeDestroying
	}
	defer tsat.destroySyncer.ClearOperation()

	keyValuesLen := len(keyValues)
	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	for _, key := range keyValues.SortedKeys() {
		if err := tsat.keys.CreateOrFind(datatypes.String(key), createKey(findKey(&idNodes, keyValues[key])), findKey(&idNodes, keyValues[key])); err != nil {
			panic(err)
		}
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

		return "", ErrorKeyValuesAlreadyExists
	}

	// always save the new IDs so we can unlock the IDNodes
	newID := tsat.idGenerator.ID()
	for _, idNode := range idNodes {
		idNode.ids[keyValuesLen-1] = append(idNode.ids[keyValuesLen-1], newID)
	}

	if newValue := onCreate(); newValue != nil {
		onCreate := func() any {
			newKeyValuesPair := datatypes.KeyValues{}
			for key, value := range keyValues {
				newKeyValuesPair[key] = value
			}

			return &associatedKeyValues{
				associatedID: newID,
				keyValues:    newKeyValuesPair,
				value:        newValue,
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
		if err := tsat.Delete(keyValues, nil); err != nil {
			panic(err)
		}

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
//	- error - any erResrors encountered with paraeters or destroy in progress
//	          1. datatypes.KeyValuesErr // error with the keyValues
//	          2. ErrorAssociatedIDEmpty
//	          3. ErrorKeyValuesHasAssociatedID
//	          4. ErrorOnCreateNil
//	          5. ErrorAssociatedIDAlreadyExists
//	          6. ErrorKeyValuesAlreadyExists
//	          7. ErrorTreeDestroying
func (tsat *threadsafeAssociatedTree) CreateWithID(associatedID string, keyValues datatypes.KeyValues, onCreate BTreeAssociatedOnCreate) error {
	// parameters checks
	if err := keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return err
	}
	if associatedID == "" {
		return ErrorAssociatedIDEmpty
	}
	if onCreate == nil {
		return ErrorOnCreateNil
	}

	// tree destroying check
	if !tsat.destroySyncer.GuardOperation() {
		return ErrorTreeDestroying
	}
	defer tsat.destroySyncer.ClearOperation()

	keyValuesLen := len(keyValues)
	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	for _, key := range keyValues.SortedKeys() {
		if err := tsat.keys.CreateOrFind(datatypes.String(key), createKey(findKey(&idNodes, keyValues[key])), findKey(&idNodes, keyValues[key])); err != nil {
			panic(err)
		}
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

		return ErrorKeyValuesAlreadyExists
	}

	// always save the new IDs so we can unlock the IDNodes
	for _, idNode := range idNodes {
		idNode.ids[keyValuesLen-1] = append(idNode.ids[keyValuesLen-1], associatedID)
	}

	if newValue := onCreate(); newValue != nil {
		onCreate := func() any {
			newKeyValuesPair := datatypes.KeyValues{}
			for key, value := range keyValues {
				newKeyValuesPair[key] = value
			}

			return &associatedKeyValues{
				associatedID: associatedID,
				keyValues:    newKeyValuesPair,
				value:        newValue,
			}
		}

		if err := tsat.associatedIDs.Create(datatypes.String(associatedID), onCreate); err != nil {
			// unlock the IDs
			for _, idNode := range idNodes {
				idNode.creating.Add(-1)
				idNode.lock.Unlock()
			}

			switch err {
			case btree.ErrorKeyAlreadyExists:
				// call delete for the item to clean any newly created items
				if err = tsat.deleteKeyValues(associatedID, keyValues); err != nil {
					panic(err)
				}

				return ErrorAssociatedIDAlreadyExists
			default:
				panic(err)
			}
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
// NOTE: The keyValues cannot contain the reserve key '_associated_id'
//
//	PARAMS:
//	- keyValues - is a map of key value pairs that compose an object's identity
//	- onCreate - is the callback used to create the value if it doesn't already exist in the tree. This must return nil, if creatiing the object failed.
//	- onFind - is the callback used when an item is found in the tree. It will recive the object's value saved in the tree (what was originally provided)
//
//	RETURNS:
//	- string - the _associatted_id when an object is created or found. Will be the empty string if an error returns
//	- error - any errors encountered with paraeters or destroy in progress
//	          1. datatypes.KeyValuesErr // error with the keyValues
//	          2. ErrorKeyValuesHasAssociatedID
//	          3. ErrorOnCreateNil
//	          4. ErrorAssociatedIDAlreadyExists
//	          5. ErrorKeyValuesAlreadyExists
//	          6. ErrorTreeDestroying
//	          7. ErrorTreeItemDestroying
func (tsat *threadsafeAssociatedTree) CreateOrFind(keyValues datatypes.KeyValues, onCreate BTreeAssociatedOnCreate, onFind BTreeAssociatedOnFind) (string, error) {
	if err := keyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return "", err
	}
	if onCreate == nil {
		return "", ErrorOnCreateNil
	}
	if onFind == nil {
		return "", ErrorOnFindNil
	}

	// tree destroying check
	if !tsat.destroySyncer.GuardOperation() {
		return "", ErrorTreeDestroying
	}
	defer tsat.destroySyncer.ClearOperation()

	// always attempt a find first so we only need read locks
	associatedID := ""
	wrappedOnFind := func(associatedKeyValues AssociatedKeyValues) bool {
		associatedID = associatedKeyValues.AssociatedID()
		onFind(associatedKeyValues)
		return false
	}
	if err := tsat.QueryAction(queryassociatedaction.KeyValuesToExactAssociatedActionQuery(keyValues), wrappedOnFind); err != nil {
		return "", err
	}
	if associatedID != "" {
		return associatedID, nil
	}

	// At this point we are 99%+ going to create the values, so our IDNodes need to use a write lock
	keyValuesLen := len(keyValues)
	var idNodes []*threadsafeIDNode

	// sort the keys so their won't be any deadlocks if everything goes smoothly
	for _, key := range keyValues.SortedKeys() {
		tsat.keys.CreateOrFind(datatypes.String(key), createKey(findKey(&idNodes, keyValues[key])), findKey(&idNodes, keyValues[key]))
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
	bTreeFind := func(key datatypes.EncapsulatedValue, item any) bool {
		onFind(item.(AssociatedKeyValues))
		return false
	}
	if idSet != nil && idSet.Size() == 1 {
		if err := tsat.associatedIDs.Find(datatypes.String(idSet.Values()[0]), v1common.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, bTreeFind); err != nil {
			// unlock the IDs
			for _, idNode := range idNodes {
				idNode.creating.Add(-1)
				idNode.lock.Unlock()
			}

			switch err {
			case btree.ErrorKeyDestroying:
				// this is fine, just need to report that the value is being destroyed
				return "", ErrorTreeItemDestroying
			default:
				panic(err)
			}
		}

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
			newKeyValuesPair := datatypes.KeyValues{}
			for key, value := range keyValues {
				newKeyValuesPair[key] = value
			}

			return &associatedKeyValues{
				associatedID: newID,
				keyValues:    newKeyValuesPair,
				value:        newValue,
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
