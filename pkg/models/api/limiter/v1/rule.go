package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

type RuleRequest struct {
	// Name of the rule
	Name string // save this as the _associated_id in the the tree?

	// These can be used to create a rule groupiing that any tags will have to match against
	GroupBy []string // these are the logical keys to know what values we are checking against on the counters

	// Limit dictates what value of grouped counter tags to allow untill a limit is reached
	Limit uint64
}

// Server side logic to parse a Rule to know it is valid
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

// Used to validate on the server side that all parameters are valid. Client's can also call this
// validation beforehand to ensure that the request is valid before sending
func (rule RuleRequest) Validate() *api.Error {
	if rule.Name == "" {
		return api.InvalidRequestBody.With("Name to be provided", "recieved empty string")
	}

	if len(rule.GroupBy) == 0 {
		return api.InvalidRequestBody.With("GroupBy tags to be provided", "recieved empty tag grouping")
	}

	return nil
}

func (ruleReq RuleRequest) ToBytes() []byte {
	data, _ := json.Marshal(ruleReq)
	return data
}

type RuleResponse struct {
	// Name of the rule
	Name string // save this as the _associated_id in the the tree?

	// These can be used to create a rule groupiing that any tags will have to match against
	GroupBy []string // these are the logical keys to know what values we are checking against on the counters

	// Limit dictates what value of grouped counter tags to allow untill a limit is reached
	Limit uint64

	// This is a "Read Only" parameter and will be ignored on create operations
	Overrides []Override
}

// Client side logic to parse a Rule
func ParseRuleResponse(reader io.ReadCloser) (*RuleResponse, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleResponse{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	return obj, nil
}

func (ruleRes RuleResponse) ToBytes() []byte {
	data, _ := json.Marshal(ruleRes)
	return data
}
