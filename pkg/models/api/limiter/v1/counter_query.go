package v1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// used with list
type CounterResponse struct {
	KeyValues datatypes.KeyValues

	Counters uint64
}

func (counterResp CounterResponse) Validate() error {
	if len(counterResp.KeyValues) == 0 {
		return fmt.Errorf("'KeyValues' requres at least 1 key + value piar")
	}

	if err := counterResp.KeyValues.Validate(); err != nil {
		return err
	}

	if counterResp.Counters == 0 {
		return fmt.Errorf("'Counters' is set to 0, which is invalid. The counter should be deleted")
	}

	return nil
}

type CountersResponse struct {
	Counters []*CounterResponse
}

//	RETURNS:
//	- error - any errors encountered with the response object
//
// Validate is used to ensure that CreateLockResponse has all required fields set
func (resp CountersResponse) Validate() error {
	if len(resp.Counters) == 0 {
		return nil
	}

	for index, counter := range resp.Counters {
		if counter == nil {
			return fmt.Errorf("error at Counters index: %d: value cannot be nil", index)
		}

		if err := counter.Validate(); err != nil {
			return fmt.Errorf("error at counters index %d: %w", index, err)
		}
	}

	return nil
}

//	RETURNS:
//	- []byte - byte array that can be sent over an http client
//
// EncodeJSON encodes the model to a valid JSON format
func (resp *CountersResponse) EncodeJSON() []byte {
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
func (resp *CountersResponse) Decode(contentType api.ContentType, reader io.ReadCloser) error {
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
