package queues

import (
	"context"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
	"go.uber.org/zap"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type Queue interface {
	// Enqueue an item onto the queue
	Enqueue(logger *zap.Logger, enqueueItem *v1.EnqueueItem) *v1.Error

	// Requied Function to read lock for tag channels
	TagReadLock()

	// Requied Function to exclusive lock for tag channels
	TagExclusiveLock()

	// Requied Function to lock queue the queue's items. Called as part of Enqueue
	QueueExclusiveLock()
}

// Managed queue defines all the managment functions that queue needs for its lifecycle.
// These functions are not useful for the client's intaction with the queue
type ManagedQueue interface {
	// Inlude all the Queue functions as well
	Queue

	// OnFind for BTree lookups
	OnFind()

	// GoAsync execution runner that handles shutdown operations
	Execute(ctx context.Context) error
}
