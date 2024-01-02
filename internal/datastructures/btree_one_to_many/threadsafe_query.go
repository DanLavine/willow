package btreeonetomany

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

//	RETURNS:
//	- error - error encountered with the
//	        1. fmt.Errorf(...) // error with the query
//	        2. ErrorOneIDEmpty
//	        3. ErrorsOnIterateNil
//	        4. ErrorOneIDDestroying
//
// Match all permutations of the KeyValues for a OneRelation
func (tree *threadsafeOneToManyTree) Query(oneID string, query datatypes.AssociatedKeyValuesQuery, onIterate OneToManyTreeIterate) error {
	// parameter checks
	if err := query.Validate(); err != nil {
		return fmt.Errorf("query error: %w", err)
	}
	if oneID == "" {
		return ErrorOneIDEmpty
	}
	if onIterate == nil {
		return ErrorOnIterateNil
	}

	bTreeAssociatedIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		return onIterate(item.Value().(OneToManyItem))
	}

	bTreeFindOne := func(item any) {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		if err := threadsafeValuesNode.associaedTree.Query(query, bTreeAssociatedIterate); err != nil {
			panic(err)
		}
	}

	if err := tree.oneKeys.Find(datatypes.String(oneID), bTreeFindOne); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			// the tree is currently destroying
			return ErrorOneIDDestroying
		default:
			panic(err)
		}
	}

	return nil
}
