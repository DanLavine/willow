package memory

import (
	"context"
	"sync/atomic"

	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
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
	queueName       string
}

func New(ctx context.Context, queue *v1willow.Queue, limiterRuleID string, limiterClient limiterclient.LimiterClient) (*memoryQueue, *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "New")

	limit := new(atomic.Int64)
	limit.Store(*queue.Spec.Properties.MaxItems)

	// need to create an override for the willow Rules in the limiter
	_, err := limiterClient.CreateOverride(ctx, limiterRuleID, &v1.Override{
		Spec: &v1.OverrideSpec{
			DBDefinition: &v1.OverrideDBDefinition{
				GroupByKeyValues: datatypes.KeyValues{
					"_willow_queue_name": datatypes.String(*queue.Spec.DBDefinition.Name),
					"_willow_enqueued":   datatypes.String("true"),
				},
			},
			Properties: &v1.OverrideProperties{
				Limit: queue.Spec.Properties.MaxItems,
			},
		},
	})

	if err != nil {
		logger.Error("Failed to create a Limiter override", zap.Error(err))
		return nil, errors.InternalServerError
	}

	return &memoryQueue{
		limiterClient:   limiterClient,
		configuredLimit: limit,
		queueName:       *queue.Spec.DBDefinition.Name,
	}, nil
}

func (mq *memoryQueue) ConfiguredLimit() int64 {
	return mq.configuredLimit.Load()
}

func (mq *memoryQueue) Update(ctx context.Context, limiterRuleID string, updateReq *v1willow.QueueProperties) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "Update")
	mq.configuredLimit.Store(*updateReq.MaxItems)

	// get the original override id
	overrides, err := mq.limiterClient.QueryOverrides(ctx, limiterRuleID, &queryassociatedaction.AssociatedActionQuery{
		Selection: &queryassociatedaction.Selection{
			KeyValues: queryassociatedaction.SelectionKeyValues{
				"_willow_queue_name": queryassociatedaction.ValueQuery{
					Value:      datatypes.String(mq.queueName),
					Comparison: v1common.Equals,
					TypeRestrictions: v1common.TypeRestrictions{
						MinDataType: datatypes.T_string,
						MaxDataType: datatypes.T_string,
					},
				},
				"_willow_enqueued": queryassociatedaction.ValueQuery{
					Value:      datatypes.String("true"),
					Comparison: v1common.Equals,
					TypeRestrictions: v1common.TypeRestrictions{
						MinDataType: datatypes.T_string,
						MaxDataType: datatypes.T_string,
					},
				},
			},
			MinNumberOfKeyValues: helpers.PointerOf(2),
			MaxNumberOfKeyValues: helpers.PointerOf(2),
		},
	})
	if err != nil {
		panic(err)
	}
	if len(overrides) != 1 {
		panic(overrides)
	}

	// need to create an override for the willow Rules in the limiter
	err = mq.limiterClient.UpdateOverride(ctx, limiterRuleID, overrides[0].State.ID, &v1.OverrideProperties{
		Limit: updateReq.MaxItems,
	})
	if err != nil {
		logger.Error("Failed to update the Limiter override", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}

func (mq *memoryQueue) Destroy(ctx context.Context, limiterRuleID, queueID string) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "Destroy")

	overrides, err := mq.limiterClient.QueryOverrides(ctx, limiterRuleID, &queryassociatedaction.AssociatedActionQuery{
		Selection: &queryassociatedaction.Selection{
			KeyValues: queryassociatedaction.SelectionKeyValues{
				"_willow_queue_name": queryassociatedaction.ValueQuery{
					Value:      datatypes.String(mq.queueName),
					Comparison: v1common.Equals,
					TypeRestrictions: v1common.TypeRestrictions{
						MinDataType: datatypes.T_string,
						MaxDataType: datatypes.T_string,
					},
				},
				"_willow_enqueued": queryassociatedaction.ValueQuery{
					Value:      datatypes.String("true"),
					Comparison: v1common.Equals,
					TypeRestrictions: v1common.TypeRestrictions{
						MinDataType: datatypes.T_string,
						MaxDataType: datatypes.T_string,
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	if len(overrides) != 1 {
		panic(overrides)
	}

	// need to delete the override for the willow Rules in the limiter
	err = mq.limiterClient.DeleteOverride(ctx, limiterRuleID, overrides[0].State.ID)
	if err != nil {
		logger.Error("Failed to delete the Limiter override", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}
