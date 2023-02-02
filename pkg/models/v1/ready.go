package v1

import (
	"encoding/json"
	"io"
	"sort"
)

// Used to notify Willow that a client is ready to accept a message for a particular broker
type Ready struct {
	// Either Queue or PubSub messages
	BrokerType BrokerType

	// Query to match for
	MatchQuery MatchQuery
}

func ParseReadyRequest(reader io.ReadCloser) (*Ready, *Error) {
	readyRequestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	return ParseReadyQuery(readyRequestBody)
}

func ParseReadyQuery(b []byte) (*Ready, *Error) {
	readyQuery := &Ready{}
	if err := json.Unmarshal(b, readyQuery); err != nil {
		return nil, ParseRequestBodyError.With("ready query to be valid json", err.Error())
	}

	sort.Strings(readyQuery.MatchQuery.BrokerTags)

	return readyQuery, nil
}
