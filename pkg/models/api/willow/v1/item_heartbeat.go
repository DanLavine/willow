package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Heartbeat struct {
	// ID of the original message being acknowledged
	ItemID string

	// KeyValues for the channel
	KeyValues datatypes.TypedKeyValues
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that ack response has all required fields set
func (heartbeat Heartbeat) Validate() *errors.ModelError {
	if heartbeat.ItemID == "" {
		return &errors.ModelError{Field: "ItemId", Err: fmt.Errorf("is an empty string")}
	}

	if err := heartbeat.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}
