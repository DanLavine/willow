package v1

import "github.com/DanLavine/willow/pkg/models/datatypes"

type BrokerType uint32

const (
	Queue BrokerType = iota
)

type BrokerInfo struct {
	// specific queue name for the message
	Name datatypes.String

	// Type of broker
	// NOTE: not currently used
	BrokerType BrokerType

	// possible tags used by the broker
	Tags datatypes.Strings
}

func (b BrokerInfo) validate() *Error {
	if b.Name == "" {
		return InvalidRequestBody.With("Name to be provided", "Name is the empty string")
	}

	if len(b.Tags) == 0 {
		b.Tags = []datatypes.String{""}
	}

	b.Tags.Sort()
	return nil
}

func (bt BrokerType) ToString() string {
	switch bt {
	case Queue:
		return "queue"
	default:
		return "unknown"
	}
}
