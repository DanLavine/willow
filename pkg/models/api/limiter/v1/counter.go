package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Counter is the full api model that is returned as part of the query operations
type Counter struct {
	// KeyValues that define the counter
	KeyValues datatypes.KeyValues

	// Total number of counters for the particular KeyValues
	Counters int64
}

//	RETURNS:
//	- error - error describing any possible issues with the Counter and the steps to rectify them
//
// Validate ensures the Lock has all required fields set
func (counter *Counter) Validate() *errors.ModelError {
	if len(counter.KeyValues) == 0 {
		return &errors.ModelError{Field: "KeyValues", Err: fmt.Errorf("need to have a length of at least 1")}
	}

	if err := counter.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err}
	}

	return nil
}
