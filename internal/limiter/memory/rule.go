package memory

import (
	"sync"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1limiter"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"github.com/DanLavine/willow/pkg/models/query"
	"go.uber.org/zap"
)

type rule struct {
	ruleModelLock *sync.RWMutex
	ruleModel     *v1limiter.RuleRequest

	// all values in the overrides are of type 'ruleOverride'
	overrides btreeassociated.BTreeAssociated
}

type ruleOverride struct {
	lock *sync.RWMutex

	keyValues datatypes.StringMap
	limit     uint64
}

func NewRule(ruleModel *v1limiter.RuleRequest) *rule {
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

func (r *rule) SetOverride(logger *zap.Logger, override *v1limiter.RuleOverride) *api.Error {
	logger = logger.Named("SetOverride")

	// set an override iff the tags aare valid
	if !r.ruleModel.Seletion.MatchTags(override.KeyValues) {
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
		ruleOverride := item.(*ruleOverride)
		ruleOverride.lock.Lock()
		defer ruleOverride.lock.Unlock()

		ruleOverride.limit = override.Limit
	}

	if err := r.overrides.CreateOrFind(override.KeyValues, onCreate, onFind); err != nil {
		logger.Error("failed to CreateOrFind a rule override", zap.Error(err))
		return errors.InternalServerError.With("", err.Error())
	}

	return nil
}

func (r *rule) DeleteOverride(logger *zap.Logger, override *v1limiter.RuleOverride) *api.Error {
	logger = logger.Named("DeleteOverride")

	if err := r.overrides.Delete(override.KeyValues, func(_ any) bool { return true }); err != nil {
		logger.Error("failed to delete a rule override", zap.Error(err))
		return errors.InternalServerError.With("", err.Error())
	}

	return nil
}

func (r *rule) FindLimit(logger *zap.Logger, keyValues datatypes.StringMap) uint64 {
	r.ruleModelLock.RLock()
	limit := r.ruleModel.Limit
	r.ruleModelLock.RUnlock()

	onFind := func(item any) {
		ruleOverride := item.(*ruleOverride)
		ruleOverride.lock.RLock()
		defer ruleOverride.lock.RUnlock()

		limit = ruleOverride.limit
	}

	// ignore these errors... should make it so it just panics
	_ = r.overrides.Find(keyValues, onFind)

	return limit
}

func (r *rule) TagsMatch(logger *zap.Logger, keyValues datatypes.StringMap) bool {
	// ensure that all the "group by" keys exists
	for _, key := range r.ruleModel.GroupBy {
		if _, ok := keyValues[key]; !ok {
			return false
		}
	}

	// ensure that the selection doesn't filter out the request
	return r.ruleModel.Seletion.MatchTags(keyValues)
}

func (r *rule) Lock() {
	r.ruleModelLock.Lock()
}

func (r *rule) Unlock() {
	r.ruleModelLock.Unlock()
}

func (r *rule) GenerateQuery(keyValues datatypes.StringMap) query.Select {
	selectQuery := query.Select{
		Where: &query.Query{
			KeyValues: map[string]query.Value{},
		},
	}

	for _, key := range r.ruleModel.GroupBy {
		value := keyValues[key]
		selectQuery.Where.KeyValues[key] = query.Value{Value: &value, ValueComparison: query.EqualsPtr()}
	}

	return selectQuery
}

func (r *rule) GetRuleResponse(includeOverrides bool) *v1limiter.RuleResponse {
	r.ruleModelLock.RLock()
	defer r.ruleModelLock.RUnlock()

	var overrides []v1limiter.RuleOverrideResponse
	if includeOverrides {
		onPagiination := func(item any) bool {
			ruleOverride := item.(*ruleOverride)
			ruleOverride.lock.Lock()
			defer ruleOverride.lock.Unlock()

			// TODO:
			// I somehow need the key values here as well...
			// this is the 2nd time i need this info (also in willow). And just saving it off here seems like a
			// waste of memory. Ideally we could iterate over everything and pass in the key vaules that make up a pair?
			overrides = append(overrides, v1limiter.RuleOverrideResponse{KeyValues: ruleOverride.keyValues, Limit: ruleOverride.limit})

			return true
		}

		_ = r.overrides.Query(query.Select{}, onPagiination)
	}

	ruleResponse := &v1limiter.RuleResponse{
		Name:          r.ruleModel.Name,
		GroupBy:       r.ruleModel.GroupBy,
		Limit:         r.ruleModel.Limit,
		RuleOverrides: overrides,
	}

	return ruleResponse
}
