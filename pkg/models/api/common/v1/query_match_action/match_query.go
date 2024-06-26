package querymatchaction

import (
	"fmt"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

// MatchKeyValuesQuery is used to match any permutations of the KeyValues against items save
// in the associated DBs. To succesfully match an item in the DB, all the DB.KeyValue pairs
// must be part of the MatchKeyValuesQuery.KeyValues
type MatchActionQuery struct {
	// KeyValues to match against
	KeyValues MatchKeyValues

	// Minimum number of keys that make up a permutation
	MinNumberOfPermutationKeyValues *int `json:"MinNumberOfPermutationKeyValues,omitempty"`
	// Maximum number of keys that make up a permutation
	MaxNumberOfPermutationKeyValues *int `json:"MaxNumberOfPermutationKeyValues,omitempty"`
}

//	RETURNS:
//	- error - error describing any possible issues with the query and the steps to rectify them
//
// Validate ensures the CreateLockRequest has all required fields set
func (matchActionQuery *MatchActionQuery) Validate() *errors.ModelError {
	if len(matchActionQuery.KeyValues) == 0 {
		return &errors.ModelError{Field: "KeyValues", Err: fmt.Errorf("requires a length of at least 1, but recevied 0")}
	} else {
		if err := matchActionQuery.KeyValues.Validate(); err != nil {
			return &errors.ModelError{Field: "KeyValues", Child: err}
		}
	}

	return nil
}

func KeyValuesToAnyMatchActionQuery(keyValues datatypes.KeyValues) *MatchActionQuery {
	matchActionQuery := MatchActionQuery{
		KeyValues: MatchKeyValues{},
	}

	for key, value := range keyValues {
		matchActionQuery.KeyValues[key] = MatchValue{
			Value: value,
			TypeRestrictions: v1.TypeRestrictions{
				MinDataType: value.Type,
				MaxDataType: datatypes.T_any,
			},
		}
	}

	return &matchActionQuery
}
