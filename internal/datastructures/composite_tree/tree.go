package compositetree

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/internal/datastructures/btree"
	idtree "github.com/DanLavine/willow/internal/datastructures/id_tree"
	"github.com/DanLavine/willow/internal/datastructures/set"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

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
