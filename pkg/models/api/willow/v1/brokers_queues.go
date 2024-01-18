package v1

import (
	"encoding/json"
	"fmt"
)

type Queues []*Queue

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (q *Queues) Validate() error {
	if len(*q) == 0 {
		return nil
	}

	for index, singleQueue := range *q {
		if singleQueue == nil {
			return fmt.Errorf("error at queues index: %d: the queue is nil", index)
		}

		if err := singleQueue.Validate(); err != nil {
			return fmt.Errorf("error at queues index %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (q *Queues) EncodeJSON() ([]byte, error) {
	return json.Marshal(q)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (q *Queues) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, q); err != nil {
		return err
	}

	return nil
}
