package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

type RuleUpdate struct {
	// update the limits for a rule
	Limit uint64
}

// Server side logic to parse a Rule to know it is valid
func ParseRuleUpdateRequest(reader io.ReadCloser) (*RuleUpdate, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &RuleUpdate{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	return obj, nil
}

// Convert the RuleUpdate logic to bytes that both the Client and Server understand
func (ru *RuleUpdate) ToBytes() []byte {
	data, _ := json.Marshal(ru)
	return data
}
