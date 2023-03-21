package queues

import (
	"context"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type Queue interface {
	// OnFind for BTree lookups
	OnFind()

	// GoAsync execution runner that handles shutdown operations
	Execute(ctx context.Context) error

	// Enqueue an item onto the queue
	Enqueue(enqueueItem *v1.EnqueueItem) *v1.Error
}
