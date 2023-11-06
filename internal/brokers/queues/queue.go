package queues

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

// QueueManager manges all queue operations such as Create/Update/Delete, etc
// It also is a shared resource that can be used by any other structs needing to find a Queue
// and thier associated readers
type QueueManager interface {
	// Create a new queue. performs a no-op if the queue already exists
	//
	// ARGS:
	// - logger - standard zap logger
	// - create - parameters to create a queue
	//
	// RETURNs:
	// - api.Error - any errors encountered when creating the queue
	Create(logger *zap.Logger, create *v1willow.Create) *api.Error

	// Find a particular queue
	//
	// ARGS:
	// - logger - standard zap logger
	// - queue - name of the queue to find
	//
	// RETURNS:
	// - Queue - queue if the queue is defined
	// - v1willow.Error - any errors encountered when finding the queue, or an error that it does not exist
	Find(logger *zap.Logger, queue string) (Queue, *api.Error)

	// Get Metrics for all Queues
	//
	// RETURNS:
	// - v1willow.MettricsResponse - data for all queues and their tag groups
	Metrics() *v1willow.MetricsResponse
}

// Managed queue defines all the managment functions that queue needs for its lifecycle.
// These functions are not useful for the client's intaction with the queue
//
//go:generate mockgen -destination=queuesfakes/managed_queue_mock.go -package=queuesfakes github.com/DanLavine/willow/internal/brokers/queues ManagedQueue
type ManagedQueue interface {
	// Inlude all the Queue functions as well
	Queue

	// GoAsync execution runner that handles shutdown operations
	Execute(ctx context.Context) error
}

// Queue is any type that can be part of a ManagedQueue, but mainly defines the functions
// available to any clients.
type Queue interface {
	// Enqueue an item onto the queue
	Enqueue(logger *zap.Logger, enqueueItem *v1willow.EnqueueItemRequest) *api.Error

	// Dequeue an item from the queue
	Dequeue(logger *zap.Logger, cancelContext context.Context, selection datatypes.AssociatedKeyValuesQuery) (*v1willow.DequeueItemResponse, func(), func(), *api.Error)

	// ACK a message
	ACK(logger *zap.Logger, ackItem *v1willow.ACK) *api.Error

	// Get Metrics info
	Metrics() *v1willow.QueueMetricsResponse
}
