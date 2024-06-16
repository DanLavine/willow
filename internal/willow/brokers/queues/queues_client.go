package queues

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type QueuesClient interface {
	// Queue operations
	CreateQueue(ctx context.Context, queueCreate *v1willow.Queue) *errors.ServerError
	GetQueue(ctx context.Context, queueName string) (*v1willow.Queue, *errors.ServerError)
	ListQueues(ctx context.Context) (v1willow.Queues, *errors.ServerError)
	UpdateQueue(ctx context.Context, queueName string, queueUpdate *v1willow.QueueProperties) *errors.ServerError
	DeleteQueue(ctx context.Context, queueName string) *errors.ServerError

	// Channel operations
	QueryChannels(ctx context.Context, queueName string, query *queryassociatedaction.AssociatedActionQuery) (v1willow.Channels, *errors.ServerError)
	DeleteChannel(ctx context.Context, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError

	// Item operations
	Enqueue(ctx context.Context, queueName string, enqueueItem *v1willow.Item) *errors.ServerError
	Dequeue(cancelContext context.Context, queueName string, dequeueQuery *queryassociatedaction.AssociatedActionQuery) (*v1willow.Item, func(), func(), *errors.ServerError)
	Ack(ctx context.Context, queueName string, ack *v1willow.ACK) *errors.ServerError
	Heartbeat(ctx context.Context, queueName string, heartbeat *v1willow.Heartbeat) *errors.ServerError
}
