package v1

import (
	"fmt"
)

type QueueCreate struct {
	// Name of the broker object
	Name string

	// Max size of the queue
	// -1 means unlimited
	QueueMaxSize int64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (qc QueueCreate) Validate() error {
	if qc.Name == "" {
		return fmt.Errorf("'Name' is the empty string")
	}

	return nil
}
