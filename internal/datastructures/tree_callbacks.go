package datastructures

import "github.com/DanLavine/willow/pkg/models/datatypes"

// Common way of interacting with any of the data structures

// When using any of the tree's we have these special callbacks that can be used
// to ensure only 1 value is being operated on at a time in a tree (like a db transaction).
//
// Example: Our tree is used for transactional locks. When creating the first
// item in a tree representing a named lock, the `OnCreate()` callback will be called once
// since the tree has an exclusive lock during the create process. So the n+ requests
// can call the optional `OnFind(...)` callback to try and grab a lock on the item being retrieved.
//
// Having `OnFind(...)` grab a lock is necessary if a delete request comes in and try to delete
// the named lock. We want to ensure that no threads have access to that named lock when it is fond
// as part of the delete operation. Then any requests that could also be waiting for the same
// named lock will create a new one. If we don't have a way of performing the special callback lock operations
// we can run into a concurrency issue (its not a race condition in code, but its a logical race condition).
//
// If we have no locks, and we have n threads all trying to grab the same item + any number of threads trying
// to delete the same item. In this case, the tree will just return the item, but if that item's key is the
// only lookup mechanic, we can essentially 'duplicate' the item in the tree from the caller's perspective
// since the they will have an in memory copy of the last item, not the shared item by key that they
// are all expecting

// Iterate is called for each value in a tree
type Iterate func(key datatypes.CompareType, item any)

// Callback for creating a value and inserting them into trees
type OnCreate func() (any, error)

// Callback for calling a function when a value is found in a tree
//
// TODO: Think about returning a different type of value. One where its a reference to the item in the tree and can call A Delete() function.
// would make the associated tree much easier to work with.
type OnFind func(item any)

// Callback to check that an item can actually be removed from a tree
type CanDelete func(item any) bool
