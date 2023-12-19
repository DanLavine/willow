package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// DSL TODO: This isn't useful at all. the client already has the sessionID. so this shouldjust be the common error. DELETE THIS FILE!

// Hearbeat error is returned to Client's as part of the heartbeat process when an error occurs server side.
// This could be any number of cases:
//  1. Client has lost the lock and did not time out localy
//  2. SessionID does not exist because the client state was corrupted
type HeartbeatError struct {
	Session string `json:"Session"`
	Error   string `json:"Error"`
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (heartbeatError *HeartbeatError) EncodeJSON() []byte {
	data, _ := json.Marshal(heartbeatError)
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
func (heartbeatError *HeartbeatError) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, heartbeatError); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return nil
}
