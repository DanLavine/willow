package btreeonetomany

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

//	PARAMETERS:
//	- oneID - id of the entire relation tree to destroy from
//	- manyKeyValues - keyValues of the child relation to destroy
//	- canDelete - optional callback to run for each the value to delete
//
//	RETURNS:
//	- error - error for the destroy parameters, or another operation is in progress
//	        1. datatypes.KeyValuesErr // error with the keyValues
//	        2. ErrorOneIDEmpty
//	        3. ErrorOneIDDestroying // if the one realtion model is already destroying
//	        4. ErrorManyKeyValuesDestroying // if the key values item is already destroying
//
// Destroy an item from the Many Relations
func (tree *threadsafeOneToManyTree) DeleteOneOfManyByKeyValues(oneID string, manyKeyValues datatypes.KeyValues, canDelete OneToManyTreeRemove) error {
	// parameter checks
	if oneID == "" {
		return ErrorOneIDEmpty
	}
	if err := manyKeyValues.Validate(); err != nil {
		return err
	}

	bTreeAssociatedDeleteMany := func(item btreeassociated.AssociatedKeyValues) bool {
		deleted := true
		if canDelete != nil {
			deleted = canDelete(item.Value().(OneToManyItem))
		}

		return deleted
	}

	var destroyErr error
	bTreeOnFindOne := func(item any) {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		if err := threadsafeValuesNode.associaedTree.Delete(manyKeyValues, bTreeAssociatedDeleteMany); err != nil {
			switch err {
			case btreeassociated.ErrorTreeItemDestroying:
				// key is already destroying
				destroyErr = ErrorManyKeyValuesDestroying
			default:
				panic(err)
			}
		}
	}

	if err := tree.oneKeys.Find(datatypes.String(oneID), bTreeOnFindOne); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			// the tree is already destroying
			return ErrorOneIDDestroying
		default:
			panic(err)
		}
	}

	return destroyErr
}
