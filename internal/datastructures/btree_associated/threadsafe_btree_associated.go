package btreeassociated

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	idtree "github.com/DanLavine/willow/internal/datastructures/id_tree"
)

type threadsafeAssociatedTree struct {
	lock *sync.RWMutex

	// ID tree stores the actual values for this tree
	idTree *idtree.IDTree

	// This tree comprised of a:
	// - Key - Int that represents how many columns make up the total Key + Value pairs
	// - Value - Another tree that is a collection of all Key + Value pairs
	//
	// Using this, we can gurantess that each item in the Value is a unique tree
	groupedKeyValueAssociation btree.BTree // each value in this tree is of type *compositeValue
}

func NewThreadSafe() *threadsafeAssociatedTree {
	bTree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &threadsafeAssociatedTree{
		lock:                       new(sync.RWMutex),
		idTree:                     idtree.NewIDTree(),
		groupedKeyValueAssociation: bTree,
	}
}
