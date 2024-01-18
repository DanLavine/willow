package v1

import (
	"encoding/json"
	"fmt"
)

type Queue struct {
	// Name of the broker object
	Name string

	// Max size of the queue
	// -1 means unlimited
	QueueMaxSize int64

	// All the channels that belong to this queue
	Channels Channels
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (q *Queue) Validate() error {
	if q.Name == "" {
		return fmt.Errorf("'Name' is the empty string")
	}

	if err := q.Channels.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (q *Queue) EncodeJSON() ([]byte, error) {
	return json.Marshal(q)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (q *Queue) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, q); err != nil {
		return err
	}

	return nil
}
