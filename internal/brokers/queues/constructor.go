package queues

import (
	"github.com/DanLavine/willow/internal/brokers/queues/memory"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/pkg/config"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

//counterfeiter:generate . QueueConstructor
type QueueConstructor interface {
	// create a new Queue
	NewQueue(createParams *v1.Create) (ManagedQueue, *v1.Error)
}

type queueConstructor struct {
	config *config.Config
}

func NewQueueConstructor(cfg *config.Config) *queueConstructor {
	return &queueConstructor{
		config: cfg,
	}
}

func (qc *queueConstructor) NewQueue(create *v1.Create) (ManagedQueue, *v1.Error) {
	switch qc.config.StorageConfig.Type {
	//case config.DiskStorage:
	//	return disk.NewQueue(qc.config.StorageConfig.Disk.StorageDir, create)
	case config.MemoryStorage:
		return memory.NewQueue(create), nil
	default:
		return nil, errors.UnknownQueueStorage
	}
}
