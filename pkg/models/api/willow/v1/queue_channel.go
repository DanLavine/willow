package v1

import (
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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
func (c *Channel) Validate() *errors.ModelError {
	if err := c.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}

type DeleteQueueChannel struct {
	KeyValues datatypes.KeyValues
}

func (c *DeleteQueueChannel) Validate() *errors.ModelError {
	if err := c.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}
