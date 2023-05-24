package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Join string

const (
	WhereAnd Join = "and" // higher precedence over or
	WhereOr  Join = "or"  // lower precedence over and
)

type KeyValues map[datatypes.String]datatypes.String

// Query to use for any APIs
type QueryInclusive struct {
	// required broker name to search
	BrokerName datatypes.String

	// specific matches to find
	Where *JoinedInclusiveWhereClause
}

type JoinedInclusiveWhereClause struct {
	Where InclusiveWhereClause

	Join            *Join
	JoinWhereClause *JoinedInclusiveWhereClause
}

// match the any tag group where all keys and values are found and eventually filtered out
type InclusiveWhereClause struct {
	// only one of these can be provided
	//// All key values must equal the provided key values only
	EqualsKeyValues KeyValues

	//// Any tag groups that match the provided key values
	MatchesKeyValues KeyValues

	//// Any tag groups that contains the requested keys
	ContainsKeys datatypes.Strings

	// optional join clause
	Join            *Join
	JoinWhereClause *InclusiveWhereClause
}

func ParseQueryInclusive(reader io.ReadCloser) (*QueryInclusive, *Error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, InvalidRequestBody.With("", err.Error())
	}

	query := &QueryInclusive{}
	if err := json.Unmarshal(body, query); err != nil {
		return nil, ParseRequestBodyError.With("query to be valid json", err.Error())
	}

	//if err := query.validate(); err != nil {
	//	return nil, err
	//}

	return query, nil
}

func (q *QueryInclusive) validate() *Error {
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
