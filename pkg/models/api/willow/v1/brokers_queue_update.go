package v1

import (
	"encoding/json"
)

type QueueUpdate struct {
	// Max size of the queue
	QueueMaxSize int64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (qu *QueueUpdate) Validate() error {
	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (qu *QueueUpdate) EncodeJSON() ([]byte, error) {
	return json.Marshal(qu)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (qu *QueueUpdate) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, qu); err != nil {
		return err
	}

	return nil
}
