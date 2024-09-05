package queues

import (
	"context"
	"fmt"

	"github.com/DanLavine/willow/internal/willow/brokers/queues/memory"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type Queue interface {
	// Get the configured limit for the queue
	ConfiguredLimit() int64

	// Update the queue parameters
	Update(ctx context.Context, limiterRuleID string, updateRequest *v1willow.QueueProperties) *errors.ServerError

	//	PARAMETERS:
	//	- logger - Logger to record any encountered errors
	//
	//	RETURNS:
	//	- uint64 - Number of total items enqueued including running
	//	- uint64 - Number of items running
	//
	// Get the current statics for the unmber of enqueued and running items
	// GetCurrentStats(logger *zap.Logger) (int64, int64, *errors.ServerError)

	// destroy the queue and any dependent resources
	Destroy(ctx context.Context, limiterRuleID string, queueName string) *errors.ServerError
}

type QueueConstructor interface {
	New(ctx context.Context, queue *v1willow.Queue, limiterRuleID string) (Queue, *errors.ServerError)
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

func (mc *memoryConstrutor) New(ctx context.Context, queue *v1willow.Queue, limiterRuleID string) (Queue, *errors.ServerError) {
	return memory.New(ctx, queue, limiterRuleID, mc.limiterClient)
}
