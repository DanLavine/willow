package queues

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	queuechannels "github.com/DanLavine/willow/internal/willow/brokers/queue_channels"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
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
}

func NewLocalQueueClient(queueConstructor QueueConstructor, queueChannelsClient queuechannels.QueueChannelsClient) *queueClientLocal {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &queueClientLocal{
		queueConstructor:    queueConstructor,
		queues:              tree,
		queueChannelsClient: queueChannelsClient,
	}
}

// Create the main queue and setup the limts on the Limiter service
func (qcl *queueClientLocal) CreateQueue(ctx context.Context, queueCreate *v1willow.QueueCreate) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("CreateQueue").With(zap.String("queue_name", queueCreate.Name))
	ctx = reporting.UpdateLogger(ctx, logger)

	var createQueueError *errors.ServerError
	bTreeOnCreate := func() any {
		var queue Queue
		queue, createQueueError = qcl.queueConstructor.New(ctx, queueCreate)
		if createQueueError != nil {
			return nil
		}

		return queue
	}

	if err := qcl.queues.Create(datatypes.String(queueCreate.Name), bTreeOnCreate); err != nil {
		switch err {
		case btree.ErrorKeyAlreadyExists:
			logger.Warn("failed to create queue. Queue already exists by that name")
			return &errors.ServerError{Message: fmt.Sprintf("Queue already exists with name '%s'", queueCreate.Name), StatusCode: http.StatusConflict}
		case btree.ErrorKeyDestroying:
			logger.Warn("failed to create queue. Queue by that name is currenly destroying")
			return &errors.ServerError{Message: fmt.Sprintf("Queue with name '%s' is currently being destroyed. Try again later", queueCreate.Name), StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to create a queue in the tree for some reason", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return createQueueError
}

func (qcl *queueClientLocal) ListQueues(ctx context.Context) (v1willow.Queues, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("ListQueues")

	var queues v1willow.Queues
	bTreeOnIterate := func(key datatypes.EncapsulatedValue, item any) bool {
		queue := item.(Queue)

		queues = append(queues, &v1willow.Queue{
			Name:         key.Value().(string),
			QueueMaxSize: queue.ConfiguredLimit(),
			Channels:     nil,
		})

		return true
	}

	if err := qcl.queues.Iterate(bTreeOnIterate); err != nil {
		logger.Error("error listing queues from tree", zap.Error(err))
		return nil, errors.InternalServerError
	}

	return queues, nil
}

func (qcl *queueClientLocal) GetQueue(ctx context.Context, queueName string, channelQuery *v1common.AssociatedQuery) (*v1willow.Queue, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("GetQueue").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)
	getQueueError := errorMissingQueueName(queueName)

	var queue *v1willow.Queue
	onFind := func(treeItem any) {
		willowQueue := treeItem.(Queue)

		queue = &v1willow.Queue{
			Name:         queueName,
			QueueMaxSize: willowQueue.ConfiguredLimit(),
			Channels:     qcl.queueChannelsClient.Channels(ctx, queueName, *channelQuery),
		}
	}

	if err := qcl.queues.Find(datatypes.String(queueName), onFind); err != nil {
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

func (qcl *queueClientLocal) UpdateQueue(ctx context.Context, queueName string, queueUpdate *v1willow.QueueUpdate) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("UpdateQueue").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)
	updateQueueError := errorMissingQueueName(queueName)

	bTreeOnFind := func(item any) {
		updateQueueError = item.(Queue).Update(ctx, queueName, queueUpdate)
	}

	if err := qcl.queues.Find(datatypes.String(queueName), bTreeOnFind); err != nil {
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
	logger := reporting.GetLogger(ctx).Named("DeleteQueue").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)

	var deleteQueueError *errors.ServerError
	destroyQueue := func(key datatypes.EncapsulatedValue, item any) bool {
		queue := item.(Queue)
		if deleteQueueError = qcl.queueChannelsClient.DestroyChannelsForQueue(ctx, queueName); deleteQueueError == nil {
			if deleteQueueError = queue.Destroy(ctx, queueName); deleteQueueError == nil {
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

func (qcl *queueClientLocal) Enqueue(ctx context.Context, queueName string, enqueueItem *v1willow.EnqueueQueueItem) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("Enqueue").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)
	enqueueQueueError := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	onFind := func(_ any) {
		enqueueQueueError = qcl.queueChannelsClient.EnqueueQueueItem(ctx, queueName, enqueueItem)
	}

	if err := qcl.queues.Find(datatypes.String(queueName), onFind); err != nil {
		logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
		return errors.InternalServerError
	}

	return enqueueQueueError
}

func (qcl *queueClientLocal) Dequeue(ctx context.Context, queueName string, dequeueQuery datatypes.AssociatedKeyValuesQuery) (*v1willow.DequeueQueueItem, func(), func(), *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("Dequeue").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)
	dequeueQueueError := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	var dequeueItem *v1willow.DequeueQueueItem
	var onSuccess func()
	var onFailure func()
	onFind := func(_ any) {
		dequeueItem, onSuccess, onFailure, dequeueQueueError = qcl.queueChannelsClient.DequeueQueueItem(ctx, queueName, dequeueQuery)
	}

	if err := qcl.queues.Find(datatypes.String(queueName), onFind); err != nil {
		logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
		return nil, nil, nil, errors.InternalServerError
	}

	return dequeueItem, onSuccess, onFailure, dequeueQueueError
}

func (qcl *queueClientLocal) DeleteChannel(ctx context.Context, queueName string, channelKeyValues datatypes.KeyValues) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("DeleteChannel").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)
	deleteChannelsError := errorMissingQueueName(queueName)

	onFind := func(_ any) {
		deleteChannelsError = qcl.queueChannelsClient.DeleteChannel(ctx, queueName, channelKeyValues)
	}

	if err := qcl.queues.Find(datatypes.String(queueName), onFind); err != nil {
		logger.Error("failed to find the queue in the tree for some reason", zap.Error(err))
		return errors.InternalServerError
	}

	return deleteChannelsError
}

func (qcl *queueClientLocal) Ack(ctx context.Context, queueName string, ack *v1willow.ACK) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("Ack").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)
	ackErr := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	onFind := func(_ any) {
		ackErr = qcl.queueChannelsClient.ACK(ctx, queueName, ack)
	}

	if err := qcl.queues.Find(datatypes.String(queueName), onFind); err != nil {
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
	logger := reporting.GetLogger(ctx).Named("CreateQueue").With(zap.String("queue_name", queueName))
	ctx = reporting.UpdateLogger(ctx, logger)
	heartbeatErr := errorMissingQueueName(queueName)

	// nothing for the item to do. but use the bTree as a guard to ensure no delete operations are happening at the same time
	onFind := func(_ any) {
		heartbeatErr = qcl.queueChannelsClient.Heartbeat(ctx, queueName, heartbeat)
	}

	if err := qcl.queues.Find(datatypes.String(queueName), onFind); err != nil {
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
