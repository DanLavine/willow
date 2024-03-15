package v1

import (
	"encoding/json"
	"fmt"
)

type Channels []Channel

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (c Channels) Validate() error {
	if len(c) == 0 {
		return nil
	}

	for index, singleChan := range c {
		if err := singleChan.Validate(); err != nil {
			return fmt.Errorf("error at channel index %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (c Channels) EncodeJSON() ([]byte, error) {
	return json.Marshal(c)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (c Channels) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}
