package btreeonetomany

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
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
	if err := manyKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
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
	bTreeOnFindOne := func(key datatypes.EncapsulatedValue, item any) bool {
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

		return false
	}

	if err := tree.oneKeys.Find(datatypes.String(oneID), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, bTreeOnFindOne); err != nil {
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
