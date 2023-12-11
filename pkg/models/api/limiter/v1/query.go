package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type Query struct {
	AssociatedKeyValues datatypes.AssociatedKeyValuesQuery
}

// Server side logic to parse a Rule to know it is valid
func ParseGeneralQuery(reader io.ReadCloser) (*Query, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := Query{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return &obj, nil
}

// Used to validate on the server side that all parameters are valid. Client's can also call this
// validation beforehand to ensure that the request is valid before sending
func (q Query) Validate() *api.Error {
	if err := q.AssociatedKeyValues.Validate(); err != nil {
		return api.InvalidRequestBody.With("Query to be valid.", err.Error())
	}

	return nil
}

func (q Query) ToBytes() []byte {
	data, _ := json.Marshal(q)
	return data
}
