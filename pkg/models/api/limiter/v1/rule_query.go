package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Rule query is used to find any Rules that match the provided KeyValues
type RuleQuery struct {
	// Find any rules that match the provided key values
	KeyValues datatypes.KeyValues

	// If true, will include any matching override values.
	// If false, will just find the top most rules that match the key values and or name
	IncludeOverrides bool
}

// Server side logic to parse a Rule to know it is valid
func ParseRuleQuery(reader io.ReadCloser) (*RuleQuery, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleQuery{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.ValidateRequest(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

// Used to validate on the server side that all parameters are valid. Client's can also call this
// validation beforehand to ensure that the request is valid before sending
func (rq *RuleQuery) ValidateRequest() *api.Error {
	if err := rq.KeyValues.Validate(); err != nil {
		return api.InvalidRequestBody.With("KeyValues to be valid", err.Error())
	}

	return nil
}
