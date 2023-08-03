package v1

import (
	"github.com/DanLavine/willow/pkg/models/query"
)

type LimiterRuleCreateRequest struct {
	// Name of the rule
	Name string

	// Optional parameter to apply a rule set to a specific broker
	BrokerName *string

	// These can be used to create a rule groupiing that any tags will have to match agains
	GroupBy []string

	// When comparing tags, use this selection to figure out if a rule applies to them
	Seletion query.Select

	// Limit Key is an optional param that can be used to dictate what value of the tags to use as a limiter
	LimitKey *LimitKey
}

type LimitKey struct {
	Default uint64

	// optional value to check against for incoming requests
	Name *string
}
