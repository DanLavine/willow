package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

type Rules []Rule

// Client side helper to parse all rules recieved fom the server
func ParseRulesResponse(reader io.ReadCloser) (Rules, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := Rules{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	return obj, nil
}
