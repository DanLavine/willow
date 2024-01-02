package overrides

import (
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// override operations
type OverrideClient interface {
	// Create a new Override
	CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *errors.ServerError

	// Update an override by name
	UpdateOverride(logger *zap.Logger, ruleName string, overrideName string, override *v1limiter.OverrideUpdate) *errors.ServerError

	// read operations
	GetOverride(logger *zap.Logger, ruleName string, overrideName string) (*v1limiter.Override, *errors.ServerError)
	MatchOverrides(logger *zap.Logger, ruleName string, query *v1common.MatchQuery) (v1limiter.Overrides, *errors.ServerError)

	// delete operations
	DestroyOverride(logger *zap.Logger, ruleName string, overrideName string) *errors.ServerError
	DestroyOverrides(logger *zap.Logger, ruleName string) *errors.ServerError

	// loogical operations
	FindOverrideLimits(logger *zap.Logger, ruleName string, keyValue datatypes.KeyValues) (v1limiter.Overrides, *errors.ServerError)
}
