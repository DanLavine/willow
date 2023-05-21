package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Join string

const (
	WhereAnd Join = "and"
	WhereOr  Join = "or"
)

type WhereClause []Where
type KeyValues map[datatypes.String]datatypes.String

// Query to use for any APIs
type Query struct {
	// required broker name to search
	BrokerName datatypes.String

	// specific matches to find
	Where WhereClause
}

// match the any tag group where all keys and values are found and eventually filtered out
type Where struct {
	// Find all values where the KeyValue Pairs Exist
	KeyValuePairs KeyValues

	// Find all values where the key exists
	KeyExists datatypes.String

	// exclude these results from the query when set to true
	Exclusion bool

	Join            *Join
	JoinWhereClause *Where
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
