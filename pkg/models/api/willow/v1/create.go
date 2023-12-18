package v1

import (
	"encoding/json"
)

type Create struct {
	// Name of the broker object
	Name string

	// max size of the dead letter queue
	// Cannot be set to  0
	QueueMaxSize uint64

	// Max Number of items to keep in the dead letter queue. If full,
	// any new items will just be dropped untill the queue is cleared by an admin.
	DeadLetterQueueMaxSize uint64
}

func (c *Create) ToBytes() []byte {
	data, _ := json.Marshal(c)
	return data
}
