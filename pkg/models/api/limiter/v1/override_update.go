package v1

import "fmt"

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
func (overrideUpdate *OverrideUpdate) Validate() error {
	if overrideUpdate.Limit < -1 {
		return fmt.Errorf("Limit is less than -1")
	}

	return nil
}
