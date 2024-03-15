package v1

import (
	"fmt"

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
	if c == nil {
		return fmt.Errorf("'Channel' can not be nil")
	}

	if err := c.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return err
	}

	return nil
}

type DeleteQueueChannel struct {
	KeyValues datatypes.KeyValues
}

func (c *DeleteQueueChannel) Validate() error {
	if c == nil {
		return fmt.Errorf("'DeleteQueueChannel' can not be nil")
	}

	if err := c.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return err
	}

	return nil
}
