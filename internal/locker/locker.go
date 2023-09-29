package locker

import (
	"context"
	"sync"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/api/v1locker"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

type GeneralLocker interface {
	// obtain all the locks that make up a collection
	//
	//  RETURNS:
	//  - func() - callback to release the lock held by the keyValues
	ObtainLocks(serverCtx, clientCtx context.Context, keyValues datatypes.StringMap) func()

	// Find all locks currently held in the tree
	ListLocks() []v1locker.Lock
}

type generalLock struct {
	lockChan               chan struct{}
	keyValues              datatypes.StringMap
	generatedFromKeyValues datatypes.StringMap

	counterLock *sync.Mutex
	counter     int
}

type generalLocker struct {
	locks btreeassociated.BTreeAssociated
}

func NewGeneralLocker(tree btreeassociated.BTreeAssociated) *generalLocker {
	if tree == nil {
		tree = btreeassociated.NewThreadSafe()
	}

	return &generalLocker{
		locks: tree,
	}
}

func (generalLocker *generalLocker) ListLocks() []v1locker.Lock {
	var locks []v1locker.Lock

	onPaginate := func(item any) bool {
		generalLock := item.(*generalLock)
		locks = append(locks, v1locker.NewLock(generalLock.keyValues, generalLock.generatedFromKeyValues))

		return true
	}

	_ = generalLocker.locks.Query(query.Select{}, onPaginate)

	return locks
}

// returns a releas lock function callback
func (generalLocker *generalLocker) ObtainLocks(serverCtx, clientCtx context.Context, keyValues datatypes.StringMap) func() {
	var createOrFound int // 0 for create 1 for find
	var lockChan chan struct{}

	onCreate := func(keyValues, generatedFromKeyValues datatypes.StringMap) func() any {
		return func() any {
			generalLock := &generalLock{
				lockChan:               make(chan struct{}, 1),
				keyValues:              keyValues,
				generatedFromKeyValues: generatedFromKeyValues,
				counterLock:            new(sync.Mutex),
				counter:                1,
			}

			createOrFound = 0
			return generalLock
		}
	}

	// NOTE: it is very important that the all procces in the order that they were found
	onFind := func(item any) {
		generalLock := item.(*generalLock)

		// increment the counter
		generalLock.counterLock.Lock()
		generalLock.counter++
		generalLock.counterLock.Unlock()

		// set the create or found logic
		createOrFound = 1

		// set the channel to wait for
		lockChan = generalLock.lockChan
	}

	// lock every single possible tag combination we might be using
	tagPairs := keyValues.GenerateTagPairs()
	for index, tagPair := range tagPairs {
		_ = generalLocker.locks.CreateOrFind(tagPair, onCreate(tagPair, tagPairs[len(tagPairs)-1]), onFind)

		switch createOrFound {
		case 0:
			// this is the created case just check that the client has not disconnected
			select {
			case <-clientCtx.Done():
				// now we need to try and cleanup any locks we currently have recored a counter on
				generalLocker.freeLocks(tagPairs[:index+1])
				return nil
			case <-serverCtx.Done():
				// now we need to try and cleanup any locks we currently have recored a counter on
				generalLocker.freeLocks(tagPairs[:index+1])
				return nil
			default:
				// nothing to do here
			}
		case 1:
			// this is the found case
			select {
			case <-lockChan:
				// a lock has been freed and we were able to claim it
			case <-clientCtx.Done():
				// now we need to try and cleanup any locks we currently have recored a counter on
				generalLocker.freeLocks(tagPairs[:index+1])
				return nil
			case <-serverCtx.Done():
				// now we need to try and cleanup any locks we currently have recored a counter on
				generalLocker.freeLocks(tagPairs[:index+1])
				return nil
			}
		}
	}

	// at this point, we have obtained the "locks" for all tag groups
	return func() {
		generalLocker.freeAllLocks(keyValues)
	}
}

func (generalLocker *generalLocker) freeAllLocks(keyValues datatypes.StringMap) {
	// attempt an unlock on every single possible tag combination we might be using
	generalLocker.freeLocks(keyValues.GenerateTagPairs())
}

func (generalLocker *generalLocker) freeLocks(keyValues []datatypes.StringMap) {
	canDelete := func(item any) bool {
		generalLock := item.(*generalLock)

		// decrement the counter n a delete
		generalLock.counterLock.Lock()
		generalLock.counter--
		counter := generalLock.counter
		generalLock.counterLock.Unlock()

		// need to always wait for this receiviing something
		if counter == 0 {
			return true
		} else {
			generalLock.lockChan <- struct{}{}
			return false
		}
	}

	// delete or at least signal to other waiting locks that we are freeing the currently held lock
	for _, keyValueGroup := range keyValues {
		generalLocker.locks.Delete(keyValueGroup, canDelete)
	}
}
