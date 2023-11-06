package memory

import (
	"net/http"
	"sync"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

type rule struct {
	ruleModelLock *sync.RWMutex
	ruleModel     *v1limiter.Rule

	// all values in the overrides are of type 'ruleOverride'
	overrides btreeassociated.BTreeAssociated
}

type ruleOverride struct {
	lock *sync.RWMutex

	keyValues datatypes.KeyValues
	limit     uint64
}

func NewRule(ruleModel *v1limiter.Rule) *rule {
	return &rule{
		ruleModelLock: new(sync.RWMutex),
		ruleModel:     ruleModel,
		overrides:     btreeassociated.NewThreadSafe(),
	}
}

func (r *rule) Update(logger *zap.Logger, newLimit uint64) {
	r.ruleModelLock.Lock()
	defer r.ruleModelLock.Unlock()

	r.ruleModel.Limit = uint64(newLimit)
}

func (r *rule) SetOverride(logger *zap.Logger, override *v1limiter.Override) *api.Error {
	logger = logger.Named("SetOverride")

	// set an override iff the tags aare valid
	if !r.ruleModel.QueryFilter.MatchTags(override.KeyValues) {
		return api.NotAcceptable.With("the provided keys values to match the rule query", "provided will never be found by the rule query")
	}

	onCreate := func() any {
		return &ruleOverride{
			lock:      &sync.RWMutex{},
			keyValues: override.KeyValues,
			limit:     override.Limit,
		}
	}

	onFind := func(item any) {
		ruleOverride := item.(*btreeassociated.AssociatedKeyValues).Value().(*ruleOverride)
		ruleOverride.lock.Lock()
		defer ruleOverride.lock.Unlock()

		ruleOverride.limit = override.Limit
	}

	if _, err := r.overrides.CreateOrFind(btreeassociated.ConverDatatypesKeyValues(override.KeyValues), onCreate, onFind); err != nil {
		logger.Error("failed to CreateOrFind a rule override", zap.Error(err))
		return errors.InternalServerError.With("", err.Error())
	}

	return nil
}

func (r *rule) DeleteOverride(logger *zap.Logger, query datatypes.AssociatedKeyValuesQuery) *api.Error {
	logger = logger.Named("DeleteOverride")

	// if err := r.overrides.Delete(override.KeyValues, func(_ any) bool { return true }); err != nil {
	// 	logger.Error("failed to delete a rule override", zap.Error(err))
	// 	return errors.InternalServerError.With("", err.Error())
	// }

	// return nil

	return &api.Error{Message: "not implemented delete override", StatusCode: http.StatusNotImplemented}
}

func (r *rule) FindLimit(logger *zap.Logger, keyValues datatypes.KeyValues) uint64 {
	r.ruleModelLock.RLock()
	limit := r.ruleModel.Limit
	r.ruleModelLock.RUnlock()

	onFind := func(item any) {
		ruleOverride := item.(*btreeassociated.AssociatedKeyValues).Value().(*ruleOverride)
		ruleOverride.lock.RLock()
		defer ruleOverride.lock.RUnlock()

		limit = ruleOverride.limit
	}

	// ignore these errors... should make it so it just panics
	_, _ = r.overrides.Find(btreeassociated.ConverDatatypesKeyValues(keyValues), onFind)

	return limit
}

func (r *rule) TagsMatch(logger *zap.Logger, keyValues datatypes.KeyValues) bool {
	// ensure that all the "group by" keys exists
	for _, key := range r.ruleModel.GroupBy {
		if _, ok := keyValues[key]; !ok {
			return false
		}
	}

	// ensure that the selection doesn't filter out the request
	return r.ruleModel.QueryFilter.MatchTags(keyValues)
}

func (r *rule) Lock() {
	r.ruleModelLock.RLock()
}

func (r *rule) Unlock() {
	r.ruleModelLock.RLock()
}

func (r *rule) GenerateQuery(keyValues datatypes.KeyValues) datatypes.AssociatedKeyValuesQuery {
	selectQuery := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{},
		},
	}

	for _, key := range r.ruleModel.GroupBy {
		value := keyValues[key]
		selectQuery.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &value, ValueComparison: datatypes.EqualsPtr()}
	}

	return selectQuery
}

func (r *rule) GetRuleResponse(includeOverrides bool) *v1limiter.Rule {
	r.ruleModelLock.RLock()
	defer r.ruleModelLock.RUnlock()

	var overrides []v1limiter.Override

	if includeOverrides {
		onPagiination := func(item any) bool {
			ruleOverride := item.(*btreeassociated.AssociatedKeyValues).Value().(*ruleOverride)
			ruleOverride.lock.Lock()
			defer ruleOverride.lock.Unlock()

			// TODO:
			// I somehow need the key values here as well...
			// this is the 2nd time i need this info (also in willow). And just saving it off here seems like a
			// waste of memory. Ideally we could iterate over everything and pass in the key vaules that make up a pair?
			overrides = append(overrides, v1limiter.Override{KeyValues: ruleOverride.keyValues, Limit: ruleOverride.limit})

			return true
		}

		_ = r.overrides.Query(datatypes.AssociatedKeyValuesQuery{}, onPagiination)
	}

	ruleResponse := &v1limiter.Rule{
		Name:      r.ruleModel.Name,
		GroupBy:   r.ruleModel.GroupBy,
		Limit:     r.ruleModel.Limit,
		Overrides: overrides,
	}

	return ruleResponse
}
