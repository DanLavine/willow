package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type MatchType string

const (
	Strict MatchType = "strict"
	Subset MatchType = "subset"
	Any    MatchType = "any"
	All    MatchType = "all"
)

type Join string

const (
	And Join = "and"
	Or  Join = "or"
)

type KeyValues map[datatypes.String]datatypes.String

// Query to use for any APIs
type Query struct {
	// required broker name to search
	BrokerName datatypes.String

	// specific matches to find
	Matches Matches
}

// When performing a queery, only 1 will be used
type Matches struct {
	// What type we are searching for
	Type MatchType

	// Where Clauses are a grouping of clases that will record results of any being true
	Where *[]WhereClause
}

// match the any tag group where all keys and values are found and eventually filtered out
type WhereClause struct {
	KeyValuePairs KeyValues

	KeyExists datatypes.String

	// exclude these results from the query when set to true
	Exclusion bool

	Join            *Join
	JoinWhereClause *WhereClause
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

	//if err := query.validate(); err != nil {
	//	return nil, err
	//}

	return query, nil
}

func (q *Query) validate() *Error {
	if q.BrokerName == "" {
		return &Error{Message: "BrokerName cannot be empty", StatusCode: http.StatusBadRequest}
	}

	//if q.Matches.All == true {
	//	if q.Matches.StrictMatches != nil || q.Matches.GeneralMatches != nil {
	//		return &Error{Message: "Invalid match query. Can only use one of [All, StrictMatches, GeneralMatches]", StatusCode: http.StatusBadRequest}
	//	}

	//	// valid global select case
	//	return nil
	//}

	//if q.Matches.StrictMatches != nil {
	//	if q.Matches.GeneralMatches != nil {
	//		return &Error{Message: "Invalid match query. Can only use one of [All, StrictMatches, GeneralMatches]", StatusCode: http.StatusBadRequest}
	//	}

	//	if len(q.Matches.StrictMatches) == 0 {
	//		return &Error{Message: "StrictMatches requires at least one tag pair to search for", StatusCode: http.StatusBadRequest}
	//	}

	//	return nil
	//}

	//return q.Matches.GeneralMatches.validate()
	return nil
}
