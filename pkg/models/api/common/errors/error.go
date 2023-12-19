package errors

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
)

// Error can be used to parse errors returned from services if the request was invalid
type Error struct {
	Message string
}

// Validate the error message. Currently a NO-OP
func (err *Error) Validate() error {
	return nil
}

func (err *Error) EncodeJSON() []byte {
	data, _ := json.Marshal(err)
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
func (e *Error) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		body, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read error's response stream from the server: %w", err)
		}

		if err := json.Unmarshal(body, e); err != nil {
			return fmt.Errorf("failed to decode error from the server: %w", err)
		}
	default:
		return fmt.Errorf("recieved unknown content type for an error from the server: %s", contentType)
	}

	return nil
}

//	RETURNS:
//	- string - error message
//
// Error returns the original message sent from the service
func (e *Error) Error() string {
	return e.Message
}
