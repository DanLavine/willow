package v1

type Create struct {
	// type of broker we want to create
	BrokerType BrokerType

	// Tag for the broker
	BrokerTags []string
}
