package v1limiter

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type RuleOverride struct {
	// name of the rule we are creating an override for
	RuleName string

	// They key value parings we are making the override for
	// NOTE: these must match the GroupBy keys in the original Rule
	KeyValues datatypes.KeyValues

	// The new limit to use for the paricular mapping
	Limit uint64
}

func ParesRuleOverride(reader io.ReadCloser) (*RuleOverride, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleOverride{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (ro *RuleOverride) Validate() *api.Error {
	if ro.RuleName == "" {
		return api.InvalidRequestBody.With("RuleName to be provided", "recieved empty string")
	}

	if len(ro.KeyValues) == 0 {
		return api.InvalidRequestBody.With("KeyValues tags to be provided", "recieved empty set")
	}

	return nil
}
