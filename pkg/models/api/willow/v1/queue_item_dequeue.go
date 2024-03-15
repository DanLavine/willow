package v1

import (
	"fmt"
	"time"

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
func (dqi DequeueQueueItem) Validate() error {
	if dqi.ItemID == "" {
		return fmt.Errorf("'ItemID' is empty")
	}

	if len(dqi.Item) == 0 {
		return fmt.Errorf("'Item' to dequeue is empty")
	}

	if err := dqi.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return err
	}

	if dqi.TimeoutDuration < time.Second {
		return fmt.Errorf("'TimeoutDuration' is less than the minimum value of 1 second")
	}

	return nil
}
