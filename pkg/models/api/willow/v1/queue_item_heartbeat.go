package v1

import (
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
func (heartbeat Heartbeat) Validate() error {
	if heartbeat.ItemID == "" {
		return fmt.Errorf("'ItemID' is the empty string")
	}

	if err := heartbeat.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return err
	}

	return nil
}
