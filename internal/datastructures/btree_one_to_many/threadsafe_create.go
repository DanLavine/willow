package btreeonetomany

import (
	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Create a new item in the OneToMany tree
//
//	PARAMETERS:
//	- oneID - the realtionship all created items belong to.
//	- associatedID - associatedID for the KeyValues
//	- keyValues - keyValues that define the object saved in relation to the oneID
//	- onCreate - callback to run when the item is newly created
//
//	RETURNS:
//	- error - error withe he parameters or the tree is already being destroyed
//	        1. datatypes.KeyValuesErr // error with the keyValues
//	        2. ErrorOneIDEmpty
//	        3. ErrorManyIDEmpty
//	        4. ErrorKeyValuesEmpty
//	        5. ErrorManyKeyValuesContainsReservedKeys
//	        6. ErrorOnCreateNil
//	        7. ErrorManyIDAlreadyExists
//	        8. ErrorManyKeyValuesAlreadyExist
//	        9. ErrorOneIDDestroying
//	        10. ErrorManyIDDestroying
func (otm *threadsafeOneToManyTree) CreateWithID(oneID string, manyID string, manyKeyValues datatypes.KeyValues, onCreate OneToManyTreeOnCreate) error {
	// parameter checks
	if oneID == "" {
		return ErrorOneIDEmpty
	}
	if manyID == "" {
		return ErrorManyIDEmpty
	}
	if err := manyKeyValues.Validate(); err != nil {
		return err
	}
	if hasResevedKeys(manyKeyValues) {
		return ErrorManyKeyValuesContainsReservedKeys
	}
	if onCreate == nil {
		return ErrorOnCreateNil
	}

	// create or find the one relationship
	var findErr error
	bTreeFindOne := func(item any) {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		associatedCreate := func() any {
			item := onCreate()
			if item == nil {
				return nil
			}

			return &oneToManyItem{
				value:         item,
				oneID:         oneID,
				manyID:        manyID,
				manyKeyValues: manyKeyValues,
			}
		}

		// create the many relationship
		err := threadsafeValuesNode.associaedTree.CreateWithID(manyID, manyKeyValues, associatedCreate)
		switch err {
		case nil:
			findErr = nil
		case btreeassociated.ErrorAssociatedIDAlreadyExists:
			findErr = ErrorManyIDAlreadyExists
		case btreeassociated.ErrorKeyValuesAlreadyExists:
			// associated id or the key values already exist. Copy those errors to the client
			findErr = ErrorManyKeyValuesAlreadyExist
		case btreeassociated.ErrorTreeItemDestroying:
			// at this point, should only find that the key is being destroyed
			findErr = ErrorManyIDDestroying
		default:
			panic(err)
		}
	}

	bTreeCreateOne := func() any {
		newNode := newThreadsafeValuesNode()
		bTreeFindOne(newNode)

		return newNode
	}

	if err := otm.oneKeys.CreateOrFind(datatypes.String(oneID), bTreeCreateOne, bTreeFindOne); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			return ErrorOneIDDestroying
		default:
			panic(err)
		}
	}

	return findErr
}
