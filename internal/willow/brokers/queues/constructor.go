package queues

import (
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/willow/brokers/queues/memory"

	servererrors "github.com/DanLavine/willow/internal/server_errors"
	v1willow "github.com/DanLavine/willow/pkg/models/api/willow/v1"
)

//go:generate mockgen -destination=queuesfakes/constructor_mock.go -package=queuesfakes github.com/DanLavine/willow/internal/willow/brokers/queues QueueConstructor
type QueueConstructor interface {
	// create a new Queue
	NewQueue(createParams *v1willow.Create) (ManagedQueue, *servererrors.ApiError)
}

type queueConstructor struct {
	config *config.WillowConfig
}

func NewQueueConstructor(cfg *config.WillowConfig) *queueConstructor {
	return &queueConstructor{
		config: cfg,
	}
}

func (qc *queueConstructor) NewQueue(create *v1willow.Create) (ManagedQueue, *servererrors.ApiError) {
	switch *qc.config.StorageConfig.Type {
	//case config.DiskStorage:
	//	return disk.NewQueue(qc.config.StorageConfig.Disk.StorageDir, create)
	case config.MemoryStorage:
		return memory.NewQueue(create), nil
	default:
		return nil, servererrors.UnknownStorageType
	}
}
