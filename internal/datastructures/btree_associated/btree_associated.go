package btreeassociated

import (
	"sort"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// KeyValues are the types that are accepted by the actual try operations. Most APIs that I am currently working with have type
// of map[string]datatypes.EnccapsulatedData to ensure that the values of the data over the wire have unique keys and are easier
// to work with for clients. But having this option allows users to define custom data types as well to insert as "reserved" keys
// so we can still make queries against them
type KeyValues map[datatypes.EncapsulatedData]datatypes.EncapsulatedData

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
	// Find an item in the assoociation tree
	// TOOD: Remove this
	Find(keyValuePairs KeyValues, onFind datastructures.OnFind) (string, error)

	// Find an item in the assoociation tree by the assocaited id
	// TOOD: Remove this
	FindByAssociatedID(associatedID string, onFind datastructures.OnFind) error

	// Create an item in the associated tree.
	// Returns an error if
	// 1. if the KeyValues already exists when creating the associated item in the tree
	Create(keyValues KeyValues, onCreate datastructures.OnCreate) (string, error)

	// CreateWithID an item in the associated tree.
	// Returns an error if
	// 1. the associatedID already exists
	// 2. if the KeyValues already exists when creating the associated item in the tree
	CreateWithID(associatedID string, keyValues KeyValues, onCreate datastructures.OnCreate) error

	// Create or Find an item in the association tree
	CreateOrFind(keyValuePairs KeyValues, onCreate datastructures.OnCreate, onFind datastructures.OnFind) (string, error)

	// Serch for any number of items in the assoociation tree
	// todo: should be able to ad _associated_id to the queries as well
	Query(query datatypes.AssociatedKeyValuesQuery, onFindPagination datastructures.OnFindPagination) error

	// Delete an item in the association tree
	Delete(keyValuePairs KeyValues, canDelete datastructures.CanDelete) error

	// Delete an item in the association tree by the AssociatedID
	DeleteByAssociatedID(associatedID string, canDelete datastructures.CanDelete) error

	// delete an number of items that match a particular query
	//DeleteByQuery(query datatypes.AssociatedKeyValuesQuery, canDelete datastructures.CanDelete) error
}

func ConverDatatypesKeyValues(keyValuePairs datatypes.KeyValues) KeyValues {
	keyValues := KeyValues{}

	for key, value := range keyValuePairs {
		keyValues[datatypes.String(key)] = value
	}

	return keyValues
}

// Check to see that the reserved keyword is for the associatedID
func (kv KeyValues) HasAssociatedID() bool {
	_, ok := kv[datatypes.String(ReservedID)]
	return ok
}

// Sort all the Keys of the KeyValues into a sorted order
func (kv KeyValues) SortedKeys() []datatypes.EncapsulatedData {
	keys := []datatypes.EncapsulatedData{}
	for key, _ := range kv {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Less(keys[j])
	})

	return keys
}

// Retrieve the custom data for a particualr KeyValues. Can be used to find what th service added to the customers request
func (kv KeyValues) RetrieveCustomDataTypes() map[datatypes.EncapsulatedData]datatypes.EncapsulatedData {
	keyValuePairs := KeyValues{}

	for key, value := range kv {
		if key.DataType == datatypes.T_custom {
			keyValuePairs[key] = value
		}
	}

	return keyValuePairs
}

// The inverse of ConverDatatypesKeyValues. Can be used to obtain the original values before conversion
func (kv KeyValues) StripAssociatedID() KeyValues {
	keyValuePairs := KeyValues{}

	for key, value := range kv {
		if key.DataType == datatypes.T_string {
			if key.Value.(string) != ReservedID {
				keyValuePairs[key] = value
			}
		} else {
			keyValuePairs[key] = value
		}
	}

	return keyValuePairs
}

func (kv KeyValues) RetrieveStringDataType() datatypes.KeyValues {
	keyValuePairs := datatypes.KeyValues{}

	for key, value := range kv {
		if key.DataType == datatypes.T_string {
			keyValuePairs[key.Value.(string)] = value
		}
	}

	return keyValuePairs
}
