package v1

type MatchRestriction int

const (
	// Must be matched exactly
	STRICT MatchRestriction = iota

	// If the Broker's Tags contain all requested tags -> true
	SUBSET

	// If any tags requested are in the Broker's Tags -> true
	ANY

	// ignore the brokers tags. Any enqued message is valid
	ALL
)

// Used to notify Willow that a client is ready to accept a message for a particular broker
type Ready struct {
	// Either Queue or PubSub messages
	BrokerType BrokerType

	// Tag of the broker we want to read from
	BrokerTagsMatch MatchRestriction
	BrokerTags      []string
}
