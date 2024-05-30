package v1

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type Queue struct {
	// Name of the broker object
	Name string `json:"Name"`

	// Max size of the queue
	// -1 means unlimited
	QueueMaxSize int64 `json:"QueueMaxSize"`

	// All the channels that belong to this queue
	// This is a Read-Only operation
	Channels Channels `json:"Channels,omitempty"`
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that Create has all required fields set
func (q Queue) Validate() *errors.ModelError {
	if q.Name == "" {
		return &errors.ModelError{Field: "Name", Err: fmt.Errorf("is an empty string")}
	}

	if err := q.Channels.Validate(); err != nil {
		return &errors.ModelError{Field: "Channels", Child: err}
	}

	return nil
}
