package rules

import (
	v1limitermodels "github.com/DanLavine/willow/internal/limiter/v1_limiter_models"
	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"

	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=rulefakes/rule_mock.go -package=rulefakes github.com/DanLavine/willow/internal/limiter/rules Rule
type Rule interface {
	// retrieve the Name of the rules
	Name() string
	// retrieve the original group by keys
	GetGroupByKeys() []string

	// Update a particualr rule's default limit values
	Update(logger *zap.Logger, update *v1limiter.RuleUpdateRquest)

	// query overrides for a particular rule
	QueryOverrides(logger *zap.Logger, query *v1common.AssociatedQuery) (v1limiter.Overrides, *errors.ServerError)

	// set an override for a particualr group of tags
	SetOverride(logger *zap.Logger, override *v1limiter.Override) *errors.ServerError

	// Delete an override for a particualr group of tags
	DeleteOverride(logger *zap.Logger, overrideName string) *errors.ServerError

	// Operation that is calledfor cascading deletes because the rule is being deleted
	CascadeDeletion(logger *zap.Logger) *errors.ServerError

	// Find the limits for a particualr group of tags including any overrides
	FindLimits(logger *zap.Logger, keyValues datatypes.KeyValues) (v1limitermodels.Limits, *errors.ServerError)

	// Get a rule response for Read operations
	Get(includeOverrides *v1limiter.RuleQuery) *v1limiter.Rule
}
