package v1locker

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

// heartbeat request
type HeartbeatLocksRequst struct {
	LockSessions []string
}

type HeartbeatError struct {
	Session string
	Error   string
}

type HeartbeatLocksResponse struct {
	HeartbeatErrors []HeartbeatError
}

func NewHeartbeatLocksResponse(errors []HeartbeatError) HeartbeatLocksResponse {
	return HeartbeatLocksResponse{
		HeartbeatErrors: errors,
	}
}

func ParseHeartbeatRequest(reader io.ReadCloser) (*HeartbeatLocksRequst, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &HeartbeatLocksRequst{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (hlr *HeartbeatLocksRequst) Validate() *api.Error {
	if len(hlr.LockSessions) == 0 {
		return api.InvalidRequestBody.With("LockSessions to be provided", "recieved empty LockSessions")
	}

	for index, sessionID := range hlr.LockSessions {
		if sessionID == "" {
			return api.InvalidRequestBody.With(fmt.Sprintf("LockSessions at index '%d' to not be an empty string", index), "received an empty string")
		}
	}

	return nil
}

func (hlr HeartbeatLocksResponse) ToBytes() []byte {
	data, _ := json.Marshal(hlr)
	return data
}
