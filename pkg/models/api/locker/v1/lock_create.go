package v1

import (
	"fmt"
	"time"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// LockCreateRequest is used to request an exclusive lock
type LockCreateRequest struct {
	// KeyValues defines the collection to obtain a lock for
	KeyValues datatypes.KeyValues

	// LockTimeout defines how long the lock should remain valid if the client fails to heartbeat.
	// If this is set to 0, then the Server's configuration will be used.
	LockTimeout time.Duration
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the LockCreateRequest has all required fields set
func (req *LockCreateRequest) Validate() error {
	if err := req.KeyValues.Validate(datatypes.MinDataType, datatypes.MaxWithoutAnyDataType); err != nil {
		return &errors.ModelError{Field: "KeyValues", Child: err.(*errors.ModelError)}
	}

	return nil
}

// LockCreateResponse is the response once a client recieves the Lock
type LockCreateResponse struct {
	// LockID is a consistent ID for all clients that trry to obtain a lock and have to wait. If the
	// lock is completely released, then the service will generate a new LockID on the next request creating the lock
	LockID string

	// SessionID is a uniquely generated ID to Heartbeat or Release a lock with.
	SessionID string

	// LockTimeout duration on the server till a lock is released if no Heartbeats are recieved.
	// Clients should ensure that multiple heartbeats are sent per timout to ensure network errors are accounted for
	LockTimeout time.Duration
}

//	RETURNS:
//	- error - error describing any possible issues with the LockCreateResponse and the steps to rectify them
//
// Validate is used to ensure that LockCreateResponse has all required fields set
func (resp *LockCreateResponse) Validate() error {
	if resp.SessionID == "" {
		return &errors.ModelError{Field: "SessionID", Err: fmt.Errorf("received an empty string")}
	}

	if resp.LockTimeout == 0 {
		return &errors.ModelError{Field: "Timeout", Err: fmt.Errorf("requires a value greater than 0")}
	}

	return nil
}
