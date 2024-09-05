package rules

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// "Clients" should always take in api objects and return api objects

// Rule client is used to interact with any rules. On the "Rules" service it will be the local
// client. But on the "Counters" service it can eventual be an http client. However, since everything
// is all in 1 for now, it is also a local client
//
//go:generate mockgen -destination=rulesfakes/rule_client_mock.go -package=rulesfakes github.com/DanLavine/willow/internal/limiter/rules RuleClient
type RuleClient interface {
	// rule operations
	//// create
	CreateRule(ctx context.Context, rule *v1limiter.Rule) (string, *errors.ServerError)
	//// update
	UpdateRule(ctx context.Context, ruleID string, update *v1limiter.RuleProperties) *errors.ServerError
	//// read
	QueryRules(ctx context.Context, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Rules, *errors.ServerError)
	MatchRules(ctx context.Context, match *querymatchaction.MatchActionQuery) (v1limiter.Rules, *errors.ServerError)
	GetRule(ctx context.Context, ruleID string) (*v1limiter.Rule, *errors.ServerError)
	//// delete operations
	DeleteRule(ctx context.Context, ruleID string) *errors.ServerError

	// override operations
	//// create
	CreateOverride(ctx context.Context, ruleID string, override *v1limiter.Override) (string, *errors.ServerError)
	//// update
	UpdateOverride(ctx context.Context, ruleID string, overrideName string, override *v1limiter.OverrideProperties) *errors.ServerError
	//// read
	QueryOverrides(ctx context.Context, ruleID string, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Overrides, *errors.ServerError)
	MatchOverrides(ctx context.Context, ruleID string, match *querymatchaction.MatchActionQuery) (v1limiter.Overrides, *errors.ServerError)
	GetOverride(ctx context.Context, ruleID string, overrideName string) (*v1limiter.Override, *errors.ServerError)
	//// delete
	DeleteOverride(ctx context.Context, ruleID string, overrideName string) *errors.ServerError

	// Logical operations
	//// FindLmits is similar to MatchRules, but has special logic to bail early if it notices a Rule or Override have a limit of 0
	FindLimits(ctx context.Context, keyValues datatypes.KeyValues) (v1limiter.Rules, *errors.ServerError)
}
