package rules

import (
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
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
	CreateRule(logger *zap.Logger, rule *v1limiter.RuleCreateRequest) *errors.ServerError
	//// update
	UpdateRule(logger *zap.Logger, ruleName string, update *v1limiter.RuleUpdateRquest) *errors.ServerError
	//// read
	GetRule(logger *zap.Logger, ruleName string, query *v1limiter.RuleGet) (*v1limiter.Rule, *errors.ServerError)
	MatchRules(logger *zap.Logger, matchQuery *v1limiter.RuleMatch) (v1limiter.Rules, *errors.ServerError)
	//// delete operations
	DeleteRule(logger *zap.Logger, ruleName string) *errors.ServerError

	// override operations
	//// create
	CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *errors.ServerError
	//// update
	UpdateOverride(logger *zap.Logger, ruleName string, overrideName string, override *v1limiter.OverrideUpdate) *errors.ServerError
	//// read
	GetOverride(logger *zap.Logger, ruleName string, overrideName string) (*v1limiter.Override, *errors.ServerError)
	MatchOverrides(logger *zap.Logger, ruleName string, query *v1common.MatchQuery) (v1limiter.Overrides, *errors.ServerError)
	//// delete
	DeleteOverride(logger *zap.Logger, ruleName string, overrideName string) *errors.ServerError

	// Logical operations
	//// FindLmits is similar to MatchRules, but has special logic to bail early if it notices a Rule or Override have a limit of 0
	FindLimits(logger *zap.Logger, keyValues datatypes.KeyValues) (v1limiter.Rules, *errors.ServerError)
}
