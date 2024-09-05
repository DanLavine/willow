package btree

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
)

var (
	// on create callback errors
	ErrorOnCreateNil = fmt.Errorf("onCreate callback cannot be nil")

	// on find callback errors
	ErrorOnFindNil = fmt.Errorf("onFind callback cannot be nil")

	// general pagination errors
	ErrorsOnIterateNil = fmt.Errorf("onIterate callback cannot be nil")

	// error when creating an item by a specific key and the key already exists
	ErrorKeyAlreadyExists = fmt.Errorf("key already exists")

	// error when destroying a key that is already in the process of being destroyed
	ErrorKeyDestroying = fmt.Errorf("key is being destroyed")

	// error when a tree is already destroying
	ErrorTreeDestroying = fmt.Errorf("tree is being destroyed")
)

// Callback to create an item in the btree
//
//	RETURNS:
//	- any - the item to save in the tree. If this is nil, the item will not be saved in the tree
type BTreeOnCreate func() any

// Callback to for when an item is found in the tree
//
//	PARAMETERS:
//	- any - the item saved in the tree
type BTreeOnFind func(item any)

// Callback to check that an item can actually be removed from a tree
//
//	PARAMS:
//	- key - key for the item saved
//	- item - the original item saved to the bTree
//
//	RETURNS:
//	- bool - if true, will remove the item item from the tree. If this ever returns false when doing bulk deletions then
//	         the deletion operations will be halted. Any objects previously destroyed before the error will not be restored
type BTreeRemove func(key datatypes.EncapsulatedValue, item any) bool

// Callback when iterating over tree values
//
//	PARAMS:
//	- key - key for the item saved
//	- item - the original item saved to the bTree
//
//	RETURNS:
//	- bool - if true, will continue iterating thrrough the tree. If this ever returns false then the pagination is halted
type BTreeIterate func(key datatypes.EncapsulatedValue, item any) bool

// BTree is a generic 2-3-4 BTree implementation.
// See https://www.geeksforgeeks.org/2-3-4-tree/ for details on what a 2-3-4 tree is
type BTree interface {
	// Inserts the keyValue into the tree if the key does not already exist. If the Key does exist
	// then an error will be returned and 'onCreate()' will not be called
	//
	//	PARAMS:
	//	- key - key to use when comparing to other possible values
	//	- onCreate - callback function to create the value if it does not exist. If the create callback was to fail, its up to the callback to perform any cleanup operations and return nil. In this case nothing will be saved to the tree
	//
	//	RETURNS:
	//	- error - any errors encontered. I.E. key is not valid
	//	          1. datatypes.EncapsulatedValueErr // error with the key
	//	          2. ErrorOnCreateNil
	//	          3. ErrorKeyAlreadyExists
	//	          4. ErrorKeyDestroying
	//	          5. ErrorTreeDestroying
	Create(key datatypes.EncapsulatedValue, onCreate BTreeOnCreate) error

	// Inserts the keyValue into the tree if the key does not already exist:
	// In this case, the keyValue returned from 'onCreate()' will be saved in the tree iff the return keyValue != nil.
	//
	// If the key already exists:
	// the key's associated keyValue will be passed to the 'onFind' callback.
	//
	//	PARAMS:
	//	- key - key to use when comparing to other possible values
	//	- onCreate - callback function to create the value if it does not exist. If the create callback was to fail, its up to the callback to perform any cleanup operations and return nil. In this case nothing will be saved to the tree
	//	- onFind - method to call if the key already exists
	//
	//	RETURNS:
	//	- error - any errors encontered. I.E. key is not valid
	//	          1. datatypes.EncapsulatedValueErr // error with the key
	//	          2. ErrorOnCreateNil
	//	          3. ErrorOnCreateNil
	//	          4. ErrorKeyDestroying
	//	          5. ErrorTreeDestroying
	CreateOrFind(key datatypes.EncapsulatedValue, onCreate BTreeOnCreate, onFind BTreeOnFind) error

	//  PARAMS:
	//  - key - key to use when comparing to other possible items
	//  - onFind - method to call if the key is found
	//
	//  RETURNS:
	//  - error - any errors encontered. I.E. key is not valid
	//
	// Find the item in the Tree and run the `OnFind(...)` function for the saved value. Will not be called if the
	// key cannot be found
	Find(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, callback BTreeIterate) error

	// Iterare over all items that don't equal the provided key
	FindNotEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, callback BTreeIterate) error

	// Iterare over all items where the key's are less than the provided value
	FindLessThan(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, callback BTreeIterate) error

	// Iterare over all items where the key's are less than or equal to the provided value
	FindLessThanOrEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, callback BTreeIterate) error

	// Iterare over all items where the key's are greater than the provided value
	FindGreaterThan(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, callback BTreeIterate) error

	// Iterare over all items where the key's are greater than or equal to the provided value
	FindGreaterThanOrEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, callback BTreeIterate) error

	// Delete a keyValue from the BTree for the given Key. If the key does not exist
	// in the tree then this performs a no-op. If the key is nil, then Delete will panic
	//
	//	PARAMS:
	//	- key - the key for the item to delete from the tree
	//	- canDelete - optional function to check if an item can be deleted. If this is nil, the item will be deleted from the tree
	//
	//	RETURNS:
	//	- error - errors encountered when validating params or destroy in progress
	//	        1. datatypes.EncapsulatedValueErr // error with the key
	//	        2. ErrorKeyDestroying
	//	        3. ErrorTreeDestroying
	Delete(key datatypes.EncapsulatedValue, canDelete BTreeRemove) error

	// Destroy is used to remove a key and set the Tree's configuration so any other calls to the key
	// returns an error.
	//
	//	RETURNS:
	//	- error - any errors when trying to destroy he key
	//	        - 1. datatypes.EncapsulatedValueErr // error with the key
	//	        - 2. ErrorKeyDestroying
	//	        - 3. ErrorTreeDestroying
	Destroy(key datatypes.EncapsulatedValue, canDelete BTreeRemove) error

	// DestroyAll values in the BTree. This operation can be a bit slow and make all other calls to the tree
	// return with the `ErrorTreeDestroying` error. It is intended to be called when the tree is no longer needed
	// and will be discarded
	//
	//	RETURNS:
	//	- error - any errors when trying to destroy he key
	//	        - 1. ErrorTreeDestroying
	DestroyAll(canDelete BTreeRemove) error

	// Check to see if there are any key value pairs in the Btree
	//
	// RETURNS:
	// - bool - true if there are any items in the Btree
	Empty() bool
}
