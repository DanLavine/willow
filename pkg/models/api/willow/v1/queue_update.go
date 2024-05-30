package v1

import "github.com/DanLavine/willow/pkg/models/api/common/errors"

type QueueUpdate struct {
	// Max size of the queue
	QueueMaxSize int64
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (qu QueueUpdate) Validate() *errors.ModelError {
	return nil
}
