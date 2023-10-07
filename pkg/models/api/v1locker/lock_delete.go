package v1locker

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

type DeleteLockRequest struct {
	SessionID string
}

func ParseDeleteLockRequest(reader io.ReadCloser) (*DeleteLockRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &DeleteLockRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (dlr *DeleteLockRequest) Validate() *api.Error {
	if dlr.SessionID == "" {
		return api.InvalidRequestBody.With("SessionID to be provided", "recieved empty string")
	}

	return nil
}

func (dlr *DeleteLockRequest) ToBytes() []byte {
	data, _ := json.Marshal(dlr)
	return data
}
