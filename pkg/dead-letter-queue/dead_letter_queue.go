package deadletterqueue

import (
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/models"
)

// Dead Letter Queues defined in this package are responsible for
// handling the following logic:
//	1. Generate a queue for a particular message if it does not exist
//  2. Enqueue a message to the prescribed dead letter queue
//  3. Retrieve a message from the prescribed dead letter queue
//  4. ACK messages to the desired dead letter queue

type DeadLetterQueue interface {
	// Create a new dead letter queue for a specific tag
	Create(brokerName, brokerTag string) *v1.Error

	// Enqueue a new massage to a queue
	Enqueue(data []byte, updateable bool, brokerName, brokerTag string) *v1.Error

	// Blocking operation to retrieve the next message for a given queue
	Message(brokerName, brokerTag string) (*v1.DequeueMessage, *v1.Error)

	// ACK a message for a particular id and tag
	ACK(id int, passed bool, brokerName, brokerTag string) *v1.Error

	// retrieve metrics for all encoders
	Metrics() *models.Metrics
}

type Encoder interface {
	// encode any data to the dead letter queue
	Enqueue([]byte) *v1.Error

	// Retrive a value that is decode
	Get(id int) (*v1.DequeueMessage, *v1.Error)

	// Requeue an item if it was not successfully processed
	Requeue(id int) *v1.Error

	// Remove an item if it was successfully processed
	Remove(id int) *v1.Error

	// retrieve the next item
	Next() (*v1.DequeueMessage, *v1.Error)

	// retrieve all metrics for a particular encoder
	Metrics() models.QueueMetrics

	// when shutting down, encoder might want to clean up some resorces
	Close() error
}
