package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type ACK struct {
	// ID of the original message being acknowledged
	ItemID string

	// KeyValues for the channel
	KeyValues datatypes.TypedKeyValues

	// Indicate a success or failure of the message
	Passed bool
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that ack response has all required fields set
func (ack ACK) Validate() *errors.ModelError {
	if ack.ItemID == "" {
		return &errors.ModelError{Field: "ItemID", Err: fmt.Errorf("is an empty string")}
	}

	if err := ack.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}
