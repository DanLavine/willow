package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// TODO: there is no way to query for the "default" queue using things. I think something is missing still
//       should the api require tags? then it won't be a problem. As it currently stands, everything is placed into
//       the tree with a "default" tag (single level) and doesn't vhave a value... might not be the best

// Query to use for any APIs
type Query struct {
	// required broker name to search
	BrokerName datatypes.String

	// specific matches to find
	Matches Matches
}

// When performing a queery, only 1 will be used
type Matches struct {
	// select the all tag groups (aka global)
	All bool

	// select only 1 item from the tag group iff all key + values match
	StrictMatches KeyValues

	// select N items from the tag group that match the query parameters
	GeneralMatches *GeneralMatches
}

type KeyValues map[datatypes.String]datatypes.String

// match the any tag group where all keys and values are found and eventually filtered out
type GeneralMatches struct {
	// Match any value that contain a key
	ContainsKeys datatypes.Strings

	// ignore any values that contain a provided key
	NotContainsKeys datatypes.Strings

	// Any potential value must contain all possible keys value pairs
	EqualKeyValues KeyValues

	// Any potential values will be removed if these rules match
	NotEqualKeyvalues KeyValues
}

func ParseQuery(reader io.ReadCloser) (*Query, *Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	query := &Query{}
	if err := json.Unmarshal(body, query); err != nil {
		return nil, ParseRequestBodyError.With("query to be valid json", err.Error())
	}

	if err := query.validate(); err != nil {
		return nil, err
	}

	return query, nil
}

func (q Query) validate() *Error {
	if q.BrokerName == "" {
		return &Error{Message: "BrokerName cannot be empty", StatusCode: http.StatusBadRequest}
	}

	if q.Matches.All == true {
		if q.Matches.StrictMatches != nil || q.Matches.GeneralMatches != nil {
			return &Error{Message: "Invalid match query. Can only use one of [All, StrictMatches, GeneralMatches]", StatusCode: http.StatusBadRequest}
		}

		// valid global select case
		return nil
	}

	if q.Matches.StrictMatches != nil {
		if q.Matches.GeneralMatches != nil {
			return &Error{Message: "Invalid match query. Can only use one of [All, StrictMatches, GeneralMatches]", StatusCode: http.StatusBadRequest}
		}

		if len(q.Matches.StrictMatches) == 0 {
			return &Error{Message: "StrictMatches requires at least one tag pair to search for", StatusCode: http.StatusBadRequest}
		}

		return nil
	}

	return q.Matches.GeneralMatches.validate()
}

func (gm *GeneralMatches) validate() *Error {
	if len(gm.ContainsKeys) == 0 && len(gm.EqualKeyValues) == 0 {
		return &Error{Message: "General Matches requires at least one Contains or Equals query parameter", StatusCode: http.StatusBadRequest}
	}

	return nil
}
