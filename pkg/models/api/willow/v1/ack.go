package v1

import (
	"encoding/json"
	"fmt"
)

type RequeueLocation uint

const (
	RequeueFront RequeueLocation = iota
	RequeueEnd
	RequeueNone
)

type ACK struct {
	// common broker info
	BrokerInfo

	// ID of the original message being acknowledged
	ID string

	// Indicate a success or failure of the message
	Passed          bool
	RequeueLocation RequeueLocation // only used when set to false
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that ack response has all required fields set
func (ack *ACK) Validate() error {
	if err := ack.BrokerInfo.validate(); err != nil {
		return err
	}

	if ack.ID == "" {
		return fmt.Errorf("'ID' is the empty string")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (ack *ACK) EncodeJSON() ([]byte, error) {
	return json.Marshal(ack)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse ACK from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// Decode can convertes the encoded byte array into the Object Decode was called on
func (ack *ACK) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, ack); err != nil {
		return err
	}

	return nil
}
