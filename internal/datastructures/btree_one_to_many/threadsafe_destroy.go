package btreeonetomany

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Destroy the OneToManyRelationship. As part of this operation we are going to remove all Many relations /w the canDelete
// callback. If that is successful, then the One relation is deleted as well. In the case of removing a 'Rule'. The guards
// for not creating/updating any other 'Overrides' has to exist on the 'Rule'. It is the 'Rules' responsibilty to ensure that
// no other operations are processing at the same time. This is ensure via a 'Delete' in the BTree
//
//	PARAMETERS:
//	- oneID - name of the entire relation tree to destroy
//	- canDelete - optional callback to run for each value in the many relationship
//
//	RETURNS:
//	- error - error for the destroy parameters, or another operation is in progress
//	        - ErrorOneIDEmpty
//	        - ErrorOneIDDestroying
func (tree *threadsafeOneToManyTree) DestroyOne(oneID string, canDelete OneToManyTreeRemove) error {
	// parameter checks
	if oneID == "" {
		return ErrorOneIDEmpty
	}

	deleted := true
	bTreeAssociatedDelete := func(item btreeassociated.AssociatedKeyValues) bool {
		if canDelete != nil {
			deleted = canDelete(item.Value().(OneToManyItem))
		}

		return deleted
	}

	bTreeDeleteOne := func(_ datatypes.EncapsulatedValue, item any) bool {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		if err := threadsafeValuesNode.associaedTree.DestroyTree(bTreeAssociatedDelete); err != nil {
			panic(err)
		}

		return deleted
	}

	if err := tree.oneKeys.Destroy(datatypes.String(oneID), bTreeDeleteOne); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			// this is fine, the key is already destroying
			return ErrorOneIDDestroying
		default:
			panic(err)
		}
	}

	return nil
}

//	PARAMETERS:
//	- oneID - id of the entire relation tree to destroy from
//	- manyID - id of the child relation to destroy
//	- canDelete - optional callback to run for each the value to delete
//
//	RETURNS:
//	- error - error for the destroy parameters, or another operation is in progress
//	- ErrorOneIDEmpty
//	- ErrorManyIDEmpty
//	- ErrorOneIDDestroying // if the one realtion model is already destroying
//	- ErrorKeyDestroying // if the key was aleady called to be destroyed
//
// Destroy an item from the Many Relations
func (tree *threadsafeOneToManyTree) DestroyOneOfManyByID(oneID string, manyID string, canDelete OneToManyTreeRemove) error {
	// parameter checks
	if oneID == "" {
		return ErrorOneIDEmpty
	}
	if manyID == "" {
		return ErrorManyIDEmpty
	}

	bTreeAssociatedDestoyMany := func(item btreeassociated.AssociatedKeyValues) bool {
		deleted := true
		if canDelete != nil {
			deleted = canDelete(item.Value().(OneToManyItem))
		}

		return deleted
	}

	var destroyErr error
	bTreeOnFindOne := func(key datatypes.EncapsulatedValue, item any) bool {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		if err := threadsafeValuesNode.associaedTree.DestroyByAssociatedID(manyID, bTreeAssociatedDestoyMany); err != nil {
			switch err {
			case btreeassociated.ErrorTreeItemDestroying:
				// key is already destroying
				destroyErr = ErrorManyIDDestroying
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
