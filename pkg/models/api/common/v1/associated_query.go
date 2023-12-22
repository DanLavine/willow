package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// AssociatedQuery is used to query a number of various apis for the
// Willow services. Since most data is saved throught the common KeyValues,
// for Willow's queues, Limiter's Rule Overrides, Limiter's Counters and Locker's Locks.
type AssociatedQuery struct {
	// Query for the KeyValues that defined the various API Models
	AssociatedKeyValues datatypes.AssociatedKeyValuesQuery
}

//	RETURNS:
//	- error - error describing any possible issues with the query and the steps to rectify them
//
// Validate ensures the CreateLockRequest has all required fields set
func (query *AssociatedQuery) Validate() error {
	if err := query.AssociatedKeyValues.Validate(); err != nil {
		return err
	}

	return nil
}

//	RETURNS:
//	- []byte - encoded JSON byte array for the AssociatedQuery
//
// EncodeJSON encodes the model to a valid JSON format
func (query *AssociatedQuery) EncodeJSON() []byte {
	data, _ := json.Marshal(query)
	return data
}

//	PARAMETERS:
//	- contentType - how to interperate the stream
//	- reader - stream to read the encoded AssociatedQuery data from
//
//	RETURNS:
//	- error - any error encoutered when reading the stream or AssociatedQuery is invalid
//
// Decode can easily parse the response body from an http request
func (query *AssociatedQuery) Decode(contentType api.ContentType, reader io.ReadCloser) error {
	switch contentType {
	case api.ContentTypeJSON:
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			return errors.FailedToReadStreamBody(err)
		}

		if err := json.Unmarshal(requestBody, query); err != nil {
			return errors.FailedToDecodeBody(err)
		}
	default:
		return errors.UnknownContentType(contentType)
	}

	return query.Validate()
}
