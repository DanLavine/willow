package deadletterqueue

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"

	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/dead-letter-queue/disk"
	"github.com/DanLavine/willow/pkg/models"
)

type diskDeadLetterQueue struct {
	lock *sync.RWMutex

	baseDir string

	// map[queue name][tag name]Encoder
	encoders map[string]map[string]Encoder
}

func NewDiskDeadLetterQueue(baseDir string) *diskDeadLetterQueue {
	return &diskDeadLetterQueue{
		lock:     new(sync.RWMutex),
		baseDir:  baseDir,
		encoders: map[string]map[string]Encoder{},
	}
}

func (ddlq *diskDeadLetterQueue) Create(name, tag string) *v1.Error {
	ddlq.lock.Lock()
	defer ddlq.lock.Unlock()

	// if it already exists, just return nil
	if queues, ok := ddlq.encoders[name]; ok {
		if _, found := queues[tag]; found {
			return nil
		} else {
			// queue exists, but not that tag
			encoder, err := disk.NewDiskEncoder(ddlq.baseDir, ddlq.queueName(name), ddlq.queueTag(tag))
			if err != nil {
				return err
			}

			ddlq.encoders[name][tag] = encoder
		}
	} else {
		// the entire ques does not exist
		encoder, err := disk.NewDiskEncoder(ddlq.baseDir, ddlq.queueName(name), ddlq.queueTag(tag))
		if err != nil {
			return err
		}

		ddlq.encoders[name] = map[string]Encoder{tag: encoder}
	}

	return nil
}

func (ddlq *diskDeadLetterQueue) Enqueue(data []byte, updateable bool, brokerName, brokerTag string) *v1.Error {
	encoder, err := ddlq.getEncoder(brokerName, brokerTag)
	if err != nil {
		return err
	}

	return encoder.Enqueue(data)
}

func (ddlq *diskDeadLetterQueue) Message(brokerName, brokerTag string) (*v1.DequeueMessage, *v1.Error) {
	encoder, err := ddlq.getEncoder(brokerName, brokerTag)
	if err != nil {
		return nil, err
	}

	return encoder.Next()
}

func (ddlq *diskDeadLetterQueue) ACK(id int, passed bool, brokerName, brokerTag string) *v1.Error {
	encoder, err := ddlq.getEncoder(brokerName, brokerTag)
	if err != nil {
		return err
	}

	if passed {
		return encoder.Remove(id)
	}

	return encoder.Requeue(id)
}

func (ddlq *diskDeadLetterQueue) Metrics() *models.Metrics {
	ddlq.lock.RLock()
	defer ddlq.lock.RUnlock()

	metrics := models.NewMetrics()

	for queueName, encoders := range ddlq.encoders {
		for tag, encoder := range encoders {
			// TODO ignore errors for now
			_ = metrics.Add(queueName, tag, encoder.Metrics())
		}
	}

	return metrics
}

func (ddlq *diskDeadLetterQueue) getEncoder(brokerName, brokerTag string) (Encoder, *v1.Error) {
	ddlq.lock.RLock()
	defer ddlq.lock.RUnlock()

	if queues, ok := ddlq.encoders[brokerName]; ok {
		if encoder, found := queues[brokerTag]; found {
			return encoder, nil
		}
	}

	return nil, &v1.Error{Message: "Failed to find queue", StatusCode: http.StatusBadRequest}
}

func (ddlq *diskDeadLetterQueue) queueName(brokerName string) string {
	sum := sha256.Sum256([]byte(brokerName))
	return fmt.Sprintf("%x", sum)
}

func (ddlq *diskDeadLetterQueue) queueTag(brokerTag string) string {
	sum := sha256.Sum256([]byte(brokerTag))
	return fmt.Sprintf("%x", sum)
}
