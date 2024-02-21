package v1

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type EnqueueQueueItem struct {
	// Item in bytes to enqueue
	Item []byte

	// KeyValues that identify the Item
	KeyValues datatypes.KeyValues

	// If the item can be updated on another request
	Updateable bool

	// How many attempts to retry in the case of a failure
	RetryAttempts uint64

	// Where to enqueue the item on a failed attempt
	RetryPosition string

	// How long to wait for heartbeats untill the item is considered failed
	TimeoutDuration time.Duration
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (eqi *EnqueueQueueItem) Validate() error {
	if len(eqi.Item) == 0 {
		return fmt.Errorf("'Item' to enqueue is empty")
	}

	if err := eqi.KeyValues.Validate(); err != nil {
		return err
	}

	switch eqi.RetryPosition {
	case "front", "back":
		// these are fine
	default:
		return fmt.Errorf("'RetryPosition' is unkown, must be [front | back]")
	}

	if eqi.TimeoutDuration < time.Second {
		return fmt.Errorf("'TimeoutDuration' is less than the minimum value of 1 second")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (eqi *EnqueueQueueItem) EncodeJSON() ([]byte, error) {
	return json.Marshal(eqi)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse Create from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the data
//
// DecodeJSON can convertes the encoded byte array into the Object Decode was called on
func (eqi *EnqueueQueueItem) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, eqi); err != nil {
		return err
	}

	if eqi.RetryPosition == "" {
		eqi.RetryPosition = "front"
	}

	return nil
}
