package v1limiter

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/query"
)

type RuleRequest struct {
	// Name of the rule
	Name string

	// These can be used to create a rule groupiing that any tags will have to match agains
	GroupBy []string

	// When comparing tags, use this selection to figure out if a rule applies to them
	Query query.AssociatedKeyValuesQuery

	// Limit Key is an optional param that can be used to dictate what value of the tags to use as a limiter
	Limit uint64
}

func ParseRuleRequest(reader io.ReadCloser) (*RuleRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (rreq *RuleRequest) Validate() *api.Error {
	if rreq.Name == "" {
		return api.InvalidRequestBody.With("Name to be provided", "recieved empty string")
	}

	if len(rreq.GroupBy) == 0 {
		return api.InvalidRequestBody.With("GroupBy tags to be provided", "recieved empty tag grouping")
	}

	if err := rreq.Query.Validate(); err != nil {
		return api.InvalidRequestBody.With("Selection query to be a valid expression", err.Error())
	}

	return nil
}
