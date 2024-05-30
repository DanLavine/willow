package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type Queues []*Queue

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (q Queues) Validate() *errors.ModelError {
	if len(q) == 0 {
		return nil
	}

	for index, singleQueue := range q {
		if singleQueue == nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Err: fmt.Errorf("Queue cannot be null")}
		}

		if err := singleQueue.Validate(); err != nil {
			return &errors.ModelError{Field: fmt.Sprintf("[%d]", index), Child: err}
		}
	}

	return nil
}
