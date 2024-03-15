package btree

import (
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// The interactions are still a bit weird... When using Find(T_any, ...) this is basically an iterate.
//
// But what about FindNotEqual(T_any, ...), should this be an iterate as well, where just the 'key' is strictly checked?

//	PARAMS:
//	- key - key to use when comparing to other possible items. This can not be T_any encapsualted value
//	- typeRestrictions - how to match the particular key
//	- onIterate - method to call if the key is found
//
//	RETURNS:
//	- error - any errors encontered. I.E. key is not valid
//
// Find is used to match a specifica keys, and optinally check for a T_any key as well depending on the type restrictions
func (btree *threadSafeBTree) Find(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) error {
	// parameter checks
	if err := key.Validate(datatypes.MinDataType, datatypes.MaxDataType); err != nil { // why was this not allowing any?
		return &errors.ModelError{Field: "key", Child: err.(*errors.ModelError)}
	}
	if err := typeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "typeRestrictions", Child: err.(*errors.ModelError)}
	}
	if onIterate == nil {
		return ErrorOnFindNil
	}

	// destroy checks
	btree.readWriteWG.Add(1)
	defer btree.readWriteWG.Add(-1)

	if err := btree.checkDestroyingWithKey(key); err != nil {
		return err
	}

	btree.lock.RLock()
	if btree.root != nil {
		switch {
		// When selecting anything other than T_any
		case datatypes.GeneralDataTypes[key.Type]:
			// find just the specific key
			shouldContinue := true

			// we can find the exact type we are searching for
			if !key.Type.LessMatchType(typeRestrictions.MinDataType) && !typeRestrictions.MaxDataType.LessMatchType(key.Type) {
				btree.root.lock.RLock()

				// can unlock the root if we are not going to search for Any
				if typeRestrictions.MaxDataType < datatypes.T_any {
					btree.lock.RUnlock()
				}

				btree.root.findMatchType(key, func(item any) {
					shouldContinue = onIterate(key, item)
				})
			}

			// check for Any
			if shouldContinue && typeRestrictions.MaxDataType == datatypes.T_any {
				// can now unlock this level as we are traversing down again
				btree.root.lock.RLock()
				btree.lock.RUnlock()

				btree.root.findMatchType(datatypes.Any(), func(item any) {
					_ = onIterate(datatypes.Any(), item)
				})
			} else if typeRestrictions.MaxDataType == datatypes.T_any {
				btree.lock.RUnlock()
			}

		// When finding values for T_any, we need to iterate over all the types
		case datatypes.AnyDataType[key.Type]:
			btree.root.lock.RLock()
			btree.lock.RUnlock()

			btree.root.iterate(typeRestrictions, onIterate)
		}
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) findMatchType(key datatypes.EncapsulatedValue, onFind BTreeOnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if !keyValue.key.LessMatchType(key) {
			// this is an exact match for the key
			if !key.LessMatchType(keyValue.key) {
				onFind(keyValue.value)
				bn.lock.RUnlock()
				return
			}

			// key must be on a child, so recurse down
			if bn.numberOfChildren != 0 {
				break
			}
		}
	}

	if bn.numberOfChildren != 0 {
		child := bn.children[index]
		child.lock.RLock()

		// at this point, we know that all children have appropriate locks
		bn.lock.RUnlock()

		// recurse down to child where the value exists
		child.findMatchType(key, onFind)
	} else {
		// no more children, so unlock this node
		bn.lock.RUnlock()
	}
}

func (bn *threadSafeBNode) iterate(typeRestrictions v1common.TypeRestrictions, callback BTreeIterate) bool {
	var i int
	var children []*threadSafeBNode

	for i = 0; i < bn.numberOfValues; i++ {
		// the key in the tree is less then the value we are looking for, iterate to the next value if it exists
		if bn.keyValues[i].key.Type.LessMatchType(typeRestrictions.MinDataType) {
			continue
		}

		// if the max types we are searching for is less than the key in the tree. Try the less than tree and return
		if typeRestrictions.MaxDataType.LessMatchType(bn.keyValues[i].key.Type) {
			break
		}

		// always attempt a recurse on the less than nodes to find all values
		if bn.numberOfChildren != 0 {
			bn.children[i].lock.RLock()
			children = append(children, bn.children[i])
		}

		// run the callback for the current node
		// NOTE: we don't need to check destroing at this point because ther locks are all heald and are read locks.
		// if that is to ever change, this will need a callback to check for items being destroyed
		if !callback(bn.keyValues[i].key, bn.keyValues[i].value) {
			// told to stop iterating, need to run the unlock operations on all the childrent we have locked
			for _, child := range children {
				child.lock.RUnlock()
			}

			bn.lock.RUnlock()
			return false
		}
	}

	// always lock the index we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[i].lock.RLock()
		children = append(children, bn.children[i])
	}

	// have a lock on all the children at this point, can release the lock at this level
	bn.lock.RUnlock()

	// need to recurse to all children
	for i := 0; i < len(children); i++ {
		// need to recurse to all potential children from the start index
		if !children[i].iterate(typeRestrictions, callback) {
			for unlockIndex := i + 1; unlockIndex < len(children); unlockIndex++ {
				bn.children[unlockIndex].lock.RUnlock()
			}

			return false
		}
	}

	return true
}

// Find any values in the BTree whos values are not equal to the provided key
func (btree *threadSafeBTree) FindNotEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) error {
	// parameter checks
	//
	// #DSL TODO: Why did I make this restriction?
	if err := key.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "key", Child: err.(*errors.ModelError)}
	}
	if err := typeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "typeRestrictions", Child: err.(*errors.ModelError)}
	}
	if onIterate == nil {
		return ErrorsOnIterateNil
	}

	// check if the whole tree is destroying
	btree.readWriteWG.Add(1)
	defer btree.readWriteWG.Add(-1)

	if err := btree.checkDestroying(); err != nil {
		return err
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()

		btree.root.iterate(typeRestrictions, func(iterateKey datatypes.EncapsulatedValue, item any) bool {
			if !key.LessMatchType(iterateKey) && !iterateKey.LessMatchType(key) {
				// this is the key we don't want so ignore this one
				return true
			}

			return onIterate(iterateKey, item)
		})
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

// Find any values in the BTree whos values are less than the provided key and respect the type of key
func (btree *threadSafeBTree) FindLessThan(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) error {
	// parameter checks
	if err := key.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "key", Child: err.(*errors.ModelError)}
	}
	if err := typeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "typeRestrictions", Child: err.(*errors.ModelError)}
	}
	if onIterate == nil {
		return ErrorsOnIterateNil
	}

	// destroy checks
	btree.readWriteWG.Add(1)
	defer btree.readWriteWG.Add(-1)

	if err := btree.checkDestroying(); err != nil {
		return err
	}

	btree.lock.RLock()
	if btree.root != nil {
		// find just the specific key
		shouldContinue := true

		// we can find specific types for the provided key and below
		if !key.Type.LessMatchType(typeRestrictions.MinDataType) {
			btree.root.lock.RLock()

			// can unlock the root if we are not going to search for Any
			if typeRestrictions.MaxDataType < datatypes.T_any {
				btree.lock.RUnlock()
			}

			shouldContinue = btree.root.findLessThan(key, typeRestrictions, onIterate)
		}

		// check for any Any
		if shouldContinue && typeRestrictions.MaxDataType == datatypes.T_any {
			// can now unlock this level as we are traversing down again
			btree.root.lock.RLock()
			btree.lock.RUnlock()

			btree.root.findMatchType(datatypes.Any(), func(item any) {
				_ = onIterate(datatypes.Any(), item)
			})
		} else if typeRestrictions.MaxDataType == datatypes.T_any {
			btree.lock.RUnlock()
		}
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) findLessThan(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) bool {
	var index int
	var children []*threadSafeBNode

	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is less than the restriction, just continue
		if keyValue.key.Type.LessMatchType(typeRestrictions.MinDataType) {
			continue
		}

		// the key in the tree is greater than the type we are looking for. So need to check the children
		if key.Type.LessMatchType(keyValue.key.Type) {
			break
		}

		// at this point, the object in the tree is within the range for the type
		if keyValue.key.LessMatchType(key) {
			// always attempt a recurse on the less than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
				children = append(children, bn.children[index])
			}

			if !onIterate(keyValue.key, keyValue.value) {
				// caller wants to stop paginating, need to unlock everything
				if bn.numberOfChildren != 0 {
					for unlockIndex := 0; unlockIndex < len(children); unlockIndex++ {
						children[unlockIndex].lock.RUnlock()
					}
				}

				bn.lock.RUnlock()
				return false
			}
		} else {
			break
		}
	}

	// always add the child for the index we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
		children = append(children, bn.children[index])
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	if len(children) != 0 {
		// need to recurse to all potential
		for i := 0; i < len(children); i++ {
			if !children[i].findLessThan(key, typeRestrictions, onIterate) {
				// need to unlock the rest of the children and return
				for unlockIndex := i + 1; unlockIndex < len(children); unlockIndex++ {
					children[unlockIndex].lock.RUnlock()
				}

				return false
			}
		}
	}

	return true
}

// Find any values in the BTree whos values are less than or Equal to the provided key
func (btree *threadSafeBTree) FindLessThanOrEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) error {
	// parameter checks
	if err := key.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "key", Child: err.(*errors.ModelError)}
	}
	if err := typeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "typeRestrictions", Child: err.(*errors.ModelError)}
	}
	if onIterate == nil {
		return ErrorsOnIterateNil
	}

	// destroy checks
	btree.readWriteWG.Add(1)
	defer btree.readWriteWG.Add(-1)

	if err := btree.checkDestroying(); err != nil {
		return err
	}

	btree.lock.RLock()
	if btree.root != nil {
		// find just the specific key
		shouldContinue := true

		// we can find specific types for the provided key and below
		if !key.Type.LessMatchType(typeRestrictions.MinDataType) {
			btree.root.lock.RLock()

			// can unlock the root if we are not going to search for Any
			if typeRestrictions.MaxDataType < datatypes.T_any {
				btree.lock.RUnlock()
			}

			shouldContinue = btree.root.findLessThanOrEqual(key, typeRestrictions, onIterate)
		}

		// check for any Any
		if shouldContinue && typeRestrictions.MaxDataType == datatypes.T_any {
			// can now unlock this level as we are traversing down again
			btree.root.lock.RLock()
			btree.lock.RUnlock()

			btree.root.findMatchType(datatypes.Any(), func(item any) {
				_ = onIterate(datatypes.Any(), item)
			})
		} else if typeRestrictions.MaxDataType == datatypes.T_any {
			btree.lock.RUnlock()
		}
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) findLessThanOrEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) bool {
	var index int
	var children []*threadSafeBNode

	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is less than the restriction, just continue
		if keyValue.key.Type.LessMatchType(typeRestrictions.MinDataType) {
			continue
		}

		// the key in the tree is greater than the type we are looking for. So need to check the children
		if key.Type.LessMatchType(keyValue.key.Type) {
			break
		}

		// at this point, the object in the tree is within the range for the type
		if keyValue.key.LessMatchType(key) {
			// always attempt a recurse on the less than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
				children = append(children, bn.children[index])
			}

			if !onIterate(keyValue.key, keyValue.value) {
				// caller wants to stop paginating, need to unlock everything
				if bn.numberOfChildren != 0 {
					for unlockIndex := 0; unlockIndex < len(children); unlockIndex++ {
						children[unlockIndex].lock.RUnlock()
					}
				}

				bn.lock.RUnlock()
				return false
			}
		} else {
			// add the equals value for the key
			if !key.LessMatchType(keyValue.key) {
				if !onIterate(keyValue.key, keyValue.value) {
					// caller wants to stop paginating, need to unlock everything
					if bn.numberOfChildren != 0 {
						for rev := 0; rev < len(children); rev++ {
							children[rev].lock.RUnlock()
						}
					}

					bn.lock.RUnlock()
					return false
				}
			}

			break
		}
	}

	// always add the child for the index we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
		children = append(children, bn.children[index])
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	// need to recurse to all children
	for i := 0; i < len(children); i++ {
		if !children[i].findLessThanOrEqual(key, typeRestrictions, onIterate) {
			// need to unlock the rest of the children and return
			for unlockIndex := i + 1; unlockIndex < len(children); unlockIndex++ {
				children[unlockIndex].lock.RUnlock()
			}

			return false
		}
	}

	return true
}

// Find any values in the BTree whos values are greater than the provided key
func (btree *threadSafeBTree) FindGreaterThan(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) error {
	// parameter checks
	if err := key.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "key", Child: err.(*errors.ModelError)}
	}
	if err := typeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "typeRestrictions", Child: err.(*errors.ModelError)}
	}
	if onIterate == nil {
		return ErrorsOnIterateNil
	}

	// destroy checks
	btree.readWriteWG.Add(1)
	defer btree.readWriteWG.Add(-1)

	if err := btree.checkDestroying(); err != nil {
		return err
	}

	btree.lock.RLock()
	if btree.root != nil {
		// we can find specific types for the provided key and above
		if !typeRestrictions.MaxDataType.LessMatchType(key.Type) {
			btree.root.lock.RLock()
			btree.lock.RUnlock()

			btree.root.findGreaterThan(key, typeRestrictions, onIterate)
		}
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThan(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) bool {
	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	var index int
	var children []*threadSafeBNode

	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is greater than the restriction, can break
		if typeRestrictions.MaxDataType.LessMatchType(keyValue.key.Type) {
			break
		}

		// if the key in the tree is less than the key we are searching for
		if keyValue.key.LessMatchType(key) {
			continue
		}

		// at this point, the object in the tree is within the range of the key and restriction
		if key.LessMatchType(keyValue.key) {
			// always attempt a recurse on the less than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
				children = append(children, bn.children[index])
			}

			if !onIterate(keyValue.key, keyValue.value) {
				// caller wants to stop paginating, need to unlock everything
				if bn.numberOfChildren != 0 {
					for unlockIndex := 0; unlockIndex < len(children); unlockIndex++ {
						children[unlockIndex].lock.RUnlock()
					}
				}

				bn.lock.RUnlock()
				return false
			}
		} else {
			// this is the exact key we are looking for, so ignore this case
		}
	}

	// always add the child for the index we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
		children = append(children, bn.children[index])
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	// need to recurse to all potential
	for i := 0; i < len(children); i++ {
		if !children[i].findGreaterThan(key, typeRestrictions, onIterate) {
			// need to unlock the rest of the children and return
			for unlockIndex := i + 1; unlockIndex < len(children); unlockIndex++ {
				children[unlockIndex].lock.RUnlock()
			}

			return false
		}
	}

	return true
}

// Find any values in the BTree whos values are greater than the provided key
func (btree *threadSafeBTree) FindGreaterThanOrEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) error {
	// parameter checks
	if err := key.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "key", Child: err.(*errors.ModelError)}
	}
	if err := typeRestrictions.Validate(); err != nil {
		return &errors.ModelError{Field: "typeRestrictions", Child: err.(*errors.ModelError)}
	}
	if onIterate == nil {
		return ErrorsOnIterateNil
	}

	// destroy checks
	btree.readWriteWG.Add(1)
	defer btree.readWriteWG.Add(-1)

	if err := btree.checkDestroying(); err != nil {
		return err
	}

	btree.lock.RLock()
	if btree.root != nil {
		// we can find specific types for the provided key and above
		if !typeRestrictions.MaxDataType.LessMatchType(key.Type) {
			btree.root.lock.RLock()
			btree.lock.RUnlock()

			btree.root.findGreaterThanOrEqual(key, typeRestrictions, onIterate)
		}
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThanOrEqual(key datatypes.EncapsulatedValue, typeRestrictions v1common.TypeRestrictions, onIterate BTreeIterate) bool {
	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	var index int
	var children []*threadSafeBNode

	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is greater than the restriction, can break
		if typeRestrictions.MaxDataType.LessMatchType(keyValue.key.Type) {
			break
		}

		// if the key in the tree is less than the key we are searching for
		if keyValue.key.LessMatchType(key) {
			continue
		}

		// at this point, the object in the tree is within the range of the key and restriction
		if key.LessMatchType(keyValue.key) {
			// always attempt a recurse on the less than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
				children = append(children, bn.children[index])
			}

			if !onIterate(keyValue.key, keyValue.value) {
				// caller wants to stop paginating, need to unlock everything
				if bn.numberOfChildren != 0 {
					for unlockIndex := 0; unlockIndex < len(children); unlockIndex++ {
						children[unlockIndex].lock.RUnlock()
					}
				}

				bn.lock.RUnlock()
				return false
			}
		} else {
			// this is the exact key we are looking for, so ignore this case
			if !keyValue.key.LessMatchType(key) {
				if !key.LessMatchType(keyValue.key) {
					if !onIterate(keyValue.key, keyValue.value) {
						// caller wants to stop paginating, need to unlock everything
						if bn.numberOfChildren != 0 {
							for unlockIndex := 0; unlockIndex < len(children); unlockIndex++ {
								children[unlockIndex].lock.RUnlock()
							}
						}

						bn.lock.RUnlock()
						return false
					}
				}
			}
		}
	}

	// always add the child for the index we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
		children = append(children, bn.children[index])
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	// need to recurse to all potential
	for i := 0; i < len(children); i++ {
		if !children[i].findGreaterThanOrEqual(key, typeRestrictions, onIterate) {
			// need to unlock the rest of the children and return
			for unlockIndex := i + 1; unlockIndex < len(children); unlockIndex++ {
				children[unlockIndex].lock.RUnlock()
			}

			return false
		}
	}

	return true
}
