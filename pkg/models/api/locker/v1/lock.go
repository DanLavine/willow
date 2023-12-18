package v1locker

import (
	"encoding/json"
	"time"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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

func (lr *CreateLockRequest) Validate() *errors.Error {
	if len(lr.KeyValues) == 0 {
		return errors.InvalidRequestBody.With("KeyValues to be provided", "recieved empty KeyValues")
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
