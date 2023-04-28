package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type MatchTagsRestrictions int

const (
	// Must be matched exactly
	STRICT MatchTagsRestrictions = iota

	// If the Broker's Tags contain all requested tags -> true
	SUBSET

	// If any tags requested are in the Broker's Tags -> true
	ANY

	// ignore the brokers tags. Any enqued message is valid
	ALL
)

// Match Query can be used to match any number of brokers with a subset of tags.
// Or all brokers and any subset of tags.
type MatchQuery struct {
	// name of the broker to chose from. Right now it is always a Queue
	BrokerName datatypes.String

	// eventually this will be useful
	//BrokerType BrokerType

	// Tags to match against
	MatchTagsRestrictions MatchTagsRestrictions
	Tags                  datatypes.Strings
}

func ParseMatchQueryRequest(reader io.ReadCloser) (*MatchQuery, *Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	matchQuery := &MatchQuery{}
	if err := json.Unmarshal(body, matchQuery); err != nil {
		return nil, ParseRequestBodyError.With("match query to be valid json", err.Error())
	}

	if err := matchQuery.validate(); err != nil {
		return nil, err
	}

	return matchQuery, nil
}

func (mq *MatchQuery) validate() *Error {
	if mq.BrokerName == "" {
		return &Error{Message: "BrokerName is a required request parameter", StatusCode: http.StatusBadRequest}
	}

	switch mq.MatchTagsRestrictions {
	case STRICT, SUBSET, ANY, ALL:
		// nothing to do here. these are valid
	default:
		return (&Error{Message: "Invalid Match Tag", StatusCode: http.StatusBadRequest}).With("[STRICT(0), SUBSET(1), ANY(2), ALL(3)]", fmt.Sprintf("%d", mq.MatchTagsRestrictions))
	}

	mq.Tags.Sort()
	return nil
}
