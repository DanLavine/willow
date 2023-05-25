package btreeassociated

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures/btree"
)

// keyValues is a common datastructure that can be used with a BTree for storing key value pairs
// where any 'key' might have multiple values.
//
// I.E. consider all unique key value pairs:
//   - map[string]string{"namespace":"default"}
//   - map[string]string{"namespace":"common"}
//   - map[string]string{"namespace":"dev"}
//   - map[string]string{"namespace":"production"}
//
// each key 'namespace' has a number of values. This common structure provided another BTree for the possible 'values'
type keyValues struct {
	// lock is needed for a particular group when inserting or deleting
	lock *sync.RWMutex

	// bTree of all possible values assoicated with the Key
	values btree.BTree
}

// Can be passed as the OnCreate callback to initialize a new KeyValue item
func newKeyValues() (any, error) {
	tree, err := btree.New(2)
	if err != nil {
		return nil, err
	}

	lock := new(sync.RWMutex)
	lock.Lock()

	return &keyValues{
		lock:   lock,
		values: tree,
	}, nil
}

// Can be passed to OnFind if the associated value might require exclusive locking
func keyValuesLock(item any) {
	keyValues := item.(*keyValues)
	keyValues.lock.Lock()
}

// Can be passed to OnFind if the associated value might require a shared read lock
func keyValuesReadLock(item any) {
	keyValues := item.(*keyValues)
	keyValues.lock.RLock()
}

// CanRemove can be used to check that the keyValues can be deleted. This will only return true
// iff there are no Values for the associated key
func keyValuesCanRemove(item any) bool {
	keyValues := item.(*keyValues)
	return keyValues.values.Empty()
}
