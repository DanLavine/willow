package compositetree

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures/btree"
)

// compositeKeyValues is a common datastructure that can be used with a BTree for storing key value pairs
// where any 'key' might have multiple values.
//
// I.E. consider all unique key value pairs:
//   - map[string]string{"namespace":"default"}
//   - map[string]string{"namespace":"common"}
//   - map[string]string{"namespace":"dev"}
//   - map[string]string{"namespace":"production"}
//
// each key 'namespace' has a number of values. This common structure provided another BTree for the possible 'values'
type compositeKeyValues struct {
	lock *sync.RWMutex

	// Tree of all possible values assoicated with the Key
	values btree.BTree
}

// Can be passed as the OnCreate callback to initialize a new KeyValue item
func newCompositeKeyValues() (any, error) {
	tree, err := btree.New(2)
	if err != nil {
		return nil, err
	}

	lock := new(sync.RWMutex)
	lock.Lock()

	return &compositeKeyValues{
		lock:   lock,
		values: tree,
	}, nil
}

// Can be passed to OnFind if the associated value might require exclusive locking
func compositeKeyValuesLock(item any) {
	compositeKeyValues := item.(*compositeKeyValues)
	compositeKeyValues.lock.Lock()
}

// Can be passed to OnFind if the associated value might require a shared read lock
func compositeKeyValuesReadLock(item any) {
	compositeKeyValues := item.(*compositeKeyValues)
	compositeKeyValues.lock.RLock()
}

// CanRemove can be used to check that the compositeKeyValues can be deleted. This will only return true
// iff there are no Values for the associated key
func canRemovecompositeKeyValues(item any) bool {
	compositeKeyValues := item.(*compositeKeyValues)
	compositeKeyValues.lock.Lock()
	defer compositeKeyValues.lock.Unlock()

	return compositeKeyValues.values.Empty()
}

// CanRemove can be used to check that the compositeKeyValues can be deleted. This will only return true
// iff there are no Values for the associated key
func cleanFailedCompositeKeyValues(item any) bool {
	compositeKeyValues := item.(*compositeKeyValues)
	return compositeKeyValues.values.Empty()
}
