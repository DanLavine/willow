package btreeassociated

import (
	"sync"
	"sync/atomic"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/idgenerator"
)

type threadsafeAssociatedTree struct {
	// these keep track to know if a tree is being deleted entierly.
	// the DeleteAll(...) operation waits until no other requests are running
	// before processing a delete
	readWriteWG *sync.WaitGroup
	destroying  *atomic.Bool

	// associated ids
	// each value here is an *AssociatedKeyValues
	associatedIDs btree.BTree

	// generator for the IDs saved to the id tree
	idGenerator idgenerator.UniqueIDs

	// KeyValues that were provided for creating an item in the tree
	// each value here is a *threadSafeValuesNode.
	keys btree.BTree
}

type threadsafeValuesNode struct {
	// each value in here is an threadsafeIDNode
	values btree.BTree
}

type threadsafeIDNode struct {
	lock *sync.RWMutex

	// this is needed for determining on create/delete race conditions if a particular key value pair is in the process of
	// being created when the delete thread ran.
	creating *atomic.Int64

	// each value in here is a threadsafeIDNode
	// which are saved in the threadsafefAssociatedTree.ids
	ids [][]string
}

func NewThreadSafe() *threadsafeAssociatedTree {
	associatedIDs, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	keys, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &threadsafeAssociatedTree{
		readWriteWG:   new(sync.WaitGroup),
		destroying:    new(atomic.Bool),
		associatedIDs: associatedIDs,
		idGenerator:   idgenerator.UUID(),
		keys:          keys,
	}
}

func newValuesNode() *threadsafeValuesNode {
	btree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &threadsafeValuesNode{
		values: btree,
	}
}

func newIDNode() *threadsafeIDNode {
	return &threadsafeIDNode{
		lock:     new(sync.RWMutex),
		creating: new(atomic.Int64),
		ids:      [][]string{},
	}
}
