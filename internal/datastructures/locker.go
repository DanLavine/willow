package datastructures

import "sync"

// TODO. this will be needed to wrap each "tags" in a lock similar to a DB transaction

// Locker is used to cordinate any dynamic locks that need to use the same
// generic resource. It generate exclusive locks on the fly and removes them
// when no more requests are waiting for the resource to lock
type Locker interface {
	Lock(key any)
	Unlock(key any)
}

type lock struct {
	wg *sync.WaitGroup
}

type locker struct {
	lock *sync.RWMutex
}

func NewLocker() *locker {
	return &locker{}
}
