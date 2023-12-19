package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type Create struct {
	// Name of the broker object
	Name string

	// max size of the dead letter queue
	// Cannot be set to  0
	QueueMaxSize uint64

	// Max Number of items to keep in the dead letter queue. If full,
	// any new items will just be dropped untill the queue is cleared by an admin.
	DeadLetterQueueMaxSize uint64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (c *Create) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("'Name' is the empty string")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (c *Create) EncodeJSON() []byte {
	data, _ := json.Marshal(c)
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
func (c *Create) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, c); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return c.Validate()
}
