package v1

import (
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type BrokerType uint32

const (
	Queue BrokerType = iota
)

// Converet the broker type to a String
func (bt BrokerType) ToString() string {
	switch bt {
	case Queue:
		return "queue"
	default:
		return "unknown"
	}
}

type BrokerInfo struct {
	// specific queue name for the message
	Name string

	// Type of broker
	// NOTE: not currently used
	//BrokerType BrokerType

	// possible tags used by the broker
	// leaving this empty (or nil) results to the default tag set
	Tags datatypes.KeyValues
}

func (b BrokerInfo) validate() *errors.Error {
	if b.Name == "" {
		return errors.InvalidRequestBody.With("Name to be provided", "Name is the empty string")
	}

	if len(b.Tags) == 0 {
		return errors.InvalidRequestBody.With("Tags are empty", "Tags require at least 1 key value pair")
	}

	return nil
}
