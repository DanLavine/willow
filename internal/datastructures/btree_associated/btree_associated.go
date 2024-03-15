package btreeassociated

import (
	"fmt"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

var (
	// on create callback errors
	ErrorOnCreateNil = fmt.Errorf("onCreate callback cannot be nil")

	// on find callback errors
	ErrorOnFindNil = fmt.Errorf("onFind callback cannot be nil")

	// general iteration errors
	ErrorsOnIterateNil = fmt.Errorf("onIterate callback cannot be nil")

	// errors with the assoicated id
	ErrorAssociatedIDEmpty = fmt.Errorf("associatedID cannot be the empty string")

	// errors when creating an item
	ErrorAssociatedIDAlreadyExists = fmt.Errorf("associatedID already exist")
	ErrorKeyValuesAlreadyExists    = fmt.Errorf("keyValues already exist with an item")
	ErrorKeyValuesHasAssociatedID  = fmt.Errorf("keyValues cannot contain a Key with the _associated_id reserved key word")

	// error when destroying a key that is already in the process of being destroyed
	ErrorTreeItemDestroying = fmt.Errorf("tree item is already being destroyed")

	// error when a tree is already destroying
	ErrorTreeDestroying = fmt.Errorf("tree is being destroyed")
)

// Callback that is used to actaully create the item in the tree
//
//	RETURNS:
//	- any - the item to save in the tree. If this is nil, the item will not be saved in the tree
type BTreeAssociatedOnCreate func() any

// Callback that is used when an item is found in the tree
//
//	PARAMETERS:
//	- associatedKeyValues - detailed information about the item saved in the tree, including a referance to the item itself
//
//	RETURNS:
//	- any - the item to save in the tree. If this is nil, the item will not be saved in the tree
type BTreeAssociatedOnFind func(associatedKeyValues AssociatedKeyValues)

// Callback to check that an item can actually be removed from a tree
//
//	PARAMS:
//	- key - key for the item saved
//	- item - the original item saved to the bTree
//
//	RETURNS:
//	- bool - if true, will remove the item item from the tree. If this ever returns false when doing bulk deletions then
//	         the deletion operations will be halted. Any objects previously destroyed before the error will not be restored
type BTreeAssociatedRemove func(associatedKeyValues AssociatedKeyValues) bool

// Callback when iterating over tree values
//
//	PARAMS:
//	- key - key for the item saved
//	- item - the original item saved to the bTree
//
//	RETURNS:
//	- bool - if true, will continue iterating thrrough the tree. If this ever returns false then the pagination is halted
type BTreeAssociatedIterate func(associatedKeyValues AssociatedKeyValues) bool

// bTreeAssociated is used to grouping arbitrary key values into a unique searchable data set.
//
// The tree is structured with these rules:
// 1. The root of the tree is all the 'keys' which are searchable via a string data type.
// 2. each KeyNode for a 'key' is another bTree with possible 'values'.
// 3. each ValueNode is a struct contains a slice of [][]string, where the 1st index is how many keys comprise a unique index.
// I.E:
//
//	0 -> 1, so will only have 1 id ever.
//	1 -> 2, so will need to look for another tag (intersecction) to know if the pair share an ID. OR a union to find all IDs that have a particular KeyValue pair
//	2 -> 3, etc
//	...
//
// Example (tree root):
//
//	       d
//	    /      \
//	   c       f,h
//	 /  \    /  |  \
//	a    b  e   g   i
//
// If we were to investigate the tree of 'a' it could just be a list of all words that begin with 'a'.
//
//			apple
//	   /    \
//	ant     axe
//
// at this point, any value will have a structure of:
//
//	type unique_ids struct {
//	  ids: [][]string
//	}
//
// So if we wanted to just find the map[string]EncapsulatedData{'a':'ant'}
// This would correspond to ids[0][0] -> general ant info it the kye value pair (or whatever is saved)
//
// but if we wanted something like large ant colonies, we could find something like map[string]EncapsulatedData{'a':'ant', 'colony size':'large'}
// with these, we could do an intersection of ant.ids[1] n 'colony size'.ids[1] -> would output all intersected ids for large an colonies
// if that is how we decided to store the data.
//
// With this flexibility, we can find any type of unique groupings, and query a generalized key value data set
type BTreeAssociated interface {
	// // Find an item in the assoociation tree
	// Find(keyValues datatypes.KeyValues, onFind BTreeAssociatedOnFind) error

	// // Find an item in the assoociation tree by the assocaited id
	// FindByAssociatedID(associatedID string, onFind BTreeAssociatedOnFind) error

	// Create an item in the associated tree.
	// Returns an error if
	// 1. if the KeyValues already exists when creating the associated item in the tree
	Create(keyValues datatypes.KeyValues, onCreate BTreeAssociatedOnCreate) (string, error)

	// CreateWithID an item in the associated tree.
	// Returns an error if
	// 1. the associatedID already exists
	// 2. if the KeyValues already exists when creating the associated item in the tree
	CreateWithID(associatedID string, keyValues datatypes.KeyValues, onCreate BTreeAssociatedOnCreate) error

	// Create or Find an item in the association tree
	CreateOrFind(keyValues datatypes.KeyValues, onCreate BTreeAssociatedOnCreate, onFind BTreeAssociatedOnFind) (string, error)

	// Delete an item in the association tree
	Delete(keyValues datatypes.KeyValues, canDelete BTreeAssociatedRemove) error

	// Delete an item in the association tree by the AssociatedID
	DeleteByAssociatedID(associatedID string, canDelete BTreeAssociatedRemove) error

	// Destroy an item in the association tree by the AssociatedID. When destroying the item, any other calls for
	// the AssociatedID will return an error. On Queries, the item is ignored
	DestroyByAssociatedID(associatedID string, canDelete BTreeAssociatedRemove) error

	// DestroyTeee can be used to delete all entries in the tree. This makes it so any other call to the tree reeturns an error
	// that the tree is being destroy. This is to be used when a tree is no longer relevant, and any callers are going
	// to remove their referance to the object
	DestroyTree(canDelete BTreeAssociatedRemove) error

	// delete an number of items that match a particular query
	//DeleteByQuery(query datatypes.AssociatedKeyValuesQuery, canDelete datastructures.CanDelete) error

	// MatchKeys can be used to find any  any permutation of the KeyValues with items saved in the tree. This can be done via a query, but the
	// query can be huge and slow. This is an optimization of finding any entries that mach all possible tag combinations
	// of key values provided.
	MatchAction(matchActionQuery *querymatchaction.MatchActionQuery, onQueryPagination BTreeAssociatedIterate) error

	// Serch for any number of items in the assoociation tree
	QueryAction(query *queryassociatedaction.AssociatedActionQuery, onIterate BTreeAssociatedIterate) error
}
