package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

type RuleRequest struct {
	// Name of the rule
	Name string // save this as the _associated_id in the the tree?

	// These can be used to create a rule groupiing that any tags will have to match against
	GroupBy []string // these are the logical keys to know what values we are checking against on the counters

	// Limit dictates what value of grouped counter tags to allow untill a limit is reached
	Limit uint64
}

// Used to validate on the server side that all parameters are valid. Client's can also call this
// validation beforehand to ensure that the request is valid before sending
func (rule RuleRequest) Validate() *errors.Error {
	if rule.Name == "" {
		return errors.InvalidRequestBody.With("Name to be provided", "recieved empty string")
	}

	if len(rule.GroupBy) == 0 {
		return errors.InvalidRequestBody.With("GroupBy tags to be provided", "recieved empty tag grouping")
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
func ParseRuleResponse(reader io.ReadCloser) (*RuleResponse, *errors.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.ReadResponseBodyError.With("", err.Error())
	}

	obj := &RuleResponse{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, errors.ParseResponseBodyError.With("", err.Error())
	}

	return obj, nil
}

func (ruleRes RuleResponse) ToBytes() []byte {
	data, _ := json.Marshal(ruleRes)
	return data
}
