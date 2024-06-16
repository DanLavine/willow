package queuechannels

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"

	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type QueueChannelsClient interface {
	// async task execution
	Execute(ctx context.Context) error

	// channel operations
	Channels(ctx context.Context, queueName string, channelQuery *queryassociatedaction.AssociatedActionQuery) v1willow.Channels
	EnqueueQueueItem(ctx context.Context, queueName string, enqueueItem *v1willow.Item) *errors.ServerError
	DequeueQueueItem(ctx context.Context, queueName string, dequeueQuery *queryassociatedaction.AssociatedActionQuery) (*v1willow.Item, func(), func(), *errors.ServerError)
	DestroyChannelsForQueue(ctx context.Context, queueName string) *errors.ServerError
	DeleteChannel(ctx context.Context, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError

	// item operations
	ACK(ctx context.Context, queueName string, ack *v1willow.ACK) *errors.ServerError
	Heartbeat(ctx context.Context, queueName string, heartbeat *v1willow.Heartbeat) *errors.ServerError
}
