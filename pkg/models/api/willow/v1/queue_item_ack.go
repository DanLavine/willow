package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type ACK struct {
	// ID of the original message being acknowledged
	ItemID string

	// KeyValues for the channel
	KeyValues datatypes.KeyValues

	// Indicate a success or failure of the message
	Passed bool
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that ack response has all required fields set
func (ack ACK) Validate() error {
	if ack.ItemID == "" {
		return fmt.Errorf("'ID' is the empty string")
	}

	if err := ack.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return err
	}

	return nil
}
