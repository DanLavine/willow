package v1limiter

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type RuleCounterRequest struct {
	// Specific key values to add or remove a counter from
	KeyValues datatypes.StringMap
}

func ParseRuleCounterRequest(reader io.ReadCloser) (*RuleCounterRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleCounterRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (rcr *RuleCounterRequest) Validate() *api.Error {
	if len(rcr.KeyValues) == 0 {
		return api.InvalidRequestBody.With("KeyValues to be provided", "recieved empty KeyValues set")
	}

	return nil
}
