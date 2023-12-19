package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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
func (resp *MetricsResponse) EncodeJSON() []byte {
	data, _ := json.Marshal(resp)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream. Valida values [application/json]
//	- reader - stream to read the encoded CreateLockResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// Decode can easily parse the response body from an http create request
func (resp *MetricsResponse) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, resp); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return resp.Validate()
}
