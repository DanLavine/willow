package v1

import (
	"encoding/json"
	"fmt"
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
func (eir *EnqueueItemRequest) EncodeJSON() ([]byte, error) {
	return json.Marshal(eir)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse EnqueueItemRequest from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (eir *EnqueueItemRequest) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, eir); err != nil {
		return err
	}

	return nil
}
