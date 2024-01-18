package queues

import (
	"fmt"

	"github.com/DanLavine/willow/internal/willow/brokers/queues/memory"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"go.uber.org/zap"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type Queue interface {
	// Get the configured limit for the queue
	ConfiguredLimit() int64

	// Update the queue parameters
	Update(logger *zap.Logger, queueName string, updateRequest *v1willow.QueueUpdate) *errors.ServerError

	//	PARAMETERS:
	//	- logger - Logger to record any encountered errors
	//
	//	RETURNS:
	//	- uint64 - Number of total items enqueued including running
	//	- uint64 - Number of items running
	//
	// Get the current statics for the unmber of enqueued and running items
	GetCurrentStats(logger *zap.Logger) (int64, int64, *errors.ServerError)

	// destroy the queue and any dependent resources
	Destroy(logger *zap.Logger, queueName string) *errors.ServerError
}

type QueueConstructor interface {
	New(logger *zap.Logger, queue *v1willow.QueueCreate) (Queue, *errors.ServerError)
}

func NewQueueConstructor(constructorType string, limiterClient limiterclient.LimiterClient) (QueueConstructor, error) {
	switch constructorType {
	case "memory":
		return &memoryConstrutor{
			limiterClient: limiterClient,
		}, nil
	default:
		return nil, fmt.Errorf("unknown constructor type")
	}
}

type memoryConstrutor struct {
	limiterClient limiterclient.LimiterClient
}

func (mc *memoryConstrutor) New(logger *zap.Logger, queue *v1willow.QueueCreate) (Queue, *errors.ServerError) {
	return memory.New(logger, queue, mc.limiterClient)
}
