package v1limiter

import (
	"encoding/json"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
)

type FindRuleRequset struct {
	// name of the specific rule to find
	Name string // optional

	// Find any rules that match the provided key values
	KeyValues datatypes.KeyValues // optional

	// If true, will include any matching override values.
	// If false, will just find the top most rules that match the key values and or name
	IncludeOverrides bool
}

// type FindRuleRsponse strut {
type FindRuleResponse []RuleResponse

type RuleResponse struct {
	// Name of the rule
	Name string

	// These can be used to create a rule groupiing that any tags will have to match agains
	GroupBy []string

	// When comparing tags, use this selection to figure out if a rule applies to them
	Query query.AssociatedKeyValuesQuery

	// Limit Key is an optional param that can be used to dictate what value of the tags to use as a limiter
	Limit uint64

	// all the rule overrides
	RuleOverrides []RuleOverrideResponse
}

type RuleOverrideResponse struct {
	// The possible key values for the overrides
	KeyValues datatypes.KeyValues

	// The new limit to use for the paricular mapping
	Limit uint64
}

func (fresp *FindRuleResponse) ToBytes() ([]byte, error) {
	return json.Marshal(fresp)
}
