package v1

import (
	"encoding/json"
	"fmt"
)

// LockClaim is used to refresh a heartbeat and eventually reclaim a lock when a service restarts
type LockClaim struct {
	// SessionID for the currently held lock
	SessionID string
}

//	RETURNS:
//	- error - error describing any possible issues and the steps to rectify them
//
// Validate ensures the LockCreateRequest has all required fields set
func (claim *LockClaim) Validate() error {
	if claim.SessionID == "" {
		return fmt.Errorf("'SessionID' is required, but received an empty string")
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the LockCreateRequest
//
// EncodeJSON encodes the model to a valid JSON format
func (claim *LockClaim) EncodeJSON() ([]byte, error) {
	return json.Marshal(claim)
}

//	PARAMETERS:
//	- data - encoded JSON data to parse the LockCreateRequest from
//
//	RETURNS:
//	- error - any error encoutered when reading or parsing the reader
//
// DecodeJSON can easily parse the response body from an http create request
func (claim *LockClaim) DecodeJSON(data []byte) error {
	if err := json.Unmarshal(data, claim); err != nil {
		return err
	}

	return nil
}
