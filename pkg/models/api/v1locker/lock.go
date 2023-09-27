package v1locker

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type LockRequest struct {
	KeyValues datatypes.StringMap
}

func ParseLockRequest(reader io.ReadCloser) (*LockRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &LockRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (lr *LockRequest) Validate() *api.Error {
	if len(lr.KeyValues) == 0 {
		return api.InvalidRequestBody.With("KeyValues to be provided", "recieved empty KeyValues")
	}

	return nil
}
