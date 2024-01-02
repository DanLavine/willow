package btreeonetomany

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

//	RETURNS:
//	- error - error encountered with the
//	        1. datatypes.KeyValuesErr // error with the keyValues
//	        2. ErrorOneIDEmpty
//	        3. ErrorsOnIterateNil
//	        4. ErrorOneIDDestroying
//
// Match all permutations of the KeyValues for a OneRelation
func (tree *threadsafeOneToManyTree) MatchPermutations(oneID string, matchKeyValues datatypes.KeyValues, onIterate OneToManyTreeIterate) error {
	// parameter checks
	if oneID == "" {
		return ErrorOneIDEmpty
	}
	if err := matchKeyValues.Validate(); err != nil {
		return err
	}
	if onIterate == nil {
		return ErrorOnIterateNil
	}

	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		return onIterate(item.Value().(OneToManyItem))
	}

	bTreeFindOne := func(item any) {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		if err := threadsafeValuesNode.associaedTree.MatchPermutations(matchKeyValues, bTreeAssociatedOnIterate); err != nil {
			panic(err)
		}
	}

	if err := tree.oneKeys.Find(datatypes.String(oneID), bTreeFindOne); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			// the tree is already destroying
			return ErrorOneIDDestroying
		default:
			panic(err)
		}
	}

	return nil
}
