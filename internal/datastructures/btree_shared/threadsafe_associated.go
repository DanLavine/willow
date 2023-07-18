package btreeshared

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	idtree "github.com/DanLavine/willow/internal/datastructures/id_tree"
)

type threadsafeAssociatedTree struct {
	// the actual saved values in the tree
	ids *idtree.IDTree

	// each value here is a threadSafeValuesNode
	keys btree.BTree
}

type threadsafeValuesNode struct {
	// each value in here is an threadsafeIDNode
	values btree.BTree
}

type threadsafeIDNode struct {
	// each value in here is an threadsafeIDNode
	// which are saved in the threadsafefAssociatedTree.ids
	ids [][]uint64
}

func NewThreadSafe() *threadsafeAssociatedTree {
	btree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &threadsafeAssociatedTree{
		ids:  idtree.NewIDTree(),
		keys: btree,
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
		ids: [][]uint64{},
	}
}
