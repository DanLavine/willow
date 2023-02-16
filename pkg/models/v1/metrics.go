package v1

import (
	"encoding/json"
)

// Metrics for all queues and dead letter queues
type Metrics struct {
	// all queue metrics
	Queues []QueueMetrics
}

// Metrics for each individual queue
type QueueMetrics struct {
	Name       string
	Ready      uint64
	Processing uint64
	Max        uint64

	DeadLetterQueueMetrics *DeadLetterQueueMetrics
}

// Metrics for the dead letter queue
type DeadLetterQueueMetrics struct {
	Count uint64
}

func (m *Metrics) ToBytes() ([]byte, *Error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, MarshelModelFailed.With("", err.Error())
	}

	return data, nil
}

func (qm *QueueMetrics) ToBytes() ([]byte, *Error) {
	data, err := json.Marshal(qm)
	if err != nil {
		return nil, MarshelModelFailed.With("", err.Error())
	}

	return data, nil
}

func (dlqm *DeadLetterQueueMetrics) ToBytes() ([]byte, *Error) {
	data, err := json.Marshal(dlqm)
	if err != nil {
		return nil, MarshelModelFailed.With("", err.Error())
	}

	return data, nil
}
