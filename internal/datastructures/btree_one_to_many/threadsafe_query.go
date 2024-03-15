package btreeonetomany

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"

	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
)

//	RETURNS:
//	- error - error encountered with the
//	        1. fmt.Errorf(...) // error with the query
//	        2. ErrorOneIDEmpty
//	        3. ErrorsOnIterateNil
//	        4. ErrorOneIDDestroying
//
// Match all permutations of the KeyValues for a OneRelation
func (tree *threadsafeOneToManyTree) QueryAction(oneID string, query *queryassociatedaction.AssociatedActionQuery, onIterate OneToManyTreeIterate) error {
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

	bTreeFindOne := func(key datatypes.EncapsulatedValue, item any) bool {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		if err := threadsafeValuesNode.associaedTree.QueryAction(query, bTreeAssociatedIterate); err != nil {
			panic(err)
		}

		return false
	}

	if err := tree.oneKeys.Find(datatypes.String(oneID), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, bTreeFindOne); err != nil {
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
