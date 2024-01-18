package queues

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type QueuesClient interface {
	// Queue operations
	CreateQueue(logger *zap.Logger, queueCreate *v1willow.QueueCreate) *errors.ServerError
	GetQueue(logger *zap.Logger, queueName string, channelQuery *v1common.AssociatedQuery) (*v1willow.Queue, *errors.ServerError)
	ListQueues(logger *zap.Logger) (v1willow.Queues, *errors.ServerError)
	UpdateQueue(logger *zap.Logger, queueName string, queueUpdate *v1willow.QueueUpdate) *errors.ServerError
	DeleteQueue(logger *zap.Logger, queueName string) *errors.ServerError

	// Channel operations
	Enqueue(logger *zap.Logger, queueName string, enqueueItem *v1willow.EnqueueQueueItem) *errors.ServerError
	Dequeue(logger *zap.Logger, cancelContext context.Context, queueName string, dequeueQuery datatypes.AssociatedKeyValuesQuery) (*v1willow.DequeueQueueItem, func(), func(), *errors.ServerError)
	DeleteChannel(logger *zap.Logger, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError

	// Item operations
	Ack(logger *zap.Logger, queueName string, ack *v1willow.ACK) *errors.ServerError
	Heartbeat(logger *zap.Logger, queueName string, heartbeat *v1willow.Heartbeat) *errors.ServerError
}
