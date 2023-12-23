package v1

import (
	"encoding/json"
	"fmt"

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
func (req *DequeueItemRequest) EncodeJSON() ([]byte, error) {
	return json.Marshal(req)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse DequeueItemRequest from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (req *DequeueItemRequest) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, req); err != nil {
		return err
	}

	return nil
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
func (resp *DequeueItemResponse) EncodeJSON() ([]byte, error) {
	return json.Marshal(resp)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse DequeueItemResponse from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (resp *DequeueItemResponse) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, resp); err != nil {
		return err
	}

	return nil
}
