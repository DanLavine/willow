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
	queryFilter   datatypes.AssociatedKeyValuesQuery
	limit         uint64

	// all values in the overrides are of type 'ruleOverride'
	overrides btreeassociated.BTreeAssociated
}

type ruleOverride struct {
	lock *sync.RWMutex

	limit uint64
}

func NewRule(ruleModel *v1limiter.Rule) *rule {
	return &rule{
		ruleModelLock: new(sync.RWMutex),
		name:          ruleModel.Name,
		groupBy:       ruleModel.GroupBy,
		queryFilter:   ruleModel.QueryFilter,
		limit:         ruleModel.Limit,
		overrides:     btreeassociated.NewThreadSafe(),
	}
}

// Get converts a rule to an API response.
//
// DSL TODO: Do I want to have some sor of actual query for the overrides to include?
// perhaps at some point, but I don't think this is usefull right now other than validation/list everything
//
//	PARAMS:
//	- includeOverrodes - iff true, will also include any rule overrides. This can be a SLOW operation.
func (r *rule) Get(includeOverrides bool) *v1limiter.Rule {
	r.ruleModelLock.RLock()
	defer r.ruleModelLock.RUnlock()

	var overrides []v1limiter.Override

	if includeOverrides {
		onPagiination := func(item any) bool {
			associatedKeyValues := item.(*btreeassociated.AssociatedKeyValues)

			ruleOverride := item.(*btreeassociated.AssociatedKeyValues).Value().(*ruleOverride)
			ruleOverride.lock.Lock()
			defer ruleOverride.lock.Unlock()

			overrides = append(overrides, v1limiter.Override{
				Name:      associatedKeyValues.AssociatedID(),
				KeyValues: associatedKeyValues.KeyValues().StripAssociatedID().RetrieveStringDataType(),
				Limit:     ruleOverride.limit,
			})

			return true
		}

		// should not error. That only happens on param validation
		if err := r.overrides.Query(datatypes.AssociatedKeyValuesQuery{}, onPagiination); err != nil {
			panic(err)
		}
	}

	ruleResponse := &v1limiter.Rule{
		Name:        r.name,
		GroupBy:     r.groupBy,
		Limit:       r.limit,
		QueryFilter: r.queryFilter,
		Overrides:   overrides,
	}

	return ruleResponse
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

	// set an override iff the tags are valid
	if !r.queryFilter.MatchTags(override.KeyValues) {
		return api.NotAcceptable.With("the provided keys values to match the rule query", "provided will never be found by the rule query")
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

func (r *rule) FindLimit(logger *zap.Logger, keyValues datatypes.KeyValues) uint64 {
	r.ruleModelLock.RLock()
	limit := r.limit
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
	for _, key := range r.groupBy {
		if _, ok := keyValues[key]; !ok {
			return false
		}
	}

	// ensure that the selection doesn't filter out the request
	return r.queryFilter.MatchTags(keyValues)
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
