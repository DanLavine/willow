package client

import (
	"sync"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

type LockerTracker interface {
	AddReleaseCallback(keyValues datatypes.StringMap, disconnectCallback func())

	HasLocks(keyValues datatypes.StringMap) bool

	ClearLocks(keyValues datatypes.StringMap)

	Disconnect()
}

// I Think this is everything that is needed for a client tracker. Does it make more sense to have 1 of these per
// conn, or just 1, that we pass the conn into each time?
//
// For now, set it up so that I just have 1 per conn

// tracker is used to keep track of messages being processed by a client.
// When a client dequeues an item, if it disconnects, we need to register the item as a failure
type lockerClientTracker struct {
	lock         *sync.RWMutex
	disconnected bool

	callbacks btreeassociated.BTreeAssociated // This can be a collection of all the key value pairings this client has
}

func NewLockerClientTracker() *lockerClientTracker {
	return &lockerClientTracker{
		lock:         new(sync.RWMutex),
		disconnected: false,
		callbacks:    btreeassociated.NewThreadSafe(),
	}
}

// add the disconnect callback to the client tracker if it is not yet disconnected
func (lockerClientTracker *lockerClientTracker) AddReleaseCallback(keyValues datatypes.StringMap, disconnectCallback func()) {
	lockerClientTracker.lock.Lock()
	defer lockerClientTracker.lock.Unlock()

	if !lockerClientTracker.disconnected {
		onCreate := func() any {
			return disconnectCallback
		}

		_ = lockerClientTracker.callbacks.CreateOrFind(keyValues, onCreate, func(item any) { panic("found a locked key") })
	}
}

// HasLock is used to enforce that only the client which obtained the initial lock can release it
func (lockerClientTracker *lockerClientTracker) HasLocks(keyValues datatypes.StringMap) bool {
	lockerClientTracker.lock.Lock()
	defer lockerClientTracker.lock.Unlock()

	found := false
	onFind := func(_ any) {
		found = true
	}

	_ = lockerClientTracker.callbacks.Find(keyValues, onFind)

	return found
}

// RemoveDisconnectCallback is called on the success of a lock being released
func (lockerClientTracker *lockerClientTracker) ClearLocks(keyValues datatypes.StringMap) {
	lockerClientTracker.lock.Lock()
	defer lockerClientTracker.lock.Unlock()

	if !lockerClientTracker.disconnected {
		canDelete := func(item any) bool {
			item.(func())()

			return true
		}

		_ = lockerClientTracker.callbacks.Delete(keyValues, canDelete)
	}
}

// Disconnect releases all locks held by a client if they lose a connection to the locker
func (lockerClientTracker *lockerClientTracker) Disconnect() {
	lockerClientTracker.lock.Lock()
	defer lockerClientTracker.lock.Unlock()

	lockerClientTracker.disconnected = true

	onPagination := func(item any) bool {
		item.(func())()
		return true
	}

	_ = lockerClientTracker.callbacks.Query(query.Select{}, onPagination)
}
