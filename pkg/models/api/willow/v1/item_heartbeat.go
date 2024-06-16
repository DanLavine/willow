package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	dbdefinition "github.com/DanLavine/willow/pkg/models/api/common/v1/db_definition"
)

type Heartbeat struct {
	// ID of the original message being acknowledged
	ItemID string

	// KeyValues for the channel
	KeyValues dbdefinition.TypedKeyValues
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that ack response has all required fields set
func (heartbeat Heartbeat) Validate() *errors.ModelError {
	if heartbeat.ItemID == "" {
		return &errors.ModelError{Field: "ItemId", Err: fmt.Errorf("is an empty string")}
	}

	if err := heartbeat.KeyValues.Validate(); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}
