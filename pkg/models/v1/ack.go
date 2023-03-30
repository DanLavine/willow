package v1

type ACK struct {
	// common broker info
	BrokerInfo BrokerInfo

	// ID of the original message being acknowledged
	ID uint64

	// Indicate a success or failure of the message
	Passed bool
}
