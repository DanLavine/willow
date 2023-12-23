package v1

import (
	"encoding/json"
	"time"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// LockCreateRequest is used to request an exclusive lock
type LockCreateRequest struct {
	// KeyValues defines the collection to obtain a lock for
	KeyValues datatypes.KeyValues

	// Timeout defines how long the lock should remain valid if the client fails to heartbeat.
	// If this is set to 0, then the Server's configuration will be used.
	Timeout time.Duration
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the LockCreateRequest has all required fields set
func (req *LockCreateRequest) Validate() error {
	if len(req.KeyValues) == 0 {
		return errors.KeyValuesLenghtInvalid
	}

	if err := req.KeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the LockCreateRequest
//
// EncodeJSON encodes the model to a valid JSON format
func (req *LockCreateRequest) EncodeJSON() ([]byte, error) {
	return json.Marshal(req)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse the LockCreateRequest from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the reader
//
// DecodeJSON can easily parse the response body from an http create request
func (req *LockCreateRequest) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, req); err != nil {
		return err
	}

	return nil
}

// LockCreateResponse is the response once a client recieves the Lock
type LockCreateResponse struct {
	// SessionID is a uniquely generated ID to Heartbeat or Release a lock with.
	SessionID string

	// Timeout duration on the server till a lock is released if no Heartbeats are recieved.
	// Clients should ensure that multiple heartbeats are sent per timout to ensure network errors are accounted for
	Timeout time.Duration
}

//	RETURNS:
//	- error - error describing any possible issues with the LockCreateResponse and the steps to rectify them
//
// Validate is used to ensure that LockCreateResponse has all required fields set
func (resp *LockCreateResponse) Validate() error {
	if resp.SessionID == "" {
		return sessionIDEmpty
	}

	if resp.Timeout == 0 {
		return timeoutIsInvalid
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the LockCreateResponse
//	- error - error encoding to JSON
//
// EncodeJSON encodes the model to a valid JSON format
func (resp *LockCreateResponse) EncodeJSON() ([]byte, error) {
	return json.Marshal(resp)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse the LockCreateResponse from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the reader
//
// DecodeJSON can easily parse the response body from an http create request
func (resp *LockCreateResponse) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, resp); err != nil {
		return err
	}

	return nil
}
