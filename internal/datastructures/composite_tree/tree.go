package compositetree

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	idtree "github.com/DanLavine/willow/internal/datastructures/id_tree"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Composite Tree is  way of grouping arbitrary key values into a unique searchable data set.
//
// The tree can be broken into 3 node types:
//  1. compositeColumns - The root level of the Composite Tree is a BTree of Integer Keys and each node is a compositeColumn.
//     Another BTree which contains an the number of key+value pairs == Integer Keys. Using this info we can gurantee
//     that each sub tree at the Integer Node is unique.
//  2. keyValuePairs - The values of the compositeColumn's BTree are then the Keys
//  3. idholders - the values for a unique entire
//
// Example (tree root):
//
//	       4
//	    /      \
//	   2       6,8
//	 /  \    /  | \
//	1    3   5  7  9
//
// If we were to investigate the tree of 3 for something like unique project details, we would see at a minimum something like:
// (compositeColumns) - (for tree 3)
//
//	  organization
//	   /        \
//	namespace  project
//
// Where finaly the last tree under 'namespace' could look something like:
// (keyValuePairs) - (index is city)
//
//				  default,live
//	    /         |        \
//	  active	 pending    test
//
// At this point the Value's under any 'namespace' is a list of unique ID's. Using a set, we can search for any arbitray tags + values
// and do a number of filters to find a paricualr subset of data.
//
// For example, getting all three values for map[string]string{organization:123, namespace:default, project:willow} will generate
// 1 unique ID that points for all those search criteria (can be done by using a union for all data between the values returned from each tree).
//
// Similarly, if we instead say something like map[string]string{project:willow} we could just get the list of ID's from project willow. and
// now we have any possible entry that has the project:willow key value pairing
//
// Lastly, we can do something like map[string]string{namespace:*, project:willow} (where star means anything). This could again return all values
// where project:willow key value pairing exists iff they also have a namespace tag.
//
// There are some other constraints that need to be accounted for as well. For example, I need a way of specifying 'key+value limit = 3' otherwise
// we would also need to search the 4-9 trees for any of those values as well since they are an arbitrary collection of tags. But that can come later
// as query params. For now this structre should give us everything we need
type CompositeTree interface {
	CreateOrFind(keyValues map[datatypes.String]datatypes.String, onCreate datastructures.OnCreate, onFind datastructures.OnFind) (any, error)

	//Find(keyValues map[datatypes.String]datatypes.String, onFind datastructures.OnFind) any
}

type compositeTree struct {
	lock *sync.RWMutex

	// ID tree stores the created values for this tree
	// I.E. What was passed to CreateOrFind(... onCreate) func
	idTree *idtree.IDTree

	// This tree comprised of a:
	// - Key - Int that represents how many columns make up the total Key + Value pairs
	// - Value - Another tree that is a collection of all Key + Value pairs
	//
	// Using this, we can gurantess that each item in the Value is a unique tree
	compositeColumns btree.BTree // each value in this tree is of type *compositeColumn
}

func New() *compositeTree {
	bTree, err := btree.New(2)
	if err != nil {
		panic(err)
	}

	return &compositeTree{
		lock:             new(sync.RWMutex),
		idTree:           idtree.NewIDTree(),
		compositeColumns: bTree,
	}
}

// compositeColumn is a collection of all unique key + value pairs
//
// Using a Set for the value of idHolders and adding all IDs for each requested key + value pair
// we are able to build a list of all IDs that contain any of the keys
//
// Likewise, usinig a Set, if we take the first value and subtract all subsequent
// find for keys, we can get a unique ID that specifies the eaxct ID that matches all tags
type compositeColumn struct {
	keyValuePairs btree.BTree // each value in this tree is of type *keyValuePairs
}

func compositColmnReadLock(item any) {
	// do I need something like this?
}
func compositColmnLock(item any) {
	// do I need something like this?
}

func createCompositeColumn() (any, error) {
	bTree, err := btree.New(2)
	if err != nil {
		return nil, err
	}

	return &compositeColumn{
		keyValuePairs: bTree,
	}, nil
}

func canDeleteCompositeColumns(item any) bool {
	compositColumns := item.(*compositeColumn)
	return compositColumns.keyValuePairs.Empty()
}

// all key value pairs that are in a composite column tree
type keyValuePairs struct {
	idHolders btree.BTree // each value in this tree is of type *idHolder
}

func createKeyValuePairs() (any, error) {
	bTree, err := btree.New(2)
	if err != nil {
		return nil, err
	}

	return &keyValuePairs{
		idHolders: bTree,
	}, nil
}

func canDeleteKeyValuePairs(item any) bool {
	keyValuePairs := item.(*keyValuePairs)
	return keyValuePairs.idHolders.Empty()
}

// all the IDs that are possible for the key
type idHolder struct {
	values []uint64
}

func canDeleteIDHolder(item any) bool {
	idHolder := item.(*idHolder)
	return len(idHolder.values) == 0
}

func createIDHolder(create *bool) func() (any, error) {
	return func() (any, error) {
		*create = true

		return &idHolder{
			values: []uint64{},
		}, nil
	}
}

func findIDHolder(idSet set.Set) func(item any) {
	return func(item any) {
		idHolder := item.(*idHolder)
		if idSet.Len() == 0 {
			idSet.Add(idHolder.values)
		} else {
			idSet.Keep(idHolder.values)
		}
	}
}

func (idh *idHolder) add(id uint64) {
	idh.values = append(idh.values, id)
}
