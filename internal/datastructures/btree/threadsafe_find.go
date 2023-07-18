package btree

import (
	"fmt"

	"github.com/DanLavine/willow/internal/datastructures"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// PARAMS:
// - key - key to use when comparing to other possible items
// - onFind - method to call if the key is found
//
// RETURNS:
// - error - any errors encontered. I.E. key is not valid
//
// Find the item in the Tree and run the `OnFind(...)` function for the saved value. Will not be called if the
// key cannot be found
func (btree *threadSafeBTree) Find(key datatypes.EncapsulatedData, onFind datastructures.OnFind) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if onFind == nil {
		return fmt.Errorf("onFind is nil, but a value is required")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	if btree.root != nil {
		btree.root.find(key, onFind)
	}

	return nil
}

func (bn *threadSafeBNode) find(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	for index := 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if !keyValue.key.Less(key) {
			// this is an exact match for the key
			if !key.Less(keyValue.key) {
				onFind(keyValue.value)
			}

			// key must be on a child, so recurse down
			if bn.numberOfChildren != 0 {
				bn.children[index].find(key, onFind)
			}
		}
	}

	// if there are children, the value must be on the greater than path
	if bn.numberOfChildren != 0 {
		bn.children[bn.numberOfChildren-1].find(key, onFind)
	}
}

// Find any values in the BTree whos values are not equal to the provided key
func (btree *threadSafeBTree) FindNotEqual(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	if btree.root != nil {
		btree.root.findNotEqual(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findNotEqual(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// always attempt a recurse on the lest than nodes to find all values
		if bn.numberOfChildren != 0 {
			bn.children[index].findNotEqual(key, onFind)
		}

		if !keyValue.key.Less(key) && !key.Less(keyValue.key) {
			// nothing to do here, this is the value we don't want
		} else {
			onFind(keyValue.value)
		}
	}

	// also need to check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findNotEqual(key, onFind)
	}
}

// Find any values in the BTree whos values are less than the provided key
func (btree *threadSafeBTree) FindLessThan(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	if btree.root != nil {
		btree.root.findLessThan(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findLessThan(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// always attempt a recurse on the lest than nodes to find all values
		if bn.numberOfChildren != 0 {
			bn.children[index].findLessThan(key, onFind)
		}

		if keyValue.key.Less(key) {
			onFind(keyValue.value)
		} else {
			return
		}
	}

	// if we hot here, also need to check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findLessThan(key, onFind)
	}
}

// Find any values in the BTree whos values are less than or Equal to the provided key
func (btree *threadSafeBTree) FindLessThanOrEqual(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	if btree.root != nil {
		btree.root.findLessThanOrEqual(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findLessThanOrEqual(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// always attempt a recurse on the lest than nodes to find all values
		if bn.numberOfChildren != 0 {
			bn.children[index].findLessThanOrEqual(key, onFind)
		}

		if keyValue.key.Less(key) {
			onFind(keyValue.value)
		} else {
			// add the equals value for the key
			if !key.Less(keyValue.key) {
				onFind(keyValue.value)
			}
			return
		}
	}

	// if we hot here, also need to check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findLessThanOrEqual(key, onFind)
	}
}

// Find any values in the BTree whos values are greater than the provided key
func (btree *threadSafeBTree) FindGreaterThan(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	if btree.root != nil {
		btree.root.findGreaterThan(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThan(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = bn.numberOfValues; index > 0; index-- {
		keyValue := bn.keyValues[index-1]

		// can attempt on the greater than values
		if bn.numberOfChildren != 0 {
			bn.children[index].findGreaterThan(key, onFind)
		}

		if key.Less(keyValue.key) {
			onFind(keyValue.value)
		} else {
			return
		}
	}

	// if we hot here, also need to check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findGreaterThan(key, onFind)
	}
}

// Find any values in the BTree whos values are greater or equal than the provided key
func (btree *threadSafeBTree) FindGreaterThanOrEqual(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
	if err := key.Validate(); err != nil {
		return fmt.Errorf("key is invalid: %w", err)
	}

	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	btree.lock.RLock()
	defer btree.lock.RUnlock()

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	if btree.root != nil {
		btree.root.findGreaterThanOrEqual(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThanOrEqual(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = bn.numberOfValues; index > 0; index-- {
		keyValue := bn.keyValues[index-1]

		// can attempt on the greater than values
		if bn.numberOfChildren != 0 {
			bn.children[index].findGreaterThan(key, onFind)
		}

		if key.Less(keyValue.key) {
			onFind(keyValue.value)
		} else {
			// this is the equal key
			if keyValue.key.Less(key) {
				onFind(keyValue.value)
			}
			return
		}
	}

	// if we hot here, also need to check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findGreaterThan(key, onFind)
	}
}
