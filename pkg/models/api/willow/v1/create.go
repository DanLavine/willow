package v1

import (
	"encoding/json"
	"fmt"
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
func (c *Create) EncodeJSON() ([]byte, error) {
	return json.Marshal(c)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (c *Create) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}
