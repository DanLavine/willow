package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type EnqueueItemRequest struct {
	// common broker info
	BrokerInfo

	// Message body that will be used by clients receiving this message
	Data []byte

	// If the message should be updatable
	// If set to true:
	//   1. Will colapse on the previous message if it has not been processed and is also updateable
	// If set to false:
	//   1. Will enque the messge as unique and won't be collapsed on. Can still collapse the previous message iff that was true
	Updateable bool
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (eir *EnqueueItemRequest) Validate() error {
	if err := eir.validate(); err != nil {
		return err
	}

	if len(eir.Data) == 0 {
		return fmt.Errorf("'Data' is empty, but is a requied field")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (eir *EnqueueItemRequest) EncodeJSON() []byte {
	data, _ := json.Marshal(eir)
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
func (eir *EnqueueItemRequest) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, eir); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return eir.Validate()
}
