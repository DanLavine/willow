package constructor

import (
	"context"
	"fmt"

	"github.com/DanLavine/willow/internal/willow/brokers/queue_channels/memory"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

//go:generate mockgen -imports v1willow="github.com/DanLavine/willow/pkg/models/api/willow/v1" -destination=constructorfakes/queue_channel_mock.go -package=constructorfakes github.com/DanLavine/willow/internal/willow/brokers/queue_channels/constructor QueueChannel
type QueueChannel interface {
	// callback used to know if on ACK, the channel can be deleted
	Delete() bool

	// used for API calls when deleting a channel, to force the deletion
	ForceDelete(logger *zap.Logger)

	Execute(ctx context.Context) error

	Enqueue(logger *zap.Logger, enqueueItem *v1willow.EnqueueQueueItem) *errors.ServerError

	Dequeue() <-chan func(logger *zap.Logger) (*v1willow.DequeueQueueItem, func(), func())

	ACK(zapLogger *zap.Logger, ack *v1willow.ACK) (bool, *errors.ServerError)

	Heartbeat(zapLogger *zap.Logger, heartbeat *v1willow.Heartbeat) *errors.ServerError
}

//go:generate mockgen -destination=constructorfakes/queue_channel_constructor_mock.go -package=constructorfakes github.com/DanLavine/willow/internal/willow/brokers/queue_channels/constructor QueueChannelsConstrutor
type QueueChannelsConstrutor interface {
	New(deleteCallback func(), queueName string, channelKeyValues datatypes.KeyValues) QueueChannel
}

func NewQueueChannelConstructor(constructorType string, limiterClient limiterclient.LimiterClient) (QueueChannelsConstrutor, error) {
	switch constructorType {
	case "memory":
		return &memoryConstructor{
			limiterClient: limiterClient,
		}, nil
	default:
		return nil, fmt.Errorf("unkown constructor type")
	}
}

// memory constructor
type memoryConstructor struct {
	limiterClient limiterclient.LimiterClient
}

func (mc *memoryConstructor) New(deleteCallback func(), queueName string, channelKeyValues datatypes.KeyValues) QueueChannel {
	return memory.New(mc.limiterClient, deleteCallback, queueName, channelKeyValues)
}
