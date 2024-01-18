package v1

import (
	"encoding/json"
	"fmt"
)

type QueueCreate struct {
	// Name of the broker object
	Name string

	// Max size of the queue
	// -1 means unlimited
	QueueMaxSize int64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (qc *QueueCreate) Validate() error {
	if qc.Name == "" {
		return fmt.Errorf("'Name' is the empty string")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (qc *QueueCreate) EncodeJSON() ([]byte, error) {
	return json.Marshal(qc)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (qc *QueueCreate) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, qc); err != nil {
		return err
	}

	return nil
}
