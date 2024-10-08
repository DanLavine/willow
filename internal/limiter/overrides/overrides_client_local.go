package overrides

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	btreeonetomany "github.com/DanLavine/willow/internal/datastructures/btree_one_to_many"
	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/middleware"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

func errorMissingOverrideName(name string) *errors.ServerError {
	return &errors.ServerError{Message: fmt.Sprintf("Override '%s' not found", name), StatusCode: http.StatusNotFound}
}

type overridesClientLocal struct {
	// where overrides are saved
	overrides btreeonetomany.BTreeOneToMany

	// constructor for the types of overrides
	constructor OverrideConstructor
}

func NewDefaultOverridesClientLocal(constructor OverrideConstructor) *overridesClientLocal {
	return &overridesClientLocal{
		overrides:   btreeonetomany.NewThreadSafe(),
		constructor: constructor,
	}
}

func NewOverridesClientLocal(tree btreeonetomany.BTreeOneToMany, constructor OverrideConstructor) *overridesClientLocal {
	return &overridesClientLocal{
		overrides:   tree,
		constructor: constructor,
	}
}

func (ocl *overridesClientLocal) CreateOverride(ctx context.Context, ruleID string, override *v1limiter.Override) (string, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "CreateOverride")

	onCreate := func() any {
		return ocl.constructor.New(override.Spec.Properties)
	}

	var alreadyExistsErr *errors.ServerError
	onFind := func(oneToManyItem btreeonetomany.OneToManyItem) {
		alreadyExistsErr = &errors.ServerError{Message: "override KeyValues already exists", StatusCode: http.StatusConflict}

	}

	id, err := ocl.overrides.CreateOrFind(ruleID, override.Spec.DBDefinition.GroupByKeyValues, onCreate, onFind)
	if err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		case btreeonetomany.ErrorManyIDDestroying:
			// override is currently being destroyed
			logger.Warn("override is being destroyed")
			return "", &errors.ServerError{Message: "override is being destroy", StatusCode: http.StatusConflict}
		default:
			logger.Error("Unexpected error creating the override", zap.Error(err))
			return "", errors.InternalServerError
		}
	}

	return id, alreadyExistsErr
}

func (ocl *overridesClientLocal) GetOverride(ctx context.Context, ruleID string, overrideName string) (*v1limiter.Override, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "GetOverride")

	var limiterOverride *v1limiter.Override
	overrideErr := errorMissingOverrideName(overrideName)

	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)

		limiterOverride = &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: item.ManyKeyValues(),
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf(override.Limit()),
				},
			},
			State: &v1limiter.OverrideState{
				ID:       item.ManyID(),
				Deleting: false,
			},
		}

		overrideErr = nil
		return false
	}

	if err := ocl.overrides.QueryAction(ruleID, queryassociatedaction.StringToAssociatedActionQuery(overrideName), onIterate); err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		default:
			logger.Error("Unexpected error creating the override", zap.Error(err))
			return nil, errors.InternalServerError
		}
	}

	return limiterOverride, overrideErr
}

// Query the overrides
func (ocl *overridesClientLocal) QueryOverrides(ctx context.Context, ruleID string, query *queryassociatedaction.AssociatedActionQuery) (v1limiter.Overrides, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "QueryOverrides")
	overrides := v1limiter.Overrides{}

	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)

		overrides = append(overrides, &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: item.ManyKeyValues(),
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf(override.Limit()),
				},
			},
			State: &v1limiter.OverrideState{
				ID:       item.ManyID(),
				Deleting: false,
			},
		})

		return true
	}

	if err := ocl.overrides.QueryAction(ruleID, query, onIterate); err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		default:
			logger.Error("Unexpected error matching all overrides", zap.Error(err))
			return nil, errors.InternalServerError
		}
	}

	return overrides, nil
}

// Query the overrides
func (ocl *overridesClientLocal) MatchOverrides(ctx context.Context, ruleID string, match *querymatchaction.MatchActionQuery) (v1limiter.Overrides, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "MatchOverrides")
	overrides := v1limiter.Overrides{}

	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)

		overrides = append(overrides, &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: item.ManyKeyValues(),
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf(override.Limit()),
				},
			},
			State: &v1limiter.OverrideState{
				ID:       item.ManyID(),
				Deleting: false,
			},
		})

		return true
	}

	if err := ocl.overrides.MatchAction(ruleID, match, onIterate); err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		default:
			logger.Error("Unexpected error matching all overrides", zap.Error(err))
			return nil, errors.InternalServerError
		}
	}

	return overrides, nil
}

func (ocl *overridesClientLocal) UpdateOverride(ctx context.Context, ruleID string, overrideName string, overrideUpdate *v1limiter.OverrideProperties) *errors.ServerError {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "UpdateOverride")
	overrideErr := errorMissingOverrideName(overrideName)

	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		override.Update(overrideUpdate)

		overrideErr = nil
		return false
	}

	if err := ocl.overrides.QueryAction(ruleID, queryassociatedaction.StringToAssociatedActionQuery(overrideName), onIterate); err != nil {
		switch err {
		//case ErrorOneIDDestroying.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		default:
			logger.Error("Unexpected error updating the override", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return overrideErr
}

func (ocl *overridesClientLocal) DestroyOverride(ctx context.Context, ruleID string, overrideName string) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "DestroyOverride")

	var deleteErr *errors.ServerError
	canDelete := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		deleteErr = override.Delete()

		return deleteErr == nil
	}

	if err := ocl.overrides.DestroyOneOfManyByID(ruleID, overrideName, canDelete); err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		case btreeonetomany.ErrorManyIDDestroying:
		// This case is fine since it is currently being destroyed. Just return a 204, the other request
		// in progress for this request will return the issue on a filure
		default:
			logger.Error("Unexpected error deleting the override", zap.Error(err))
			return errors.InternalServerError
		}
	}

	if deleteErr != nil {
		logger.Error("failed to delete the override for the rule", zap.Error(deleteErr))
		return errors.InternalServerError
	}

	return nil
}

func (ocl *overridesClientLocal) DestroyOverrides(ctx context.Context, ruleID string) *errors.ServerError {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "DEstroyOVerrides")

	var deleteErr *errors.ServerError
	canDelete := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		deleteErr = override.Delete()

		return deleteErr == nil
	}

	if err := ocl.overrides.DestroyOne(ruleID, canDelete); err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		default:
			logger.Error("Unexpected error destroying all the override", zap.Error(err))
			return errors.InternalServerError
		}
	}

	if deleteErr != nil {
		logger.Error("failed to delete the overrides for the rule", zap.Error(deleteErr))
		return errors.InternalServerError
	}

	return nil
}

func (ocl *overridesClientLocal) FindOverrideLimits(ctx context.Context, ruleID string, keyValues datatypes.KeyValues) (v1limiter.Overrides, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "FindOverrideLimits")
	overrides := v1limiter.Overrides{}

	// stop if a limit is at 0
	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		limit := override.Limit()

		overrides = append(overrides, &v1limiter.Override{
			Spec: &v1limiter.OverrideSpec{
				DBDefinition: &v1limiter.OverrideDBDefinition{
					GroupByKeyValues: item.ManyKeyValues(),
				},
				Properties: &v1limiter.OverrideProperties{
					Limit: helpers.PointerOf(override.Limit()),
				},
			},
			State: &v1limiter.OverrideState{
				ID:       item.ManyID(),
				Deleting: false,
			},
		})

		return limit != 0
	}

	if err := ocl.overrides.MatchAction(ruleID, querymatchaction.KeyValuesToAnyMatchActionQuery(keyValues), onIterate); err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		default:
			logger.Error("Unexpected error matching override selection", zap.Error(err))
			return nil, errors.InternalServerError
		}
	}

	return overrides, nil
}
