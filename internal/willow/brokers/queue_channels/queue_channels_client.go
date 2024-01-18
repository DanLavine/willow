package queuechannels

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type QueueChannelsClient interface {
	// async task execution
	Execute(ctx context.Context) error

	DestroyChannelsForQueue(logger *zap.Logger, queueName string) *errors.ServerError

	DeleteChannel(logger *zap.Logger, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError

	EnqueueQueueItem(zapLogger *zap.Logger, queueName string, enqueueItem *v1willow.EnqueueQueueItem) *errors.ServerError

	DequeueQueueItem(logger *zap.Logger, cancelContext context.Context, queueName string, dequeueQuery datatypes.AssociatedKeyValuesQuery) (*v1willow.DequeueQueueItem, func(), func(), *errors.ServerError)

	ACK(logger *zap.Logger, queueName string, ack *v1willow.ACK) *errors.ServerError

	Heartbeat(logger *zap.Logger, queueName string, heartbeat *v1willow.Heartbeat) *errors.ServerError

	Channels(logger *zap.Logger, queueName string, channelQuery v1common.AssociatedQuery) v1willow.Channels
}
