package v1

import (
	"encoding/json"
	"io"
	"time"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
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
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (lock *Lock) EncodeJSON() []byte {
	data, _ := json.Marshal(lock)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the reader stream
//	- reader - stream to read the encoded Lock data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response or Validating the Lock
//
// Decode can easily parse the response body from an http create request
func (lock *Lock) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, lock); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return nil
}

// LockQueryResponse show any locks that match the  AssociatedQuery.
// In the future, this will be an ADMIN API as the SessionID should be hiden from
// any malicious actors. But for now, it stands a useful debug API to inspect the
// state of the world.
type LockQueryResponse struct {
	Locks []*Lock
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the LockQueryResponse
//
// EncodeJSON encodes the model to a valid JSON format
func (lockQueryRespinse *LockQueryResponse) EncodeJSON() []byte {
	data, _ := json.Marshal(lockQueryRespinse)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the reader stream
//	- reader - stream to read the encoded LockQueryResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the stream or LockQueryResponse is invalid
//
// Decode can easily parse the response body from an http create request
func (lockQueryResponse *LockQueryResponse) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, lockQueryResponse); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return nil
}
