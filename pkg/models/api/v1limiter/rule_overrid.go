package v1limiter

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type RuleOverrideRequest struct {
	// They key value parings we are making the override for
	// NOTE: these must match the GroupBy keys in the original Rule
	KeyValues datatypes.KeyValues

	// The new limit to use for the paricular mapping
	Limit uint64
}

func ParesRuleOverrideRequst(reader io.ReadCloser) (*RuleOverrideRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleOverrideRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (ro *RuleOverrideRequest) Validate() *api.Error {
	if len(ro.KeyValues) == 0 {
		return api.InvalidRequestBody.With("KeyValues tags to be provided", "recieved empty set")
	}

	return nil
}
