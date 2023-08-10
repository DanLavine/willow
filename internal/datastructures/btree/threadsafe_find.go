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
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.find(key, onFind)
	} else {
		btree.lock.RUnlock()
	}

	return nil
}

func (bn *threadSafeBNode) find(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		if !keyValue.key.Less(key) {
			// this is an exact match for the key
			if !key.Less(keyValue.key) {
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
		bn.children[index].lock.RLock()

		// at this point, we know that all children have appropriate locks
		bn.lock.RUnlock()

		// recurse down to child where the value exists
		bn.children[index].find(key, onFind)
	} else {
		// no more children, so unlock this node
		bn.lock.RUnlock()
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findNotEqual(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findNotEqual(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	// iterate through all the current values
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// always attempt to lock a child
		if bn.numberOfChildren != 0 {
			bn.children[index].lock.RLock()
		}

		if !keyValue.key.Less(key) && !key.Less(keyValue.key) {
			// nothing to do here, this is the value we don't want
		} else {
			onFind(keyValue.value)
		}
	}

	// also need to check the greater than side
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children already have locks now as well
	bn.lock.RUnlock()

	for i := 0; i < bn.numberOfChildren; i++ {
		bn.children[i].findNotEqual(key, onFind)
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findNotEqualMatchType(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findNotEqualMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	startIndex := -1
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is less than what we are looking for, we can go to the next index
		if keyValue.key.LessType(key) {
			continue
		}

		// the key in the tree is greater than the type we are looking for. So just check the less than child and return
		if key.LessType(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			break
		}

		// at this point, the key types match
		if startIndex == -1 {
			startIndex = index
		}

		// always attempt a recurse on the lest than nodes to find all values
		if bn.numberOfChildren != 0 {
			bn.children[index].lock.RLock()
		}

		if !keyValue.key.LessValue(key) && !key.LessValue(keyValue.key) {
			// nothing to do here, this is the value we don't want
		} else {
			onFind(keyValue.value)
		}
	}

	// lock the last child index hat we need to iterate down
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// recurse through the greater than side
			bn.children[index].findNotEqualMatchType(key, onFind)
		} else {
			// need to recurse to all potential children from the start index
			for i := startIndex; i <= index; i++ {
				bn.children[i].findNotEqualMatchType(key, onFind)
			}
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findLessThan(key, findCallback)
	} else {
		btree.lock.RUnlock()
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

			// will attempt a recurse on the lest than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
			}
		} else {
			break
		}
	}

	// always lock the index  we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	// one last attempt to look at the last less than values
	if bn.numberOfChildren != 0 {
		for i := 0; i <= index; i++ {
			bn.children[i].findLessThan(key, onFind)
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findLessThanMatchType(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findLessThanMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	startIndex := -1
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key in the tree is less than what we are looking for, we can go to the next index
		if keyValue.key.LessType(key) {
			continue
		}

		// the key in the tree is greater than the type we are looking for. So just check the less than child and return
		if key.LessType(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			break
		}

		// at this point, the key types match
		if startIndex == -1 {
			startIndex = index
		}

		if keyValue.key.LessValue(key) {
			onFind(keyValue.value)

			// always attempt a recurse on the less than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
			}
		} else {
			break
		}
	}

	// always lock the index  we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// key must be grater than all values we checked, must be on the greater than side
			bn.children[index].findLessThanMatchType(key, onFind)
		} else {
			// need to recurse to all potential children from the start index
			for i := startIndex; i <= index; i++ {
				bn.children[i].findLessThanMatchType(key, onFind)
			}
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findLessThanOrEqual(key, findCallback)
	} else {
		btree.lock.RUnlock()
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
				bn.children[index].lock.RLock()
			}
		} else {
			// add the equals value for the key
			if !key.Less(keyValue.key) {
				onFind(keyValue.value)
			}

			break
		}
	}

	// always lock the index  we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	// one last attempt to look at the last less than or greater than values
	if bn.numberOfChildren != 0 {
		for i := 0; i <= index; i++ {
			bn.children[i].findLessThanOrEqual(key, onFind)
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findLessThanOrEqualMatchType(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findLessThanOrEqualMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	startIndex := -1
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// key's type in the tree is less than what we are looking for so move to the next index
		if keyValue.key.LessType(key) {
			continue
		}

		// key's type in the tree is greater than what we are looking for so move to the next index
		if key.LessType(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			break
		}

		// at this point, we know the keys type match
		if startIndex == -1 {
			startIndex = index
		}

		if keyValue.key.LessValue(key) {
			onFind(keyValue.value)

			// always attempt a recurse on the lest than nodes to find all values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
			}
		} else {
			// add the equals value for the key
			if !key.LessValue(keyValue.key) {
				onFind(keyValue.value)
			}

			break
		}
	}

	// always lock the index  we broke on since we need to check the less than side, or is the last child node (greater than)
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	// one last attempt to look at the last less than values
	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// key must be grater than all values we checked, must be on the greater than side
			bn.children[index].findLessThanOrEqualMatchType(key, onFind)
		} else {
			// need to recurse to all potential children from the start index
			for i := startIndex; i <= index; i++ {
				bn.children[i].findLessThanOrEqualMatchType(key, onFind)
			}
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findGreaterThan(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThan(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	startIndex := -1
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// key in the tree is greater than our provided key
		if key.Less(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			onFind(keyValue.value)

			// can attempt on the less than values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
			}
		}
	}

	// always lock the the last recurse child node
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// check the greater than side
			bn.children[index].findGreaterThan(key, onFind)
		} else {
			// recurse down to additional keys
			for i := startIndex; i <= index; i++ {
				bn.children[i].findGreaterThan(key, onFind)
			}
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findGreaterThanMatchType(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThanMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	startIndex := -1
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key we are looking for, has a type less than in the tree, break and check the child
		if key.LessType(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			break
		}

		// if the key value in the tree is less than the key we care about iterate to the next value
		if keyValue.key.LessType(key) {
			continue
		}

		// at this point, we know the keys type's match

		// key we are looking for is less than the key in the tree
		if key.LessValue(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			onFind(keyValue.value)

			// can attempt the less than values as well
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
			}
		}
	}

	// always lock the the last child node to recurse down
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	// if we got here, also need to check the greater than side
	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// check the greater than side
			bn.children[index].findGreaterThanMatchType(key, onFind)
		} else {
			// recurse down to additional keys
			for i := startIndex; i <= index; i++ {
				bn.children[i].findGreaterThanMatchType(key, onFind)
			}
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findGreaterThanOrEqual(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThanOrEqual(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	startIndex := -1
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// key in the tree is greater than our provided key
		if key.Less(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			onFind(keyValue.value)

			// can attempt on the less than values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
			}
		} else {
			// this is the equal key
			if !keyValue.key.Less(key) {
				onFind(keyValue.value)
			}
		}
	}

	// always lock the the last recurse child node
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// check the greater than side
			bn.children[index].findGreaterThanOrEqual(key, onFind)
		} else {
			// recurse down to additional keys
			for i := startIndex; i <= index; i++ {
				bn.children[i].findGreaterThanOrEqual(key, onFind)
			}
		}
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

	items := []any{}
	findCallback := func(item any) {
		items = append(items, item)
	}

	btree.lock.RLock()
	if btree.root != nil {
		btree.root.lock.RLock()
		btree.lock.RUnlock()
		btree.root.findGreaterThanOrEqualMatchType(key, findCallback)
	} else {
		btree.lock.RUnlock()
	}

	// always attempt the callback, even if nothing was found
	if len(items) != 0 {
		callback(items)
	}

	return nil
}

func (bn *threadSafeBNode) findGreaterThanOrEqualMatchType(key datatypes.EncapsulatedData, onFind datastructures.OnFind) {
	// NOTE; we need to travers from less than to greater than so we don't hit any deadlocks going the reverse way.
	// all traversal needs to be less than to greater than
	startIndex := -1
	var index int
	for index = 0; index < bn.numberOfValues; index++ {
		keyValue := bn.keyValues[index]

		// if the key we are looking for, has a type less than in the tree, break and check the child
		if key.LessType(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			break
		}

		// if the key value in the tree is less than the key we care about iterate to the next value
		if keyValue.key.LessType(key) {
			continue
		}

		if key.LessValue(keyValue.key) {
			if startIndex == -1 {
				startIndex = index
			}

			onFind(keyValue.value)

			// can attempt on the greater than values
			if bn.numberOfChildren != 0 {
				bn.children[index].lock.RLock()
			}
		} else {
			// this is the equal key
			if !keyValue.key.LessValue(key) {
				onFind(keyValue.value)
			}
		}
	}

	// always lock the the last child node to recurse down
	if bn.numberOfChildren != 0 {
		bn.children[index].lock.RLock()
	}

	// Can unlock here, knowing that all the children we care about already have locks now as well
	bn.lock.RUnlock()

	if bn.numberOfChildren != 0 {
		if startIndex == -1 {
			// check the greater than side
			bn.children[index].findGreaterThanOrEqualMatchType(key, onFind)
		} else {
			// recurse down to additional keys
			for i := startIndex; i <= index; i++ {
				bn.children[i].findGreaterThanOrEqualMatchType(key, onFind)
			}
		}
	}
}
