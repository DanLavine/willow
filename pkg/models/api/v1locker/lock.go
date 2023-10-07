package v1locker

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// create request
type CreateLockRequest struct {
	KeyValues datatypes.StringMap
}
type CreateLockResponse struct {
	SessionID string
}

// List request
type Lock struct {
	KeyValues          datatypes.StringMap
	LocksHeldOrWaiting int
}
type ListLockResponse struct {
	Locks []Lock
}

// Delete request

func NewListLockResponse(listLocks []Lock) ListLockResponse {
	return ListLockResponse{
		Locks: listLocks,
	}
}

func ParseLockRequest(reader io.ReadCloser) (*CreateLockRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &CreateLockRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return obj, nil
}

func (lr *CreateLockRequest) Validate() *api.Error {
	if len(lr.KeyValues) == 0 {
		return api.InvalidRequestBody.With("KeyValues to be provided", "recieved empty KeyValues")
	}

	return nil
}

func (lr CreateLockResponse) ToBytes() []byte {
	data, _ := json.Marshal(lr)
	return data
}

func (lr ListLockResponse) ToBytes() []byte {
	data, _ := json.Marshal(lr)
	return data
}
