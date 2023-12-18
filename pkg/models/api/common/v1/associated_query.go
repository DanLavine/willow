package v1

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

type AssociatedQuery struct {
	AssociatedKeyValues datatypes.AssociatedKeyValuesQuery
}

// Used to validate on the server side that all parameters are valid
func (q AssociatedQuery) Validate() *errors.Error {
	if err := q.AssociatedKeyValues.Validate(); err != nil {
		return &errors.Error{Message: err.Error()}
	}

	return nil
}

func (q AssociatedQuery) ToBytes() []byte {
	data, _ := json.Marshal(q)
	return data
}
