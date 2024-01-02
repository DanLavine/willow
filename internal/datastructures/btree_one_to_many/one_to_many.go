package btreeonetomany

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
)

// top level tree
type threadsafeOneToManyTree struct {
	// Each value here is a *threadsafeVauesNode
	oneKeys btree.BTree
}

// contains all the relations to the one tree
type threadsafeValuesNode struct {
	associaedTree btreeassociated.BTreeAssociated
}

func (threadsafeValuesNode *threadsafeValuesNode) Delete() error {
	return nil
}

func NewThreadSafe() *threadsafeOneToManyTree {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &threadsafeOneToManyTree{
		oneKeys: tree,
	}
}

func newThreadsafeValuesNode() *threadsafeValuesNode {
	return &threadsafeValuesNode{
		associaedTree: btreeassociated.NewThreadSafe(),
	}
}
