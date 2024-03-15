package btreeonetomany

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
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
func (tree *threadsafeOneToManyTree) MatchAction(oneID string, matchActionQuery *querymatchaction.MatchActionQuery, onIterate OneToManyTreeIterate) error {
	// parameter checks
	if oneID == "" {
		return ErrorOneIDEmpty
	}
	if err := matchActionQuery.Validate(); err != nil {
		return err
	}
	if onIterate == nil {
		return ErrorOnIterateNil
	}

	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		return onIterate(item.Value().(OneToManyItem))
	}

	bTreeFindOne := func(key datatypes.EncapsulatedValue, item any) bool {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		if err := threadsafeValuesNode.associaedTree.MatchAction(matchActionQuery, bTreeAssociatedOnIterate); err != nil {
			panic(err)
		}

		return true
	}

	if err := tree.oneKeys.Find(datatypes.String(oneID), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, bTreeFindOne); err != nil {
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
