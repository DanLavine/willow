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
//	- manyID - manyID for the manyKeyValues
//	- manyKeyValues - keyValues that define the object saved in relation to the oneID
//	- onCreate - callback to run when the item is newly created
//
//	RETURNS:
//	- error - error withe he parameters or the tree is already being destroyed
//	        1. datatypes.KeyValuesErr // error with the keyValues
//	        2. ErrorOneIDEmpty
//	        3. ErrorManyIDEmpty
//	        4. ErrorManyKeyValuesContainsReservedKeys
//	        5. ErrorOnCreateNil
//	        6. ErrorManyIDAlreadyExists
//	        7. ErrorManyKeyValuesAlreadyExist
//	        8. ErrorOneIDDestroying
//	        9. ErrorManyIDDestroying
func (otm *threadsafeOneToManyTree) CreateWithID(oneID string, manyID string, manyKeyValues datatypes.KeyValues, onCreate OneToManyTreeOnCreate) error {
	// parameter checks
	if oneID == "" {
		return ErrorOneIDEmpty
	}
	if manyID == "" {
		return ErrorManyIDEmpty
	}
	if err := manyKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
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

// Create or Find a new item in the OneToMany tree
//
//	PARAMETERS:
//	- oneID - the realtionship all created items belong to.
//	- manyKeyValues - keyValues that define the object saved in relation to the oneID
//	- onCreate - callback to run when the item is newly created
//	- onFind- callback to run when the item was found
//
//	RETURNS:
//	- error - error withe he parameters or the tree is already being destroyed
//	        1. datatypes.KeyValuesErr // error with the keyValues
//	        2. ErrorOneIDEmpty
//	        3. ErrorManyKeyValuesContainsReservedKeys
//	        4. ErrorOnCreateNil
//	        5. ErrorOnFindNil
//	        7. ErrorManyKeyValuesAlreadyExist
//	        8. ErrorOneIDDestroying
//	        9. ErrorManyKeyValuesDestroying
func (otm *threadsafeOneToManyTree) CreateOrFind(oneID string, manyKeyValues datatypes.KeyValues, onCreate OneToManyTreeOnCreate, onFind OneToManyTreeOnFind) (string, error) {
	// parameter checks
	if oneID == "" {
		return "", ErrorOneIDEmpty
	}
	if err := manyKeyValues.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil {
		return "", err
	}
	if hasResevedKeys(manyKeyValues) {
		return "", ErrorManyKeyValuesContainsReservedKeys
	}
	if onCreate == nil {
		return "", ErrorOnCreateNil
	}
	if onFind == nil {
		return "", ErrorOnFindNil
	}

	// created or found the one relationship
	var findErr error
	var createOrFindID string
	bTreeFindOne := func(item any) {
		threadsafeValuesNode := item.(*threadsafeValuesNode)

		//create the many relationship
		var newOneToManyItem *oneToManyItem
		created := false
		associatedCreate := func() any {
			// create the original item passed in
			item := onCreate()
			if item == nil {
				return nil
			}

			created = true
			newOneToManyItem = &oneToManyItem{
				value:         item,
				oneID:         oneID,
				manyKeyValues: manyKeyValues,
			}

			return newOneToManyItem
		}

		// found the many relationship
		associatedFind := func(item btreeassociated.AssociatedKeyValues) {
			onFind(item.Value().(OneToManyItem))
		}

		// create the many relationship
		var err error
		createOrFindID, err = threadsafeValuesNode.associaedTree.CreateOrFind(manyKeyValues, associatedCreate, associatedFind)
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

		// record the many ID if it was newly created. This is still thread safe as it is in the BTree's create method
		if created {
			newOneToManyItem.manyID = createOrFindID
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
			return "", ErrorOneIDDestroying
		default:
			panic(err)
		}
	}

	return createOrFindID, findErr
}
