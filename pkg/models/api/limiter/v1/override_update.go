package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// OverrideUpdate is used to update a particular override
type OverrideUpdate struct {
	// The new limit to use for the paricular Override
	// Setting this value to -1 means unlimited
	Limit int64
}

//	RETURNS:
//	- error - any errors encountered with the OverrideUpdate
//
// Validate is used to ensure that Override has all required fields set
func (overrideUpdate *OverrideUpdate) Validate() *errors.ModelError {
	if overrideUpdate.Limit < -1 {
		return &errors.ModelError{Field: "Limit", Err: fmt.Errorf("recevied a value less than -1")}
	}

	return nil
}
