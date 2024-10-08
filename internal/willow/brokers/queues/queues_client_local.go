package queues

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	queuechannels "github.com/DanLavine/willow/internal/willow/brokers/queue_channels"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

func errorMissingQueueName(name string) *errors.ServerError {
	return &errors.ServerError{Message: fmt.Sprintf("failed to find queue '%s' by name", name), StatusCode: http.StatusNotFound}
}

type queueClientLocal struct {
	// queue constructor for creating and managing queues
	queueConstructor QueueConstructor

	// all possible queues a user created
	queues btree.BTree

	// client to interact with the channels in a queue
	queueChannelsClient queuechannels.QueueChannelsClient

	limiterRuleID string
}

func NewLocalQueueClient(queueConstructor QueueConstructor, queueChannelsClient queuechannels.QueueChannelsClient, limiterRuleID string) *queueClientLocal {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &queueClientLocal{
		queueConstructor:    queueConstructor,
		queues:              tree,
		queueChannelsClient: queueChannelsClient,
		limiterRuleID:       limiterRuleID,
	}
}

// Create the main queue and setup the limts on the Limiter service
func (qcl *queueClientLocal) CreateQueue(ctx context.Context, queueCreate *v1willow.Queue) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "CreateQueue")

	var createQueueError *errors.ServerError
	bTreeOnCreate := func() any {
		var queue Queue
		queue, createQueueError = qcl.queueConstructor.New(ctx, queueCreate, qcl.limiterRuleID)
		if createQueueError != nil {
			return nil
		}

		return queue
	}

	if err := qcl.queues.Create(datatypes.String(*queueCreate.Spec.DBDefinition.Name), bTreeOnCreate); err != nil {
		switch err {
		case btree.ErrorKeyAlreadyExists:
			logger.Warn("failed to create queue. Queue already exists by that name")
			return &errors.ServerError{Message: fmt.Sprintf("Queue already exists with name '%s'", *queueCreate.Spec.DBDefinition.Name), StatusCode: http.StatusConflict}
		case btree.ErrorKeyDestroying:
			logger.Warn("failed to create queue. Queue by that name is currenly destroying")
			return &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is currently being destroyed. Try again later", *queueCreate.Spec.DBDefinition.Name), StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to create a queue in the tree for some reason", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return createQueueError
}

func (qcl *queueClientLocal) ListQueues(ctx context.Context) (v1willow.Queues, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "ListQueues")

	var queues v1willow.Queues
	bTreeOnIterate := func(key datatypes.EncapsulatedValue, item any) bool {
		queue := item.(Queue)

		queues = append(queues, &v1willow.Queue{
			Spec: &v1willow.QueueSpec{
				DBDefinition: &v1willow.QueueDBDefinition{
					Name: helpers.PointerOf[string](key.Data.(string)),
				},
				Properties: &v1willow.QueueProperties{
					MaxItems: helpers.PointerOf(queue.ConfiguredLimit()),
				},
			},
			State: &v1willow.QueueState{
				Deleting: false,
			},
		})

		return true
	}

	if err := qcl.queues.Find(datatypes.Any(), v1.TypeRestrictions{MinDataType: datatypes.MinDataType, MaxDataType: datatypes.MaxDataType}, bTreeOnIterate); err != nil {
		logger.Error("error listing queues from tree", zap.Error(err))
		return nil, errors.InternalServerError
	}

	return queues, nil
}

func (qcl *queueClientLocal) GetQueue(ctx context.Context, queueName string) (*v1willow.Queue, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "GetQueue")
	getQueueError := errorMissingQueueName(queueName)

	var queue *v1willow.Queue
	onFind := func(key datatypes.EncapsulatedValue, item any) bool {
		willowQueue := item.(Queue)

		queue = &v1willow.Queue{
			Spec: &v1willow.QueueSpec{
				DBDefinition: &v1willow.QueueDBDefinition{
					Name: helpers.PointerOf[string](queueName),
				},
				Properties: &v1willow.QueueProperties{
					MaxItems: helpers.PointerOf(willowQueue.ConfiguredLimit()),
				},
			},
			State: &v1willow.QueueState{
				Deleting: false,
			},
		}

		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			logger.Warn("failed to get queue. Queue by that name is currenly destroying")
			return nil, &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is currently being destroyed. Refusing to get the queue and channels", queueName), StatusCode: http.StatusConflict}
		default:
			logger.Error("error listing queues from tree", zap.Error(err))
			return nil, errors.InternalServerError
		}
	}

	if queue == nil {
		return nil, getQueueError
	}

	return queue, nil
}

func (qcl *queueClientLocal) UpdateQueue(ctx context.Context, queueName string, queueUpdate *v1willow.QueueProperties) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "UpdateQueue")
	updateQueueError := errorMissingQueueName(queueName)

	bTreeOnFind := func(key datatypes.EncapsulatedValue, item any) bool {
		updateQueueError = item.(Queue).Update(ctx, qcl.limiterRuleID, queueUpdate)
		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, bTreeOnFind); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			logger.Warn("failed to update queue. Queue by that name is currenly destroying")
			return &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is currently being destroyed. Refusing to update the queue", queueName), StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to create a queue in the tree for some reason", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return updateQueueError
}

func (qcl *queueClientLocal) DeleteQueue(ctx context.Context, queueName string) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "DeleteQueue")

	var deleteQueueError *errors.ServerError
	destroyQueue := func(key datatypes.EncapsulatedValue, item any) bool {
		queue := item.(Queue)
		if deleteQueueError = qcl.queueChannelsClient.DestroyChannelsForQueue(ctx, queueName); deleteQueueError == nil {
			if deleteQueueError = queue.Destroy(ctx, qcl.limiterRuleID, queueName); deleteQueueError == nil {
				return true
			}
		}

		return false
	}

	if err := qcl.queues.Destroy(datatypes.String(queueName), destroyQueue); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			logger.Warn("Failed to destroy queue. Queue by that name is already destroying")
			return &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is already being destroyed", queueName), StatusCode: http.StatusConflict}
		default:
			logger.Error("Failed to destroy a queue in the tree for some reason", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return deleteQueueError
}

func (qcl *queueClientLocal) QueryChannels(ctx context.Context, queueName string, query *queryassociatedaction.AssociatedActionQuery) (v1willow.Channels, *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "QueryChannels")
	channels := v1willow.Channels{}

	onIterate := func(key datatypes.EncapsulatedValue, item any) bool {
		channels = qcl.queueChannelsClient.Channels(ctx, queueName, query)
		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onIterate); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			logger.Warn("failed to ack item. Queue by that name is currenly destroying")
			return channels, &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is currently being destroyed. Refusing to acck the item since it is being destroyed too", queueName), StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
			return channels, errorMissingQueueName(queueName)
		}
	}

	return channels, nil
}

func (qcl *queueClientLocal) Enqueue(ctx context.Context, queueName string, enqueueItem *v1willow.Item) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "Enqueue")
	enqueueQueueError := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	onFind := func(key datatypes.EncapsulatedValue, item any) bool {
		enqueueQueueError = qcl.queueChannelsClient.EnqueueQueueItem(ctx, queueName, enqueueItem)
		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
		return errors.InternalServerError
	}

	return enqueueQueueError
}

func (qcl *queueClientLocal) Dequeue(ctx context.Context, queueName string, dequeueQuery *queryassociatedaction.AssociatedActionQuery) (*v1willow.Item, func(), func(), *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "Dequeue")
	dequeueQueueError := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	var dequeueItem *v1willow.Item
	var onSuccess func()
	var onFailure func()
	onFind := func(key datatypes.EncapsulatedValue, item any) bool {
		dequeueItem, onSuccess, onFailure, dequeueQueueError = qcl.queueChannelsClient.DequeueQueueItem(ctx, queueName, dequeueQuery)
		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
		return nil, nil, nil, errors.InternalServerError
	}

	return dequeueItem, onSuccess, onFailure, dequeueQueueError
}

func (qcl *queueClientLocal) DeleteChannel(ctx context.Context, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "DeleteChannel")
	deleteChannelsError := errorMissingQueueName(queueName)

	onFind := func(key datatypes.EncapsulatedValue, item any) bool {
		deleteChannelsError = qcl.queueChannelsClient.DeleteChannel(ctx, queueName, channelKeyValues)
		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
		return errors.InternalServerError
	}

	return deleteChannelsError
}

func (qcl *queueClientLocal) Ack(ctx context.Context, queueName string, ack *v1willow.ACK) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "Ack")
	ackErr := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	onFind := func(key datatypes.EncapsulatedValue, item any) bool {
		ackErr = qcl.queueChannelsClient.ACK(ctx, queueName, ack)
		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			logger.Warn("failed to ack item. Queue by that name is currenly destroying")
			return &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is currently being destroyed. Refusing to acck the item since it is being destroyed too", queueName), StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return ackErr
}

func (qcl *queueClientLocal) Heartbeat(ctx context.Context, queueName string, heartbeat *v1willow.Heartbeat) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "Heartbeat")
	heartbeatErr := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	onFind := func(key datatypes.EncapsulatedValue, item any) bool {
		heartbeatErr = qcl.queueChannelsClient.Heartbeat(ctx, queueName, heartbeat)
		return false
	}

	if err := qcl.queues.Find(datatypes.String(queueName), v1.TypeRestrictions{MinDataType: datatypes.T_string, MaxDataType: datatypes.T_string}, onFind); err != nil {
		switch err {
		case btree.ErrorKeyDestroying:
			logger.Warn("failed to update queue. Queue by that name is currenly destroying")
			return &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is currently being destroyed. Refusing to update the queue", queueName), StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return heartbeatErr
}
