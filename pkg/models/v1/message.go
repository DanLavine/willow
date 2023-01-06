package v1

type EnqueMessage struct {
	// Either Queue or PubSub messages
	BrokerType BrokerType

	// specific tag for a queue
	BrokerTags []string

	// Message body that will be used by clients receiving this message
	Data []byte

	// If the message should be updatable
	// If set to true:
	//   1. Will colapse on the previous message if it has not been processed and is also updateable
	// If set to false:
	//   1. Will enque the messge as unique and won't be collapsed on
	Updateable bool
}

type DequeueMessage struct {
	// ID of the message that can be ACKed
	ID uint64

	// specific tag for a queue
	BrokerTags []string

	// Message body that will be used by clients receiving this message
	Data []byte
}
