package memory

import (
	"net/http"
	"sync"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/pkg/models/api"
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

		addOverride := true
		switch includeOverrides.OverrideQuery {
		case v1limiter.Match:
			// match the rule's limit for the kay values with the provided query key values
			if ruleOverride.limit < uint64(len(*includeOverrides.KeyValues)) {
				addOverride = false
			}
		default: // v1limiter.All:
			// fall through
		}

		if addOverride {
			overrides = append(overrides, v1limiter.Override{
				Name:      associatedKeyValues.AssociatedID(),
				KeyValues: associatedKeyValues.KeyValues().StripAssociatedID().RetrieveStringDataType(),
				Limit:     ruleOverride.limit,
			})
		}

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

func (r *rule) FindLimit(logger *zap.Logger, keyValues datatypes.KeyValues) (uint64, *api.Error) {
	r.ruleModelLock.RLock()
	defer r.ruleModelLock.RUnlock()

	limit := r.limit

	// TODO: fix me
	//onFind := func(item any) bool {
	//	ruleOverride := item.(*btreeassociated.AssociatedKeyValues).Value().(*ruleOverride)
	//	ruleOverride.lock.RLock()
	//	defer ruleOverride.lock.RUnlock()
	//
	//	limit = ruleOverride.limit
	//	return false
	//}
	//
	//// setup the query for the override we are looking for
	//lenKeyValues := len(keyValues)
	//
	//// This query also isn't correct. It should find all values. that match. Like in the manager to find the rules,
	//// I need that special "match" query.
	//query := datatypes.AssociatedKeyValuesQuery{
	//	KeyValueSelection: &datatypes.KeyValueSelection{
	//		KeyValues: map[string]datatypes.Value{},
	//		Limits: &datatypes.KeyLimits{
	//			NumberOfKeys: &lenKeyValues,
	//		},
	//	},
	//}
	//
	//for key, value := range keyValues {
	//	query.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &value, ValueComparison: datatypes.EqualsPtr()}
	//}
	//
	//if err := r.overrides.Query(query, onFind); err != nil {
	//	return 0, errors.InternalServerError
	//}

	return limit, nil
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
func (r *rule) SetOverride(logger *zap.Logger, override *v1limiter.Override) *api.Error {
	logger = logger.Named("SetOverride")

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
		// need to lock just to ensure an update isn't taking place at the same time
		override := item.(*btreeassociated.AssociatedKeyValues).Value().(*ruleOverride)
		override.lock.Lock()
		defer override.lock.Unlock()

		logger.Debug("Successfully deleted override")
		deleted = true
		return true
	}

	if err := r.overrides.DeleteByAssociatedID(overrideName, canDelete); err != nil {
		logger.Error("Failed to delete override. Unexpected error from BtreeAssociated", zap.Error(err))
		return (&api.Error{Message: "failed to delete override. Internal server error", StatusCode: http.StatusInternalServerError}).With("", err.Error())
	}

	if !deleted {
		logger.Debug("Failed to delete the override. Could not find by the requested name")
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

// DSL TODO: These are unused right now

func (r *rule) TagsMatch(logger *zap.Logger, keyValues datatypes.KeyValues) bool {
	// ensure that all the "group by" keys exists
	for _, key := range r.groupBy {
		if _, ok := keyValues[key]; !ok {
			return false
		}
	}

	return true
}

func (r *rule) GenerateQuery(keyValues datatypes.KeyValues) datatypes.AssociatedKeyValuesQuery {
	selectQuery := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{},
		},
	}

	for _, key := range r.groupBy {
		value := keyValues[key]
		selectQuery.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &value, ValueComparison: datatypes.EqualsPtr()}
	}

	return selectQuery
}
