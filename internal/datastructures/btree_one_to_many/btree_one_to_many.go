package btreeonetomany

import (
	"fmt"
	"strings"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var (
	ReservedID = "_associated_id"

	// on create callback errors
	ErrorOnCreateNil = fmt.Errorf("onCreate callback cannot be nil")

	// on find callback errors
	ErrorOnFindNil = fmt.Errorf("onFind callback cannot be nil")

	// general iteration errors
	ErrorOnIterateNil = fmt.Errorf("onIterate callback cannot be nil")

	// paramater id errors
	ErrorOneIDEmpty  = fmt.Errorf("oneID parameter cannot be the empty string")
	ErrorManyIDEmpty = fmt.Errorf("manyID parameter cannot be the empty string")

	// creation errors
	ErrorManyIDAlreadyExists               = fmt.Errorf("manyID already exists")
	ErrorManyKeyValuesAlreadyExist         = fmt.Errorf("keyValues for the Many relation already exist")
	ErrorManyKeyValuesContainsReservedKeys = fmt.Errorf("keyValues contain reserved keys that begin with an '_'")

	// ID destroying quick failures
	ErrorOneIDDestroying  = fmt.Errorf("oneID is in the process of being destroyed")
	ErrorManyIDDestroying = fmt.Errorf("manyID is in the process of being destroyed")
)

// Callback that is used to actaully create the item in the tree
//
//	RETURNS:
//	- any - the item to save in the tree. If this is nil, the item will not be saved in the tree
type OneToManyTreeOnCreate func() any

// Callback that is used when an item is found in the tree
//
//	PARAMETERS:
//	- associatedKeyValues - detailed information about the item saved in the tree, including a referance to the item itself
//
//	RETURNS:
//	- any - the item to save in the tree. If this is nil, the item will not be saved in the tree
type OneToManyTreeOnFind func(oneToManyItem OneToManyItem)

// Callback to check that an item can actually be removed from a tree
//
//	PARAMS:
//	- key - key for the item saved
//	- item - the original item saved to the bTree
//
//	RETURNS:
//	- bool - if true, will remove the item item from the tree. If this ever returns false when doing bulk deletions then
//	         the deletion operations will be halted. Any objects previously destroyed before the error will not be restored
type OneToManyTreeRemove func(oneToManyItem OneToManyItem) bool

// Callback when iterating over tree values
//
//	PARAMS:
//	- key - key for the item saved
//	- item - the original item saved to the bTree
//
//	RETURNS:
//	- bool - if true, will continue iterating thrrough the tree. If this ever returns false then the pagination is halted
type OneToManyTreeIterate func(oneToManyItem OneToManyItem) bool

//go:generate mockgen -destination=btreeonetomanyfakes/one_to_many_mock.go -package=btreeonetomanyfakes github.com/DanLavine/willow/internal/datastructures/btree_one_to_many BTreeOneToMany
type BTreeOneToMany interface {
	// create a new entry in the BTReeOneToMany
	//
	//	PARAMETERS:
	//	- oneID - the realtionship all created items belong to.
	//	- associatedID - associatedID for the KeyValues
	//	- keyValues - keyValues that define the object saved in relation to the oneID
	//	- onCreate - callback to run when the item is newly created
	//
	//	RETURNS:
	//	- error - error withe he parameters or the tree is already being destroyed
	//	        1. datatypes.KeyValuesErr // error with the keyValues
	//	        2. ErrorOneIDEmpty
	//	        3. ErrorManyIDEmpty
	//	        4. ErrorKeyValuesEmpty
	//	        5. ErrorManyKeyValuesContainsReservedKeys
	//	        6. ErrorOnCreateNil
	//	        7. ErrorManyIDAlreadyExists
	//	        8. ErrorManyKeyValuesAlreadyExist
	//	        9. ErrorOneIDDestroying
	//	        10. ErrorManyIDDestroying
	CreateWithID(oneID string, associatedID string, keyValues datatypes.KeyValues, onCreate OneToManyTreeOnCreate) error

	//	PARAMETERS:
	//	- oneID - relation id to query
	//	- query - query to search for any values related to the oneID
	//	- onIterate - callback to run on for any values that match the query
	//
	//	RETURNS:
	//	- error - error encountered with the
	//	        1. fmt.Errorf(...) // error with the query
	//	        2. ErrorOneIDEmpty
	//	        3. ErrorsOnIterateNil
	//	        4. ErrorOneIDDestroying
	//
	// Query for the one to many relations
	//
	// TODO: I believe that this should be changed to just a query and the oneID should be removed.
	// but there is some unkown around how do pagination properly. Becase this doesn't affect services api
	// right now, goint to do the easy query, but need to revist this once Willow is up and running
	//
	// Open quesstion as well. should this return an error if the oneID is not found? to be able to distinguish
	// if there is an issue with finding the OneID or the Query item?
	Query(oneID string, query datatypes.AssociatedKeyValuesQuery, onIterate OneToManyTreeIterate) error

	//	RETURNS:
	//	- error - error encountered with the
	//	        1. datatypes.KeyValuesErr // error with the keyValues
	//	        2. ErrorOneIDEmpty
	//	        3. ErrorsOnIterateNil
	//	        4. ErrorOneIDDestroying
	//
	// Match for the one to many relations
	//
	// TODO: I believe that this should be changed to just a match request and the oneID should be removed.
	// but there is some unkown around how do pagination properly. Becase this doesn't affect services api
	// right now, goint to do the easy match operation, but need to revist this once Willow is up and running
	MatchPermutations(oneID string, match datatypes.KeyValues, onPagination OneToManyTreeIterate) error

	//	PARAMETERS:
	//	- oneID - name of the entire relation tree to destroy
	//	- canDelete - optional callback to run for each value in the many relationship
	//
	//	RETURNS:
	//	- error - error for the destroy parameters, or another operation is in progress
	//	        - ErrorOneIDEmpty
	//	        - ErrorOneIDDestroying
	// remove the tree with all children from the One relations and therfore the Many as well
	DestroyOne(oneID string, canDelete OneToManyTreeRemove) error

	//	PARAMETERS:
	//	- oneID - id of the entire relation tree to destroy from
	//	- manyID - id of the child relation to destroy
	//	- canDelete - optional callback to run for each the value to delete
	//
	//	RETURNS:
	//	- error - error for the destroy parameters, or another operation is in progress
	//	- ErrorOneIDEmpty
	//	- ErrorManyIDEmpty
	//	- ErrorOneIDDestroying // if the one realtion model is already destroying
	//	- ErrorKeyDestroying // if the key was aleady called to be destroyed
	//
	// destroy an item from the Many relation
	DestroyOneOfManyByID(oneID string, associatedID string, canDelete OneToManyTreeRemove) error
}

// Check to see that the reserved keyword is for the associatedID
func hasResevedKeys(keyValues datatypes.KeyValues) bool {
	for key, _ := range keyValues {
		if strings.HasPrefix(key, "_") {
			return true
		}
	}
	return false
}
