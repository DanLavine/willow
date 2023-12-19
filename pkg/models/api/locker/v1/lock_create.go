package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// CreateLockRequest is used to request an exclusive lock
type CreateLockRequest struct {
	// KeyValues defines the collection to obtain a lock for
	KeyValues datatypes.KeyValues

	// Timeout defines how long the lock should remain valid if the client fails to heartbeat
	Timeout time.Duration
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the CreateLockRequest has all required fields set
func (req *CreateLockRequest) Validate() error {
	if len(req.KeyValues) == 0 {
		return fmt.Errorf("'KeValues' is empty, but requires a length of at least 1")
	}

	if err := req.KeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (req *CreateLockRequest) EncodeJSON() []byte {
	data, _ := json.Marshal(req)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream. Valida values [application/json]
//	- reader - stream to read the encoded CreateLockResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// Decode can easily parse the response body from an http create request
func (req *CreateLockRequest) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, req); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return req.Validate()
}

// CreateLockResponse is the response once a client recieves the Lock
type CreateLockResponse struct {
	// SessionID is a uniquely generated ID to Heartbeat or Release a lock with
	// This way, malicious clients cannot easily release locks held by other clients
	SessionID string

	// Timeout duration on the server till a lock is released if no Heartbeats are recieved.
	// Clients should ensure that multiple heartbeats are sent per timout to ensure network errors are accounted for
	Timeout time.Duration
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (resp *CreateLockResponse) Validate() error {
	if resp.SessionID == "" {
		return fmt.Errorf("'SessionID' is the empty string")
	}

	if resp.Timeout == 0 {
		return fmt.Errorf("'Timeout' is set to 0")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (resp *CreateLockResponse) EncodeJSON() []byte {
	data, _ := json.Marshal(resp)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream. Valida values [application/json]
//	- reader - stream to read the encoded CreateLockResponse data from
//
//	RETURNS:
//	- error - any error encoutered when reading the response
//
// Decode can easily parse the response body from an http create request
func (resp *CreateLockResponse) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, resp); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return resp.Validate()
}
