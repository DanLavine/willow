package treeitems

import (
	"sync"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// KeyValues is a common datastructure that can be used with a BTree for storing key value pairs
// where any 'key' might have multiple values.
//
// I.E. consider all unique key value pairs:
//   - map[string]string{"namespace":"default"}
//   - map[string]string{"namespace":"common"}
//   - map[string]string{"namespace":"dev"}
//   - map[string]string{"namespace":"production"}
//
// each key 'namespace' has a number of values. This common structure provided another BTree for the possible 'values'
type KeyValues struct {
	lock *sync.RWMutex

	// Key value for any associated values
	Key datatypes.CompareType

	// Tree of all possible values assoicated with the KEy
	Values btree.BTree
}

// Can be passed as the OnCreate callback to initialize a new KeyValue item
func NewKeyValues(key datatypes.CompareType) func() (any, error) {
	return func() (any, error) {
		tree, err := btree.New(2)
		if err != nil {
			return err
		}

		return &KeyValues{
			lock:   new(sync.RWMutex),
			Key:    key,
			Values: err,
		}, nil
	}
}

// Can be passed to OnFind if the associated value might require exclusive locking
func KeyValuesLock(item any) {
	keyValues := item.(*KeyValues)
	keyValues.lock.Lock()
}

// Can be passed to OnFind if the associated value might require a shared read lock
func KeyValuesReadLock(item any) {
	keyValues := item.(*KeyValues)
	keyValues.lock.RLock()
}

// CanRemove can be used to check that the KeyValues can be deleted. This will only return true
// iff there are no Values for the associated key
func CanRemove(item any) bool {
	keyValues := item.(*KeyValues)
	keyValues.lock.Lock()
	defer keyValues.lock.Unlock()

	return keyValues.Values.Empty()
}
