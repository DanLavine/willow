package queues

import (
	"github.com/DanLavine/willow/internal/brokers/queues/memory"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1willow"
)

//go:generate mockgen -destination=queuesfakes/constructor_mock.go -package=queuesfakes github.com/DanLavine/willow/internal/brokers/queues QueueConstructor
type QueueConstructor interface {
	// create a new Queue
	NewQueue(createParams *v1willow.Create) (ManagedQueue, *api.Error)
}

type queueConstructor struct {
	config *config.WillowConfig
}

func NewQueueConstructor(cfg *config.WillowConfig) *queueConstructor {
	return &queueConstructor{
		config: cfg,
	}
}

func (qc *queueConstructor) NewQueue(create *v1willow.Create) (ManagedQueue, *api.Error) {
	switch *qc.config.StorageConfig.Type {
	//case config.DiskStorage:
	//	return disk.NewQueue(qc.config.StorageConfig.Disk.StorageDir, create)
	case config.MemoryStorage:
		return memory.NewQueue(create), nil
	default:
		return nil, errors.UnknownQueueStorage
	}
}
