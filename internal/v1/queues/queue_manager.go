package queues

import (
	"context"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type QueueManager interface {
	// Create a new queue
	Create(queueParams v1.Create) *v1.Error

	// Enqueue a new item
	Enqueue(data []byte, updateable bool, queueTags []string) *v1.Error

	// Get the next queue item
	Item(ctx context.Context, matchQuery v1.MatchQuery) (*v1.DequeueMessage, *v1.Error)

	// Process an item
	ACK(id uint64, queueTags []string, passed bool) *v1.Error

	// metrics
	QueueManagerMetrics
}

// metrics for all queues
type QueueManagerMetrics interface {
	// Get metrics for specific queue types
	Metrics(matchRestriction *v1.MatchQuery) *v1.Metrics
}
