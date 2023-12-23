package v1

import (
	"encoding/json"
	"time"

	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// Lock is the full representation of the Lock Object
type Lock struct {
	// SessionID associated with the lock clients are using to heartbeat or release a lock with
	SessionID string

	// KeyValues collection that deffines the lock
	KeyValues datatypes.KeyValues

	// Timeout for the lock
	Timeout time.Duration
	// Time until the lock expires if no heartbeats are received
	TimeTillExipre time.Duration

	// LocksHeldOrWaiting show how many clients are all trying to obtaiin the same lock
	LocksHeldOrWaiting int
}

//	RETURNS:
//	- error - error describing any possible issues with the Lock and the steps to rectify them
//
// Validate ensures the Lock has all required fields set
func (lock *Lock) Validate() error {
	if lock.SessionID == "" {
		return sessionIDEmpty
	}

	if len(lock.KeyValues) == 0 {
		return keyValuesLenghtInvalid
	}

	if err := lock.KeyValues.Validate(); err != nil {
		return err
	}

	if lock.Timeout == 0 {
		return timeoutIsInvalid
	}

	if lock.LocksHeldOrWaiting == 0 {
		return locksHeldOrWaitingIsInvalid
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the Lock
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (lock *Lock) EncodeJSON() ([]byte, error) {
	return json.Marshal(lock)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse the Lock from
//
//	RETURNS:
//	- error - any error encoutered when reading the response or Validating the Lock
//
// Decode can easily parse the response body from an http create request
func (lock *Lock) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, lock); err != nil {
		return err
	}

	return nil
}
