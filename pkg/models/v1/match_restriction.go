package v1

import (
	"encoding/json"
	"io"
	"sort"
)

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

func ValidMatchRestrition(mr MatchRestriction) bool {
	switch mr {
	case STRICT, SUBSET, ANY, ALL:
		return true
	default:
		return false
	}
}

type MatchQuery struct {
	// Tag of the broker we want to read from
	MatchRestriction MatchRestriction
	BrokerTags       []string
}

func (mq *MatchQuery) MatchTags(tags []string) bool {
	if mq.MatchRestriction == ALL {
		return true
	}

	if mq.MatchRestriction == STRICT {
		if len(tags) == len(mq.BrokerTags) {
			for index, tag := range tags {
				if mq.BrokerTags[index] != tag {
					return false
				}
			}

			return true
		}

		return false
	}

	if mq.MatchRestriction == SUBSET {
		// only check tags that are guranteed to at least have all the elements
		if len(tags) >= len(mq.BrokerTags) {
			lastFound := 0
			for _, searchTag := range mq.BrokerTags {

				found := false
				for i := lastFound; i < len(tags); i++ {
					if tags[i] == searchTag {
						lastFound = i + 1 // advance next search start since things are sorted.
						found = true
						break
					}
				}

				if !found {
					return false
				}
			}

			// found all the tags,return true
			return true
		}

		return false
	}

	if mq.MatchRestriction == ANY {
		for _, searchTag := range mq.BrokerTags {
			for _, tag := range tags {
				if searchTag == tag {
					return true
				}
			}
		}
	}

	return false

}

func ParseMatchRequest(reader io.ReadCloser) (*MatchQuery, *Error) {
	matchRequestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	return ParseMatchQuery(matchRequestBody)
}

func ParseMatchQuery(b []byte) (*MatchQuery, *Error) {
	matchQuery := &MatchQuery{}
	if err := json.Unmarshal(b, matchQuery); err != nil {
		return nil, ParseRequestBodyError.With("match query to be valid json", err.Error())
	}

	sort.Strings(matchQuery.BrokerTags)

	return matchQuery, nil
}
