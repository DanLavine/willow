package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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
func (qc QueueCreate) Validate() *errors.ModelError {
	if qc.Name == "" {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("is an empty string")}
	}

	return nil
}
