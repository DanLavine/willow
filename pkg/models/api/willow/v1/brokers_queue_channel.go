package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Channel struct {
	// KeyValues that define the channel
	KeyValues datatypes.KeyValues

	// Total number of enqueued items including running
	EnqueuedItems int64

	// Total numbe of running items for this channel
	RunningItems int64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (c *Channel) Validate() error {
	if err := c.KeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (c *Channel) EncodeJSON() ([]byte, error) {
	return json.Marshal(c)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (c *Channel) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}
