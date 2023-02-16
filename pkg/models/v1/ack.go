package v1

type ACK struct {
	// ID of the original message being acknowledged
	ID uint64

	// name of the queue to ack
	Name string

	// specific tag that the original message was processed from
	Tags []string

	// Indicate a success or failure of the message
	Passed bool
}
