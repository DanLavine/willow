package v1

import (
	"fmt"
)

type Queues []*Queue

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (q Queues) Validate() error {
	if len(q) == 0 {
		return nil
	}

	for index, singleQueue := range q {
		if err := singleQueue.Validate(); err != nil {
			return fmt.Errorf("error at queues index %d: %w", index, err)
		}
	}

	return nil
}
