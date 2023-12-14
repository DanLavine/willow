package v1

import (
	"encoding/json"
	"io"

	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// DSL TODO:
//
// I think adding pagination/limits/order by makes sense here rather than on the datatypes.AssociatedKeyValuesQuery
// But I have no need for those values at the moment, so for now I am ignoring that feature set
type GeneralAssociatedQuery struct {
	AssociatedKeyValues datatypes.AssociatedKeyValuesQuery
}

// Server side logic to parse a Rule to know it is valid
func ParseGeneralAssociatedQuery(reader io.ReadCloser) (*GeneralAssociatedQuery, *api.Error) {
	requestBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, api.ReadRequestBodyError.With("", err.Error())
	}

	obj := GeneralAssociatedQuery{}
	if err := json.Unmarshal(requestBody, &obj); err != nil {
		return nil, api.ParseRequestBodyError.With("", err.Error())
	}

	if validateErr := obj.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return &obj, nil
}

// Used to validate on the server side that all parameters are valid
func (q GeneralAssociatedQuery) Validate() *api.Error {
	if err := q.AssociatedKeyValues.Validate(); err != nil {
		return api.InvalidRequestBody.With("Query to be valid.", err.Error())
	}

	return nil
}

func (q GeneralAssociatedQuery) ToBytes() []byte {
	data, _ := json.Marshal(q)
	return data
}
