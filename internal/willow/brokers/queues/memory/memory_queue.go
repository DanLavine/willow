package memory

import (
	"context"
	"sync/atomic"

	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

type memoryQueue struct {
	// use for the limiter:
	// 1. create a new queue, need to add the overrides
	// 2. update a queue limits, need to update the overrides
	limiterClient limiterclient.LimiterClient

	// queue details
	configuredLimit *atomic.Int64
}

func New(ctx context.Context, queue *v1willow.QueueCreate, limiterClient limiterclient.LimiterClient) (*memoryQueue, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("New")

	limit := new(atomic.Int64)
	limit.Store(queue.QueueMaxSize)

	// need to create an override for the willow Rules in the limiter
	err := limiterClient.CreateOverride("_willow_queue_enqueued_limits", &v1.Override{
		Name: queue.Name,
		KeyValues: datatypes.KeyValues{
			"_willow_queue_name": datatypes.String(queue.Name),
			"_willow_enqueued":   datatypes.String("true"),
		},
		Limit: queue.QueueMaxSize,
	}, reporting.GetTraceHeaders(ctx))

	if err != nil {
		logger.Error("Failed to create a Limiter override", zap.Error(err))
		return nil, errors.InternalServerError
	}

	return &memoryQueue{
		limiterClient:   limiterClient,
		configuredLimit: limit,
	}, nil
}

func (mq *memoryQueue) ConfiguredLimit() int64 {
	return mq.configuredLimit.Load()
}

func (mq *memoryQueue) Update(ctx context.Context, queueName string, updateReq *v1willow.QueueUpdate) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("Update")

	mq.configuredLimit.Store(updateReq.QueueMaxSize)

	// need to create an override for the willow Rules in the limiter
	err := mq.limiterClient.UpdateOverride("_willow_queue_enqueued_limits", queueName, &v1.OverrideUpdate{
		Limit: updateReq.QueueMaxSize,
	}, reporting.GetTraceHeaders(ctx))
	if err != nil {
		logger.Error("Failed to update the Limiter override", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}

func (mq *memoryQueue) Destroy(ctx context.Context, queueName string) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("Destroy")

	// need to delete the override for the willow Rules in the limiter
	err := mq.limiterClient.DeleteOverride("_willow_queue_enqueued_limits", queueName, reporting.GetTraceHeaders(ctx))
	if err != nil {
		logger.Error("Failed to delete the Limiter override", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}
