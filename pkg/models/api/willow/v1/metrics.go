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

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (resp *MetricsResponse) Validate() error {
	// #TODO
	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (resp *MetricsResponse) EncodeJSON() ([]byte, error) {
	return json.Marshal(resp)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Metrics from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (resp *MetricsResponse) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, resp); err != nil {
		return err
	}

	return nil
}
