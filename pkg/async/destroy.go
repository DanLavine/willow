package async

import (
	"sync"
	"sync/atomic"
)

type DestroySyncer interface {
	// GuardOperation is used to ensure that a call to Destroy() is blocked untill a corresponding ClearOperation is called
	//
	//	RETURNS:
	//	- bool - true iff the opration is succesfully guraded and guranteed to block Destroy
	GuardOperation() bool

	// ClearOperation is used after a GuardOperation to allow Destroy() to process. This will panic if not called with
	// the GuardOperation 1:1
	ClearOperation()

	// WaitDestroy blocks untill all current operations are done processing
	//
	//	RETURNS:
	//	- bool - true iff the caller processed the destroy operation. In the case of multiple `WaitDestroy` calls at once, only 1 will process
	WaitDestroy() bool

	// ClearDestroy() is used to allow the GuradOperations to start processing again
	ClearDestroy()
}

type DestroySync struct {
	rwlock     *sync.RWMutex
	destroying *atomic.Bool
}

func NewDestroySync() *DestroySync {
	return &DestroySync{
		rwlock:     new(sync.RWMutex),
		destroying: new(atomic.Bool),
	}
}

func (ds *DestroySync) GuardOperation() bool {
	return ds.rwlock.TryRLock()
}

func (ds *DestroySync) ClearOperation() {
	ds.rwlock.RUnlock()
}

func (ds *DestroySync) WaitDestroy() bool {
	if !ds.destroying.CompareAndSwap(false, true) {
		return false
	}

	ds.rwlock.Lock()

	return true
}

func (ds *DestroySync) ClearDestroy() {
	ds.rwlock.Unlock()
	ds.destroying.Store(false)
}
