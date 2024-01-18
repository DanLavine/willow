package v1

import (
	"encoding/json"
	"fmt"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Heartbeat struct {
	// ID of the original message being acknowledged
	ItemID string

	// KeyValues for the channel
	KeyValues datatypes.KeyValues
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that ack response has all required fields set
func (heartbeat *Heartbeat) Validate() error {
	if heartbeat.ItemID == "" {
		return fmt.Errorf("'ItemID' is the empty string")
	}

	if err := heartbeat.KeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (heartbeat *Heartbeat) EncodeJSON() ([]byte, error) {
	return json.Marshal(heartbeat)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse ACK from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (heartbeat *Heartbeat) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, heartbeat); err != nil {
		return err
	}

	return nil
}
