package disk

import (
	"sync"

	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/v1/models"
	"github.com/DanLavine/willow/internal/v1/queues/disk/encoder"
	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type DiskDeadLetterQueue struct {
	lock *sync.Mutex

	// number of items in the dead letter queue
	size    uint64
	maxSize uint64

	// all the items in the dead letter queue
	items []*models.Location

	// disk encoder for dead letter queue
	encoderDeadLetter *encoder.EncoderDeadLetter
}

func NewDiskDeadLetterQueue(configQueue config.ConfigQueue, createParams v1.Create) (*DiskDeadLetterQueue, *v1.Error) {
	encoderDeadLetter, err := encoder.NewEncoderDeadLetter(configQueue.ConfigDisk.StorageDir, createParams.BrokerTags)
	if err != nil {
		return nil, err
	}

	return &DiskDeadLetterQueue{
		lock:              new(sync.Mutex),
		size:              0,
		maxSize:           createParams.DeadLetterQueueParams.MaxSize,
		items:             []*models.Location{},
		encoderDeadLetter: encoderDeadLetter,
	}, nil
}

func (ddlq *DiskDeadLetterQueue) Enqueue(data []byte) *v1.Error {
	ddlq.lock.Lock()
	defer ddlq.lock.Unlock()

	if ddlq.size >= ddlq.maxSize {
		return errors.DeadLetterQueueFull
	}

	diskLocation, err := ddlq.encoderDeadLetter.Write(data)
	if err != nil {
		return err
	}

	ddlq.items = append(ddlq.items, &models.Location{ID: ddlq.size, StartIndex: diskLocation.StartIndex, Size: diskLocation.Size})
	ddlq.size++

	return nil
}

func (ddlq *DiskDeadLetterQueue) Get(index uint64, tags []string) (*v1.DequeueMessage, *v1.Error) {
	ddlq.lock.Lock()
	defer ddlq.lock.Unlock()

	if ddlq.size == 0 || index > ddlq.size-1 {
		return nil, errors.DeadLetterItemNotfound
	}

	location := ddlq.items[index]

	data, err := ddlq.encoderDeadLetter.Read(location.StartIndex, location.Size)
	if err != nil {
		return nil, err
	}

	return &v1.DequeueMessage{ID: location.ID, BrokerTags: tags, Data: data}, nil
}

func (ddlq *DiskDeadLetterQueue) Count() uint64 {
	return ddlq.size
}

func (ddlq *DiskDeadLetterQueue) Clear() *v1.Error {
	ddlq.lock.Lock()
	defer ddlq.lock.Unlock()

	ddlq.items = []*models.Location{}
	ddlq.size = 0

	if err := ddlq.encoderDeadLetter.Clear(); err != nil {
		return err
	}

	return nil
}

func (ddlq *DiskDeadLetterQueue) Close() {
	ddlq.lock.Lock()
	defer ddlq.lock.Unlock()

	ddlq.encoderDeadLetter.Close()
}
