package v1limiter

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

type RuleUpdateRequest struct {
	// TODO: remove me
	Name string

	// Limit Key is an optional param that can be used to dictate what value of the tags to use as a limiter
	Limit uint64
}

func ParseRuleUpdateRequest(reader io.ReadCloser) (*RuleUpdateRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleUpdateRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	return obj, nil
}
