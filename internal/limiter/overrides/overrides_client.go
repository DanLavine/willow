package overrides

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// override operations
type OverrideClient interface {
	// Create a new Override
	CreateOverride(ctx context.Context, ruleName string, override *v1limiter.Override) *errors.ServerError

	// Update an override by name
	UpdateOverride(ctx context.Context, ruleName string, overrideName string, override *v1limiter.OverrideUpdate) *errors.ServerError

	// read operations
	GetOverride(ctx context.Context, ruleName string, overrideName string) (*v1limiter.Override, *errors.ServerError)
	MatchOverrides(ctx context.Context, ruleName string, query *v1common.MatchQuery) (v1limiter.Overrides, *errors.ServerError)

	// delete operations
	DestroyOverride(ctx context.Context, ruleName string, overrideName string) *errors.ServerError
	DestroyOverrides(ctx context.Context, ruleName string) *errors.ServerError

	// loogical operations
	FindOverrideLimits(ctx context.Context, ruleName string, keyValue datatypes.KeyValues) (v1limiter.Overrides, *errors.ServerError)
}
