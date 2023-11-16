package btreeassociatedquery

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var ErrorAssociatedIDAlreadyExists = fmt.Errorf("associated id already exists")
var ErrorCreateFailureQueryExist = fmt.Errorf("all query items already exist for another query")

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

// Open question. Do I actually want an "onCreate" callback, or can this just take a proper value?
// If there is already a query, that matches all the possible values, should I create 2 items in the tree?
//
// A. I think it is an error if there is alreay an item in the tree that matches all the queries of an item being created

// Create a new value in the Association Tree. This is thread safe to call with any other functions on the same object.
//
// NOTE: The KeyValues cannot contain the reserve key '_associated_id'
//
//	PARAMS:
//	- query - query is matched against datatypes.KeyValues and when it is true, the value of onCreate is called
//	- onCreate - is the callback used to create the value if it doesn't already exist in the tree. This must return nil, if creatiing the object failed.
//
//	RETURNS:
//	- string - the _associatted_id when an object is created or found. Will be the empty string if an error returns
//	- error - any errors encountered with the parameters or the keys are already taken
func (tsat *threadsafeAssociatedQueryTree) Create(query datatypes.AssociatedKeyValuesQuery, onCreate datastructures.OnCreate) (string, error) {
	if err := query.Validate(); err != nil {
		return "", err
	}
	if onCreate == nil {
		return "", fmt.Errorf("onCreate cannot be nil")
	}

	var idNodesList [][]*threadsafeIDNode
	insertableKeyValues := convertAssociatedKeyValuesQuery(query)

	// for all possible query joins, setup a collection that points to the same _associated_id
	var newIDNodes []*threadsafeIDNode
	for _, insertableKeyValue := range insertableKeyValues {
		newIDNodes = []*threadsafeIDNode{}

		sortedInsertableKeyValue := insertableKeyValue.SortedKeys()

		// these are needed to insert into the tree for a single value of the _associiated_id
		for _, key := range sortedInsertableKeyValue {
			tsat.keys.CreateOrFind(datatypes.String(key), createKey(findKey(&newIDNodes, insertableKeyValue[key])), findKey(&newIDNodes, insertableKeyValue[key]))
		}

	}

	// append all the ID nodes that make up this section of the query
	if newIDNodes != nil {
		idNodesList = append(idNodesList, newIDNodes)
	}

	// check to see what ID if any is part of all the sets
	var idSet set.Set[string]

	// for each list of the idnodes
	for _, idNodes := range idNodesList {
		// check each id node to get the list of possible _associated_ids
		for _, idNode := range idNodes {
			idNode.lock.RLock()

			if idSet == nil {
				// add al possible _associated_ids it could be
				idSet = set.New[string](idNode.ids.Values()...)
			} else {
				// intersect with the previous ids to trim down the values
				idSet.Intersection(idNode.ids.Values())
			}

			idNode.lock.RUnlock()
		}
	}

	// the query already exist with an item. need to return an error because we failed to create
	if idSet != nil && idSet.Size() >= 1 {
		// decrement the creating counter
		for _, idNodes := range idNodesList {
			for _, idNode := range idNodes {
				idNode.creating.Add(-1)
			}
		}

		return "", ErrorCreateFailureQueryExist
	}

	return "", nil
}
