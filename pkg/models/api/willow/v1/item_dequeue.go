package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type DequeueItemRequest struct {
	// common broker info
	Name string

	// query for what readeers to select
	Query datatypes.AssociatedKeyValuesQuery
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (req *DequeueItemRequest) Validate() error {
	if req.Name == "" {
		return fmt.Errorf("'Name' is the empty string, but requires a proper value")
	}

	if err := req.Query.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (req *DequeueItemRequest) EncodeJSON() []byte {
	data, _ := json.Marshal(req)
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
func (req *DequeueItemRequest) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, req); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return req.Validate()
}

type DequeueItemResponse struct {
	// common broker info
	BrokerInfo BrokerInfo

	// ID of the message that can be ACKed
	ID string

	// Message body that will be used by clients receiving this message
	Data []byte
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (resp *DequeueItemResponse) Validate() error {
	if err := resp.BrokerInfo.validate(); err != nil {
		return err
	}

	if resp.ID == "" {
		return fmt.Errorf("'ID' is set as the empty string")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (resp *DequeueItemResponse) EncodeJSON() []byte {
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
func (resp *DequeueItemResponse) Decode(contentType api.ContentType, reader io.ReadCloser) error {
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
