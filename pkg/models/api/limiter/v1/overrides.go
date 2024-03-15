package v1

import (
	"fmt"
)

// Overrides is used as part of a Query or Match lookup operation
type Overrides []*Override

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (overrides Overrides) Validate() error {
	if len(overrides) == 0 {
		return nil
	}

	for index, override := range overrides {
		if override == nil {
			return fmt.Errorf("error at Overrides index %d. Override can not be nil", index)
		}

		if err := override.Validate(); err != nil {
			return fmt.Errorf("error at overrides index %d: %w", index, err)
		}
	}

	return nil
}
