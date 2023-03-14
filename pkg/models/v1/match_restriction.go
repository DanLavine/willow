package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
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
	// If BrokerMatches are provided, then these are the prefered matching restrcitions
	BrokerMatches []MatchBrokers

	// If BrokerMatchers are nil. then TagMatches are the default. If both are nil, then will find everything
	TagMatches []MatchTags
}

func (mq *MatchQuery) validate() *Error {
	if len(mq.BrokerMatches) != 0 {
		if len(mq.TagMatches) != 0 {
			return &Error{Message: "Invalid Match Query. Can only support BrokerMatches or TagMatches not both", StatusCode: http.StatusBadRequest}
		}

		for _, brokerMatch := range mq.BrokerMatches {
			if err := brokerMatch.validate(); err != nil {
				return err
			}
		}

		return nil
	}

	if len(mq.TagMatches) == 0 {
		return &Error{Message: "Invalid Match Query. Requires BrokerMatches or TagMatches to be set", StatusCode: http.StatusBadRequest}
	}

	for _, matchTags := range mq.TagMatches {
		if err := matchTags.validate(); err != nil {
			return err
		}
	}

	return nil
}

type MatchBrokers struct {
	Name      string
	MatchTags MatchTags
}

func (mb MatchBrokers) validate() *Error {
	if mb.Name == "" {
		return &Error{Message: "MatchBroker cannot have an empty Name", StatusCode: http.StatusBadRequest}
	}

	return mb.MatchTags.validate()
}

type MatchTags struct {
	MatchTagsRestrictions MatchTagsRestrictions
	Tags                  []string
}

func (mt MatchTags) validate() *Error {
	switch mt.MatchTagsRestrictions {
	case STRICT, SUBSET, ANY, ALL:
		// nothing to do here. these are valid
	default:
		return (&Error{Message: "Invalid Match Tag", StatusCode: http.StatusBadRequest}).With("[STRICT(0), SUBSET(1), ANY(2), ALL(3)]", fmt.Sprintf("%d", mt.MatchTagsRestrictions))
	}

	sort.Strings(mt.Tags)

	return nil
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
