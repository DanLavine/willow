package btreeassociated

import (
	"sync"
	"sync/atomic"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/idgenerator"
)

type threadsafeAssociatedTree struct {
	// the actual saved values in the tree
	// each value in here is a AssociatedKeyValues node
	ids btree.BTree // change this to just be another key _associated_id -> which saves the actual value

	// generator for the IDs saved to the id tree
	idGenerator idgenerator.UniqueIDs

	// each value here is a threadSafeValuesNode
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
	ids, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	keys, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &threadsafeAssociatedTree{
		ids:         ids,
		idGenerator: idgenerator.UUID(),
		keys:        keys,
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
