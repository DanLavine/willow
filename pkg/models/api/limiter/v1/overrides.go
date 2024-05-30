package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// Overrides is used as part of a Query or Match lookup operation
type Overrides []*Override

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (overrides Overrides) Validate() *errors.ModelError {
	if len(overrides) == 0 {
		return nil
	}

	for index, override := range overrides {
		if override == nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Err: fmt.Errorf("override cannot be nil")}
		}

		if err := override.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Child: err}
		}
	}

	return nil
}
