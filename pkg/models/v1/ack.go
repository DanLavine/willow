package v1

type ACK struct {
	// ID of the original message being acknowledged
	ID uint64

	// Either Queue or PubSub messages
	BrokerType BrokerType

	// Tag for the broker
	BrokerTags []string

	// Indicate a success or failure of the message
	Passed bool
}
