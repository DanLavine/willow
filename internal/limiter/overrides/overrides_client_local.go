package overrides

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	btreeonetomany "github.com/DanLavine/willow/internal/datastructures/btree_one_to_many"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
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

func (ocl *overridesClientLocal) CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *errors.ServerError {
	logger = logger.Named("CreateOverride").With(zap.String("override_name", override.Name))

	onCreate := func() any {
		return ocl.constructor.New(override)
	}

	if err := ocl.overrides.CreateWithID(ruleName, override.Name, override.KeyValues, onCreate); err != nil {
		switch err {
		//case datastructures.ErrorOneIDDestroying:
		// This shouldn't happen as the deletion of the `Rule` should block all these request`
		case btreeonetomany.ErrorManyIDDestroying:
			// override is currently being destroyed
			logger.Warn("override is being destroyed")
			return &errors.ServerError{Message: "override is being destroy", StatusCode: http.StatusConflict}
		case btreeonetomany.ErrorManyIDAlreadyExists:
			// override name already exists
			logger.Warn("override name is already taken")
			return &errors.ServerError{Message: "override Name alreayd exists", StatusCode: http.StatusConflict}
		case btreeonetomany.ErrorManyKeyValuesAlreadyExist:
			// key values for the override already exist
			logger.Warn("override key values are already taken", zap.Any("key_values", override.KeyValues))
			return &errors.ServerError{Message: "override KeyValues alreayd exists", StatusCode: http.StatusConflict}
		default:
			logger.Error("Unexpected error creating the override", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return nil
}

func (ocl *overridesClientLocal) GetOverride(logger *zap.Logger, ruleName string, overrideName string) (*v1limiter.Override, *errors.ServerError) {
	logger = logger.Named("GetOverride").With(zap.String("override_name", overrideName))

	var limiterOverride *v1limiter.Override
	overrideErr := errorMissingOverrideName(overrideName)

	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)

		limiterOverride = &v1limiter.Override{
			Name:      item.ManyID(),
			KeyValues: item.ManyKeyValues(),
			Limit:     override.Limit(),
		}

		overrideErr = nil
		return false
	}

	overrideNameValue := datatypes.String(overrideName)
	query := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{
				"_associated_id": datatypes.Value{Value: &overrideNameValue, ValueComparison: datatypes.EqualsPtr()},
			},
		},
	}

	if err := ocl.overrides.Query(ruleName, query, onIterate); err != nil {
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

func (ocl *overridesClientLocal) MatchOverrides(logger *zap.Logger, ruleName string, query *v1common.MatchQuery) (v1limiter.Overrides, *errors.ServerError) {
	logger = logger.Named("MatchOverrides")
	overrides := v1limiter.Overrides{}

	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)

		overrides = append(overrides, &v1limiter.Override{
			Name:      item.ManyID(),
			KeyValues: item.ManyKeyValues(),
			Limit:     override.Limit(),
		})

		return true
	}

	if query.KeyValues == nil {
		// match all
		if err := ocl.overrides.Query(ruleName, datatypes.AssociatedKeyValuesQuery{}, onIterate); err != nil {
			switch err {
			//case datastructures.ErrorOneIDDestroying:
			// This shouldn't happen as the deletion of the `Rule` should block all these request`
			default:
				logger.Error("Unexpected error matching all overrides", zap.Error(err))
				return nil, errors.InternalServerError
			}
		}
	} else {
		// match against the key values
		if err := ocl.overrides.MatchPermutations(ruleName, *query.KeyValues, onIterate); err != nil {
			switch err {
			//case datastructures.ErrorOneIDDestroying:
			// This shouldn't happen as the deletion of the `Rule` should block all these request`
			default:
				logger.Error("Unexpected error matching override selection", zap.Error(err))
				return nil, errors.InternalServerError
			}
		}
	}

	return overrides, nil
}

func (ocl *overridesClientLocal) UpdateOverride(logger *zap.Logger, ruleName string, overrideName string, overrideUpdate *v1limiter.OverrideUpdate) *errors.ServerError {
	logger = logger.Named("CreateOverride").With(zap.String("override_name", overrideName))
	overrideErr := errorMissingOverrideName(overrideName)

	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		override.Update(overrideUpdate)

		overrideErr = nil
		return false
	}

	overrideNameValue := datatypes.String(overrideName)
	query := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{
				"_associated_id": datatypes.Value{Value: &overrideNameValue, ValueComparison: datatypes.EqualsPtr()},
			},
		},
	}

	if err := ocl.overrides.Query(ruleName, query, onIterate); err != nil {
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

func (ocl *overridesClientLocal) DestroyOverride(logger *zap.Logger, ruleName string, overrideName string) *errors.ServerError {
	logger = logger.Named("DeleteOverride")

	var deleteErr *errors.ServerError
	canDelete := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		deleteErr = override.Delete()

		return deleteErr == nil
	}

	if err := ocl.overrides.DestroyOneOfManyByID(ruleName, overrideName, canDelete); err != nil {
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
		logger.Error("failed to delete the override for the rule", zap.String("override_name", overrideName), zap.Error(deleteErr))
		return errors.InternalServerError
	}

	return nil
}

func (ocl *overridesClientLocal) DestroyOverrides(logger *zap.Logger, ruleName string) *errors.ServerError {
	logger = logger.Named("DeleteOverride")

	var deleteErr *errors.ServerError
	canDelete := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		deleteErr = override.Delete()

		return deleteErr == nil
	}

	if err := ocl.overrides.DestroyOne(ruleName, canDelete); err != nil {
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

func (ocl *overridesClientLocal) FindOverrideLimits(logger *zap.Logger, ruleName string, keyValues datatypes.KeyValues) (v1limiter.Overrides, *errors.ServerError) {
	logger = logger.Named("FindOverrideLimits")
	overrides := v1limiter.Overrides{}

	// stop if a limit is at 0
	onIterate := func(item btreeonetomany.OneToManyItem) bool {
		override := item.Value().(Override)
		limit := override.Limit()

		overrides = append(overrides, &v1limiter.Override{
			Name:      item.ManyID(),
			KeyValues: item.ManyKeyValues(),
			Limit:     limit,
		})

		return limit != 0
	}

	if err := ocl.overrides.MatchPermutations(ruleName, keyValues, onIterate); err != nil {
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
