package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
)

// Single rule
type RuleResponse struct {
	// Name of the rule
	Name string // save this as the _associated_id in the the tree?

	// These can be used to create a rule groupiing that any tags will have to match against
	GroupBy []string // these are the logical keys to know what values we are checking against on the counters

	// Limit dictates what value of grouped counter tags to allow untill a limit is reached
	Limit uint64

	// This is a "Read Only" parameter and will be ignored on create operations
	Overrides []Override
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (resp *RuleResponse) Validate() error {
	if resp.Name == "" {
		return fmt.Errorf("'Name' is the empty string")
	}

	if len(resp.GroupBy) == 0 {
		return fmt.Errorf("'GroupBy' tags requres at least 1 Key")
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (resp *RuleResponse) EncodeJSON() []byte {
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
func (resp *RuleResponse) Decode(contentType api.ContentType, reader io.ReadCloser) error {
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
