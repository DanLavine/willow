package overrides

import (
	"context"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
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
	QueryOverrides(ctx context.Context, ruleName string, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Overrides, *errors.ServerError)
	MatchOverrides(ctx context.Context, ruleName string, match *querymatchaction.MatchActionQuery) (v1limiter.Overrides, *errors.ServerError)

	// delete operations
	DestroyOverride(ctx context.Context, ruleName string, overrideName string) *errors.ServerError
	DestroyOverrides(ctx context.Context, ruleName string) *errors.ServerError

	// loogical operations
	FindOverrideLimits(ctx context.Context, ruleName string, keyValue datatypes.KeyValues) (v1limiter.Overrides, *errors.ServerError)
}
