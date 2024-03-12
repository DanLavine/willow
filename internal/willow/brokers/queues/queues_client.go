package queues

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type QueuesClient interface {
	// Queue operations
	CreateQueue(ctx context.Context, queueCreate *v1willow.QueueCreate) *errors.ServerError
	GetQueue(ctx context.Context, queueName string, channelQuery *v1common.AssociatedQuery) (*v1willow.Queue, *errors.ServerError)
	ListQueues(ctx context.Context) (v1willow.Queues, *errors.ServerError)
	UpdateQueue(ctx context.Context, queueName string, queueUpdate *v1willow.QueueUpdate) *errors.ServerError
	DeleteQueue(ctx context.Context, queueName string) *errors.ServerError

	// Channel operations
	Enqueue(ctx context.Context, queueName string, enqueueItem *v1willow.EnqueueQueueItem) *errors.ServerError
	Dequeue(cancelContext context.Context, queueName string, dequeueQuery datatypes.AssociatedKeyValuesQuery) (*v1willow.DequeueQueueItem, func(), func(), *errors.ServerError)
	DeleteChannel(ctx context.Context, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError

	// Item operations
	Ack(ctx context.Context, queueName string, ack *v1willow.ACK) *errors.ServerError
	Heartbeat(ctx context.Context, queueName string, heartbeat *v1willow.Heartbeat) *errors.ServerError
}
