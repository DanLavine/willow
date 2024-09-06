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
	CreateOverride(ctx context.Context, ruleID string, override *v1limiter.Override) (string, *errors.ServerError)

	// Update an override by name
	UpdateOverride(ctx context.Context, ruleID string, overrideName string, override *v1limiter.OverrideProperties) *errors.ServerError

	// read operations
	GetOverride(ctx context.Context, ruleID string, overrideName string) (*v1limiter.Override, *errors.ServerError)
	QueryOverrides(ctx context.Context, ruleID string, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Overrides, *errors.ServerError)
	MatchOverrides(ctx context.Context, ruleID string, match *querymatchaction.MatchActionQuery) (v1limiter.Overrides, *errors.ServerError)

	// delete operations
	DestroyOverride(ctx context.Context, ruleID string, overrideName string) *errors.ServerError
	DestroyOverrides(ctx context.Context, ruleID string) *errors.ServerError

	// loogical operations
	FindOverrideLimits(ctx context.Context, ruleID string, keyValue datatypes.KeyValues) (v1limiter.Overrides, *errors.ServerError)
}
