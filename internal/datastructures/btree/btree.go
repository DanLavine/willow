package btree

import (
	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// BTree is a generic 2-3-4 BTree implementation.
// See https://www.geeksforgeeks.org/2-3-4-tree/ for details on what a 2-3-4 tree is
type BTree interface {
	//  PARAMS:
	//  - key - key to use when comparing to other possible items
	//  - onFind - method to call if the key is found
	//
	//  RETURNS:
	//  - error - any errors encontered. I.E. key is not valid
	//
	// Find the item in the Tree and run the `OnFind(...)` function for the saved value. Will not be called if the
	// key cannot be found
	Find(key datatypes.EncapsulatedData, onFind datastructures.OnFind) error

	// Iterare over all items that don't equal the provided key
	FindNotEqual(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items where the key's are less than the provided value
	FindLessThan(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items where the key's are less than or equal to the provided value
	FindLessThanOrEqual(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items where the key's are greater than the provided value
	FindGreaterThan(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items where the key's are greater than or equal to the provided value
	FindGreaterThanOrEqual(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items that don't equal the provided key, where the key's in the BTree match the search key's data type
	FindNotEqualMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items less than the provided key, where the key's in the BTree match the search key's data type
	FindLessThanMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items less than or equal to the provided key, where the key's in the BTree match the search key's data type
	FindLessThanOrEqualMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items greater than the provided key, where the key's in the BTree match the search key's data type
	FindGreaterThanMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// Iterare over all items greater than or euqal to the provided key, where the key's in the BTree match the search key's data type
	FindGreaterThanOrEqualMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindPagination) error

	// If the provided key does not exist, the onCreate function will be called to initalize a new object.
	// Otherwise the onFind callback will be invoked for the value associated with the key
	//
	// PARAMS:
	// - key - key to use when comparing to other possible items
	// - onCreate - callback function to create the item if it does not exist. If the create callback was to fail, its up
	//              to the callback to perform any cleanup operations and return nil. In this case nothing will be saved to the tree
	// - onFind - method to call if the key already exists
	//
	// RETURNS:
	// - error - any errors encontered. I.E. key is not valid
	CreateOrFind(key datatypes.EncapsulatedData, onCreate datastructures.OnCreate, onFind datastructures.OnFind) error

	// TODO DSL
	// add an update function, similiar to onFind, but uses an exclusive lock. might make things easier for clients to not need locks all over the place
	// Update(key datatypes.EncapsulatedData, onUpdate datastructures.OnFind)

	// Iterate over the tree and for each value found invoke the callback with the node's value
	//
	// PARAMS:
	// - callback - Each value in the BTree will run the provided callback
	//
	// RETURNS:
	// - error - any errors with parameters encontered. I.E. callback is nil
	Iterate(callback datastructures.OnFindPagination) error

	IterateMatchType(dataType datatypes.DataType, callback datastructures.OnFindPagination) error

	// Delete an item in the Tree
	//
	// PARAMS:
	// - key - key to delete
	// - canDelete - optional function to check if a value can be deleted
	//
	// RETURNS:
	// - error - any errors encontered. I.E. key is not valid
	Delete(key datatypes.EncapsulatedData, canDelete datastructures.CanDelete) error

	// Check to see if there are any key value pairs in the Btree
	//
	// RETURNS:
	// - bool - true if there are any items in the Btree
	Empty() bool
}
