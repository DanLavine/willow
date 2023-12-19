package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type RequeueLocation uint

const (
	RequeueFront RequeueLocation = iota
	RequeueEnd
	RequeueNone
)

type ACK struct {
	// common broker info
	BrokerInfo

	// ID of the original message being acknowledged
	ID string

	// Indicate a success or failure of the message
	Passed          bool
	RequeueLocation RequeueLocation // only used when set to false
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that ack response has all required fields set
func (ack *ACK) Validate() error {
	if err := ack.BrokerInfo.validate(); err != nil {
		return err
	}

	if ack.ID == "" {
		return fmt.Errorf("'ID' is the empty string")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (ack *ACK) EncodeJSON() []byte {
	data, _ := json.Marshal(ack)
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
func (ack *ACK) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, ack); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return ack.Validate()
}
