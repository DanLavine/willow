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

// Find any values in the BTree whos values are not equal to the provided key
func (btree *threadSafeBTree) FindNotEqualMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
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
		btree.root.findNotEqualMatchType(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findNotEqualMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is less than what we are looking for, we can go to the next index
		if keyValue.key.LessType(key) {
			continue
		}

		// the key in the tree is greater than the type we are looking for. So just check the less than child and return
		if key.LessType(keyValue.key) {
			if bn.numberOfChildren != 0 {
				bn.children[index].findNotEqualMatchType(key, onFind)
			}

			return
		}

		// at this point, the key types match

		// always attempt a recurse on the lest than nodes to find all values
		if bn.numberOfChildren != 0 {
			bn.children[index].findNotEqualMatchType(key, onFind)
		}

		if !keyValue.key.LessValue(key) && !key.LessValue(keyValue.key) {
			// nothing to do here, this is the value we don't want
		} else {
			onFind(keyValue.value)
		}
	}

	// also need to check the greater than side
	// if this is false, will bail early
	if bn.numberOfChildren != 0 {
		bn.children[index].findNotEqualMatchType(key, onFind)
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

		if keyValue.key.Less(key) {
			onFind(keyValue.value)

			// always attempt a recurse on the lest than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].findLessThan(key, onFind)
			}
		} else {
			break
		}
	}

	// one last attempt to look at the last less than or greater than values
	if bn.numberOfChildren != 0 {
		bn.children[index].findLessThan(key, onFind)
	}
}

// Find any values in the BTree whos values are less than the provided key and respect the type of key
func (btree *threadSafeBTree) FindLessThanMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
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
		btree.root.findLessThanMatchType(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findLessThanMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is less than what we are looking for, we can go to the next index
		if keyValue.key.LessType(key) {
			continue
		}

		// the key in the tree is greater than the type we are looking for. So just check the less than child and return
		if key.LessType(keyValue.key) {
			if bn.numberOfChildren != 0 {
				bn.children[index].findLessThanMatchType(key, onFind)
			}

			return
		}

		// at this point, the types match

		if keyValue.key.LessValue(key) {
			onFind(keyValue.value)

			// always attempt a recurse on the lest than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].findLessThanMatchType(key, onFind)
			}
		} else {
			return
		}
	}

	// one last attempt to look at the greater than values
	if bn.numberOfChildren != 0 {
		bn.children[index].findLessThanMatchType(key, onFind)
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

		// keyValue in the tree is less than what we are looking for
		if keyValue.key.Less(key) {
			onFind(keyValue.value)

			// always attempt a recurse on the lest than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].findLessThanOrEqual(key, onFind)
			}
		} else {
			// add the equals value for the key
			if !key.Less(keyValue.key) {
				onFind(keyValue.value)
			}

			break
		}
	}

	// one last attempt to look at the last less than or greater than values
	if bn.numberOfChildren != 0 {
		bn.children[index].findLessThanOrEqual(key, onFind)
	}
}

// Find any values in the BTree whos values are less than or Equal to the provided key and respects the type of key
func (btree *threadSafeBTree) FindLessThanOrEqualMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
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
		btree.root.findLessThanOrEqualMatchType(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findLessThanOrEqualMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// key's type in the tree is less than what we are looking for so move to the next index
		if keyValue.key.LessType(key) {
			continue
		}

		// key's type in the tree is greater than what we are looking for so move to the next index
		if key.LessType(keyValue.key) {
			if bn.numberOfChildren != 0 {
				bn.children[index].findLessThanOrEqualMatchType(key, onFind)
			}

			return
		}

		// at this point, we know the keys type match

		if keyValue.key.LessValue(key) {
			onFind(keyValue.value)

			// always attempt a recurse on the lest than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].findLessThanOrEqualMatchType(key, onFind)
			}
		} else {
			// add the equals value for the key
			if !key.LessValue(keyValue.key) {
				onFind(keyValue.value)
			}

			break
		}
	}

	// one last attempt to look at the less than or greater than values
	if bn.numberOfChildren != 0 {
		bn.children[index].findLessThanOrEqualMatchType(key, onFind)
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

	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// key in the tree is greater than our provided key
		if key.Less(keyValue.key) {
			onFind(keyValue.value)

			// can attempt on the less than values
			if bn.numberOfChildren != 0 {
				bn.children[index].findGreaterThan(key, onFind)
			}
		}
	}

	// last attempt to always check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findGreaterThan(key, onFind)
	}
}

// Find any values in the BTree whos values are greater than the provided key and respects the type of key
func (btree *threadSafeBTree) FindGreaterThanMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
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
		btree.root.findGreaterThanMatchType(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThanMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int

	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key we are looking for, has a type less than what we are comparing, contine
		if key.LessType(keyValue.key) {
			continue
		}

		// if the key value we are compaing is greater than the key we care about, we found all values and can exit
		if keyValue.key.LessType(key) {
			return
		}

		// at this point, we know the keys type's match

		// key we are looking for is less than the key in the tree
		if key.LessValue(keyValue.key) {
			onFind(keyValue.value)

			// can attempt the less than values as well
			if bn.numberOfChildren != 0 {
				bn.children[index].findGreaterThanMatchType(key, onFind)
			}
		}
	}

	// if we got here, also need to check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findGreaterThanMatchType(key, onFind)
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

	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// key in the tree is greater than our provided key
		if key.Less(keyValue.key) {
			onFind(keyValue.value)

			// can attempt on the less than values
			if bn.numberOfChildren != 0 {
				bn.children[index].findGreaterThanOrEqual(key, onFind)
			}
		} else {
			// this is the equal key
			if !keyValue.key.Less(key) {
				onFind(keyValue.value)
			}
		}
	}

	// last attempt to always check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findGreaterThanOrEqual(key, onFind)
	}
}

// Find any values in the BTree whos values are greater or equal than the provided key and respects the type of key
func (btree *threadSafeBTree) FindGreaterThanOrEqualMatchType(key datatypes.EncapsulatedData, callback datastructures.OnFindSelection) error {
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
		btree.root.findGreaterThanOrEqualMatchType(key, findCallback)
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThanOrEqualMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int

	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key we are looking for, has a type less than what we are comparing, contine
		if key.LessType(keyValue.key) {
			continue
		}

		// if the key value we are compaing is greater than the key we care about, we found all values and can exit
		if keyValue.key.LessType(key) {
			return
		}

		if key.LessValue(keyValue.key) {
			onFind(keyValue.value)

			// can attempt on the greater than values
			if bn.numberOfChildren != 0 {
				bn.children[index].findGreaterThanOrEqualMatchType(key, onFind)
			}
		} else {
			// this is the equal key
			if !keyValue.key.LessValue(key) {
				onFind(keyValue.value)
			}
		}
	}

	// last attempt to always check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].findGreaterThanOrEqualMatchType(key, onFind)
	}
}
