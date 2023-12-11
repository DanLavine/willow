package memory

import (
	"fmt"
	"net/http"
	"sync"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/errors"
	v1limitermodels "github.com/DanLavine/willow/internal/limiter/v1_limiter_models"
	"github.com/DanLavine/willow/pkg/models/api"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

type rule struct {
	ruleModelLock *sync.RWMutex
	name          string
	groupBy       []string
	limit         uint64

	// all values in the overrides are of type 'ruleOverride'
	overrides btreeassociated.BTreeAssociated
}

type ruleOverride struct {
	lock *sync.RWMutex

	limit uint64
}

func NewRule(ruleModel *v1limiter.RuleRequest) *rule {
	return &rule{
		ruleModelLock: new(sync.RWMutex),
		name:          ruleModel.Name,
		groupBy:       ruleModel.GroupBy,
		limit:         ruleModel.Limit,
		overrides:     btreeassociated.NewThreadSafe(),
	}
}

// Get converts a rule to an API response.
//
//	PARAMS:
//	- includeOverrodes - iff true, will also include any rule overrides. This can be a SLOW operation.
func (r *rule) Get(includeOverrides *v1limiter.RuleQuery) *v1limiter.RuleResponse {
	r.ruleModelLock.RLock()
	defer r.ruleModelLock.RUnlock()

	var overrides []v1limiter.Override
	onPagination := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
		ruleOverride := associatedKeyValues.Value().(*ruleOverride)
		ruleOverride.lock.RLock()
		defer ruleOverride.lock.RUnlock()

		overrides = append(overrides, v1limiter.Override{
			Name:      associatedKeyValues.AssociatedID(),
			KeyValues: associatedKeyValues.KeyValues().StripAssociatedID().RetrieveStringDataType(),
			Limit:     ruleOverride.limit,
		})

		return true
	}

	switch includeOverrides.OverrideQuery {
	case v1limiter.None:
		// nothing to do here
	case v1limiter.Match:
		// match all the key values
		if err := r.overrides.MatchPermutations(btreeassociated.ConverDatatypesKeyValues(*includeOverrides.KeyValues), onPagination); err != nil {
			panic(err)
		}
	case v1limiter.All:
		// should not error. That only happens on param validation
		if err := r.overrides.Query(datatypes.AssociatedKeyValuesQuery{}, onPagination); err != nil {
			panic(err)
		}
	}

	ruleResponse := &v1limiter.RuleResponse{
		Name:      r.name,
		GroupBy:   r.groupBy,
		Limit:     r.limit,
		Overrides: overrides,
	}

	return ruleResponse
}

// DSL TODO: There is an optimization here, where I can find a "subset" of all the key values if they have a lower
// value and just use that. I believe that holds true.
//
//	I.E
//	1. {"key1":"1", "key2":"2"}, Limit 5
//	2. {"key1":"1"}, Limit 2 <- this is always more restrictive and we don't care about the 1st rule anymore
func (r *rule) FindLimits(logger *zap.Logger, keyValues datatypes.KeyValues) (v1limitermodels.Limits, *api.Error) {
	r.ruleModelLock.RLock()
	defer r.ruleModelLock.RUnlock()

	// setup initial limits for the rule
	limitKeyValues := datatypes.KeyValues{}
	for _, key := range r.groupBy {
		limitKeyValues[key] = keyValues[key]
	}
	limits := v1limitermodels.Limits{v1limitermodels.Limit{Name: r.name, KeyValues: limitKeyValues, Limit: r.limit}}

	// account for any overrides
	counter := 0
	onPagination := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
		ruleOverride := associatedKeyValues.Value().(*ruleOverride)
		ruleOverride.lock.RLock()
		defer ruleOverride.lock.RUnlock()

		newLimitKeyValues := datatypes.KeyValues{}
		for key, value := range associatedKeyValues.KeyValues().StripAssociatedID().RetrieveStringDataType() {
			newLimitKeyValues[key] = value
		}

		if counter == 0 {
			// take the 1st override
			limits = v1limitermodels.Limits{v1limitermodels.Limit{Name: r.name, KeyValues: newLimitKeyValues, Limit: ruleOverride.limit}}
			counter++
		} else {
			// append additional limits
			limits = append(limits, v1limitermodels.Limit{Name: r.name, KeyValues: newLimitKeyValues, Limit: ruleOverride.limit})
		}

		// can exit early since we have an override with 0. The request is 100% rejected
		return ruleOverride.limit != 0
	}

	// match all the override KeyValue permutations
	if err := r.overrides.MatchPermutations(btreeassociated.ConverDatatypesKeyValues(keyValues), onPagination); err != nil {
		logger.Error("error finding limits", zap.Error(err))
		return limits, errors.InternalServerError
	}

	return limits, nil
}

func (r *rule) Update(logger *zap.Logger, update *v1limiter.RuleUpdate) {
	r.ruleModelLock.Lock()
	defer r.ruleModelLock.Unlock()

	r.limit = uint64(update.Limit)
	logger.Debug("updated rule")
}

// Create an override for a specific rule.
//
// NOTE: we don't need to ensure on the Override's KeyValues that they have all the Rule's GroupBy tags. Thise are already
// finding the inital rule to lookup
func (r *rule) QueryOverrides(logger *zap.Logger, query *v1limiter.Query) (v1limiter.Overrides, *api.Error) {
	logger = logger.Named("QueryOverrides")

	var overrides v1.Overrides
	var overrideErr *api.Error

	onfindPagination := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
		ruleOverride := associatedKeyValues.Value().(*ruleOverride)
		ruleOverride.lock.RLock()
		defer ruleOverride.lock.RUnlock()

		overrides = append(overrides, v1limiter.Override{
			Name:      associatedKeyValues.AssociatedID(),
			KeyValues: associatedKeyValues.KeyValues().StripAssociatedID().RetrieveStringDataType(),
			Limit:     ruleOverride.limit,
		})
		return true
	}

	if err := r.overrides.Query(query.AssociatedKeyValues, onfindPagination); err != nil {
		logger.Error("Failed to query overrides", zap.Error(err))
		return overrides, &api.Error{Message: "Failed to query overrides", StatusCode: http.StatusInternalServerError}
	}

	fmt.Println("overrides:", overrides)
	return overrides, overrideErr
}

// Create an override for a specific rule.
//
// NOTE: we don't need to ensure on the Override's KeyValues that they have all the Rule's GroupBy tags. Thise are already
// finding the inital rule to lookup
func (r *rule) SetOverride(logger *zap.Logger, override *v1limiter.Override) *api.Error {
	logger = logger.Named("SetOverride")

	// ensure that the override has all the group by keys
	for _, key := range r.groupBy {
		if _, ok := override.KeyValues[key]; !ok {
			return &api.Error{Message: fmt.Sprintf("Missing Rule's GroubBy keys in the override: %s", key), StatusCode: http.StatusBadRequest}
		}
	}

	// create custom override rule paramters
	keyValues := btreeassociated.ConverDatatypesKeyValues(override.KeyValues)

	onCreate := func() any {
		return &ruleOverride{
			lock:  &sync.RWMutex{},
			limit: override.Limit,
		}
	}

	if err := r.overrides.CreateWithID(override.Name, keyValues, onCreate); err != nil {
		logger.Error("failed to Create rule override", zap.Error(err), zap.String("name", override.Name))

		switch err {
		case btreeassociated.ErrorCreateFailureKeyValuesExist:
			return (&api.Error{Message: "failed to create rule override", StatusCode: http.StatusUnprocessableEntity}).With("key values to not have an override already", "key values aready in use")
		case btreeassociated.ErrorAssociatedIDAlreadyExists:
			return (&api.Error{Message: "failed to create rule override", StatusCode: http.StatusUnprocessableEntity}).With("name to not be in use", "name is already in use by another override")
		default:
			return (&api.Error{Message: "failed to create, unexpected error", StatusCode: http.StatusInternalServerError}).With("", err.Error())
		}
	}

	return nil
}

func (r *rule) DeleteOverride(logger *zap.Logger, overrideName string) *api.Error {
	logger = logger.Named("DeleteOverride").With(zap.String("override name", overrideName))

	deleted := false
	canDelete := func(item any) bool {
		deleted = true

		// need to lock just to ensure an update isn't taking place at the same time
		override := item.(*btreeassociated.AssociatedKeyValues).Value().(*ruleOverride)
		override.lock.Lock()
		defer override.lock.Unlock()

		logger.Debug("Successfully deleted override")
		return true
	}

	if err := r.overrides.DeleteByAssociatedID(overrideName, canDelete); err != nil {
		logger.Error("Failed to delete override. Unexpected error from BtreeAssociated", zap.Error(err))
		return (&api.Error{Message: "failed to delete override. Internal server error", StatusCode: http.StatusInternalServerError}).With("", err.Error())
	}

	if !deleted {
		logger.Debug("Failed to delete the override. Could not find by the requested name")
		return &api.Error{Message: fmt.Sprintf("Override %s not found", overrideName), StatusCode: http.StatusNotFound}
	}

	return nil
}

// General helpers for the rule

// CascadeDeletion is called when the Rule itself is being deleted. On the memory implementation
// we don't need to do anything as the object will be garbage collected
func (r *rule) CascadeDeletion(logger *zap.Logger) *api.Error {
	return nil
}

func (r *rule) Name() string {
	return r.name
}

func (r *rule) GetGroupByKeys() []string {
	return r.groupBy
}
