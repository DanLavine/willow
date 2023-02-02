package queues

import (
	"context"

	v1 "github.com/DanLavine/willow/pkg/models/v1"
)

type Queue interface {
	// create a new queue with the given tags
	//
	// PARAMS:
	// * queueTags - the tags to create the queue with
	//
	// RETURNS:
	// * error - any errors with creating the queue
	Create(queueTags []string) *v1.Error

	// Enqueu a new message.
	//
	// PARAMS:
	// * data - the message to enqueue
	// * updateable - indcate if the message can be overwritte if no client is processing it when another message comes in
	// * queueTags - tags to match the queue where the message will be enqueue
	//
	// RETURNS:
	// * error - any errors with enquing the data
	Enqueue(data []byte, updateable bool, queueTags []string) *v1.Error

	// Retrieve the next item from a particular queue. This is a blocking operation
	//
	// PARAMS:
	// * ctx - context that can be used to cancel waiting for a message
	// * matchRestriction - type of queue we want to match our tags agains
	// * queueTags - tags to find the queue with in conjuction with the matchRestriction
	//
	// RETURNS:
	// * DequeueMessage - message that contains relivent info to process from the client and respond wif the message finished processing
	// * error - any errors with enquing the data
	Item(ctx context.Context, matchRestriction v1.MatchRestriction, queueTags []string) (*v1.DequeueMessage, *v1.Error)

	// Respond to a processing message for the client.
	//
	// PARAMS:
	// * id - id of the message that was originally processing
	// * passed - if the message passed or not. On a faiure the message will either be re-enqueued or sent to the dead letter queue.
	//            on success, the message will be removed entierly
	// * queueTags - tags to find the queue with
	//
	// RETURNS:
	// * error - any errors with processing the message
	ACK(id uint64, passed bool, queueTags []string) *v1.Error

	// Get basic metrics about all the current queues. Right now mostly helpful for debugging
	//
	// RETURNS:
	// * metrics - the metrics associated for all queues
	Metrics() *v1.Metrics

	// TODO
	// DeadLetterQueue() DeadLetterQueue
}

type DeadLetterQueue interface {
	// TODO
}
