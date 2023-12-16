package v1locker

import (
	"encoding/json"
	"io"
	"time"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// create request
type CreateLockRequest struct {
	// the key values to create a lock for
	KeyValues datatypes.KeyValues

	// how long it takes for the lock to timeout
	// heartbeats should send 3 requests per timeout just to acount for network disruptions
	//
	// Open Question. Should there be a min Timeout?
	Timeout time.Duration
}
type CreateLockResponse struct {
	// The session to Delete or heartbeat a lock with
	SessionID string

	// How long it takes for a lock to timeout. Heartbeats should be done 3x per timeout to account occassional network issues
	Timeout time.Duration
}

// List request
type Lock struct {
	SessionID          string
	KeyValues          datatypes.KeyValues
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

func ParseLockRequest(reader io.ReadCloser, defaultTimeout time.Duration) (*CreateLockRequest, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := &CreateLockRequest{}
	if err := json.Unmarshal(requestBody, obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if obj.Timeout == 0 {
		obj.Timeout = defaultTimeout
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
