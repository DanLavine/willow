package v1

import (
	"fmt"
	"time"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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
func (eqi EnqueueQueueItem) Validate() *errors.ModelError {
	if len(eqi.Item) == 0 {
		return &errors.ModelError{Field: "Item", Err: fmt.Errorf("is null")}
	}

	if err := eqi.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	switch eqi.RetryPosition {
	case "front", "back":
		// these are fine
	default:
		return &errors.ModelError{Field: "RetryPosition", Err: fmt.Errorf("received an invalid value '%s', must be either [front | back]", eqi.RetryPosition)}
	}

	if eqi.TimeoutDuration < time.Second {
		return &errors.ModelError{Field: "TimeoutDuration", Err: fmt.Errorf("is less than the minimum value of 1 second")}
	}

	return nil
}

// func (eqi *EnqueueQueueItem) UnmarshalJSON(b []byte) error {
// 	custom := struct {
// 		// Item in bytes to enqueue
// 		Item []byte

// 		// KeyValues that identify the Item
// 		KeyValues datatypes.KeyValues

// 		// If the item can be updated on another request
// 		Updateable bool

// 		// How many attempts to retry in the case of a failure
// 		RetryAttempts uint64

// 		// Where to enqueue the item on a failed attempt
// 		RetryPosition string

// 		// How long to wait for heartbeats untill the item is considered failed
// 		TimeoutDuration time.Duration
// 	}{}

// 	if err := json.Unmarshal(b, &custom); err != nil {
// 		return err
// 	}

// 	eqi.Item = custom.Item
// 	eqi.KeyValues = custom.KeyValues
// 	eqi.Updateable = custom.Updateable
// 	eqi.RetryAttempts = custom.RetryAttempts
// 	eqi.RetryPosition = custom.RetryPosition
// 	eqi.TimeoutDuration = custom.TimeoutDuration

// 	// setup any defaults
// 	if eqi.RetryPosition == "" {
// 		eqi.RetryPosition = "front"
// 	}

// 	return nil
// }
