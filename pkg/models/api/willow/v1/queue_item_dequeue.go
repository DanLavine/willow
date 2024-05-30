package v1

import (
	"fmt"
	"time"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type DequeueQueueItem struct {
	// ID of the item that needs to be heartbeat and acked
	ItemID string

	// KeyValues for what Channel the item was pulled from
	KeyValues datatypes.KeyValues

	// Item body that will be used by clients receiving this message
	Item []byte

	// How long the Item is valid for before the service considers it as failed
	TimeoutDuration time.Duration
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (dqi DequeueQueueItem) Validate() *errors.ModelError {
	if dqi.ItemID == "" {
		return &errors.ModelError{Field: "ItemID", Err: fmt.Errorf("is the empty string")}
	}

	if len(dqi.Item) == 0 {
		return &errors.ModelError{Field: "Item", Err: fmt.Errorf("data is null")}
	}

	if err := dqi.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	if dqi.TimeoutDuration < time.Second {
		return &errors.ModelError{Field: "TimeoutDuration", Err: fmt.Errorf("is less than the minimum of 1 second")}
	}

	return nil
}
