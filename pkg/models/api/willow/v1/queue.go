package v1

import (
	"fmt"
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
func (q Queue) Validate() error {
	if q.Name == "" {
		return fmt.Errorf("'Name' is the empty string")
	}

	if err := q.Channels.Validate(); err != nil {
		return err
	}

	return nil
}
