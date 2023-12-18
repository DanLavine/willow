package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Metrics for all queues and dead letter queues
type MetricsResponse struct {
	// all queue metrics
	Queues []*QueueMetricsResponse
}

// Metrics for each individual queue
// TODO put response on this
type QueueMetricsResponse struct {
	Name  string
	Total uint64
	Max   uint64

	Tags []*TagMetricsResponse

	DeadLetterQueueMetrics *DeadLetterQueueMetricsResponse
}

// Metrics for all tags Groups for a given queue
type TagMetricsResponse struct {
	Tags       datatypes.KeyValues
	Ready      uint64
	Processing uint64
}

// Metrics for the dead letter queue
type DeadLetterQueueMetricsResponse struct {
	Count uint64
}

func (m *MetricsResponse) ToBytes() []byte {
	data, _ := json.Marshal(m)
	return data
}
