package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

type Rules []*RuleResponse

// Client side helper to parse all rules recieved fom the server
func ParseRulesResponse(reader io.ReadCloser) (Rules, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadResponseBodyError.With("", err.Error())
	}

	obj := Rules{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseResponseBodyError.With("", err.Error())
	}

	return obj, nil
}

func (rules Rules) ToBytes() []byte {
	data, _ := json.Marshal(rules)
	return data
}
