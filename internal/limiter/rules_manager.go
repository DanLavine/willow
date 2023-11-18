package limiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/limiter/counters"
	"github.com/DanLavine/willow/internal/limiter/rules"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

// Handles CRUD backend logic for Limit operations
type RulesManager interface {
	// create
	Create(logger *zap.Logger, rule *v1limiter.RuleRequest) *api.Error

	// update
	Update(logger *zap.Logger, name string, update *v1limiter.RuleUpdate) *api.Error

	// read
	Get(logger *zap.Logger, name string, query *v1limiter.RuleQuery) *v1limiter.RuleResponse
	List(logger *zap.Logger, query *v1limiter.RuleQuery) (v1limiter.Rules, *api.Error)

	// delete operations
	Delete(logger *zap.Logger, name string) *api.Error

	// override operations
	CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *api.Error
	DeleteOverride(logger *zap.Logger, ruleName string, overrideName string) *api.Error

	// increment a partiular group of tags
	IncrementKeyValues(logger *zap.Logger, increment *v1limiter.Counter) *api.Error

	// decrement a particular group of tags
	DecrementKeyValues(logger *zap.Logger, decrement *v1limiter.Counter)
}

type rulesManger struct {
	// locker client to ensure that all locks are respected
	lockerClient lockerclient.LockerClient

	// rule constructor for creating and managing rules in a proper configuration
	ruleConstructor rules.RuleConstructor

	// all possible rules a user had created
	rules btreeassociated.BTreeAssociated

	// all possible tag groups and their counters
	counters btreeassociated.BTreeAssociated
}

func NewRulesManger(ruleConstructor rules.RuleConstructor, lockerClient lockerclient.LockerClient) *rulesManger {
	return &rulesManger{
		lockerClient:    lockerClient,
		ruleConstructor: ruleConstructor,
		rules:           btreeassociated.NewThreadSafe(),
		counters:        btreeassociated.NewThreadSafe(),
	}
}

// Create new group rule operation
func (rm *rulesManger) Create(logger *zap.Logger, rule *v1limiter.RuleRequest) *api.Error {
	logger = logger.Named("Create")
	onCreate := func() any {
		return rm.ruleConstructor.New(rule)
	}

	// record the name as a value. This will make the group by + name unique keys
	keyValues := btreeassociated.KeyValues{}
	for _, groupByName := range rule.GroupBy {
		keyValues[datatypes.String(groupByName)] = datatypes.String("")
	}

	// create the rule only if the name is free
	if err := rm.rules.CreateWithID(rule.Name, keyValues, onCreate); err != nil {
		switch err {
		case btreeassociated.ErrorCreateFailureKeyValuesExist:
			logger.Warn("failed to create new rule because keys exist", zap.Error(err))
			return (&api.Error{Message: "failed to create rule.", StatusCode: http.StatusUnprocessableEntity}).With("group by keys to not be in use", "group by keys are in use by another rule")
		case btreeassociated.ErrorAssociatedIDAlreadyExists:
			logger.Warn("failed to create new rule because name exist", zap.Error(err))
			return (&api.Error{Message: "failed to create rule.", StatusCode: http.StatusUnprocessableEntity}).With("name to not be in use", "name is already in use by another rule")
		default:
			logger.Error("failed to create or find a new rule", zap.Error(err))
			return (&api.Error{Message: "failed to create, unexpected error.", StatusCode: http.StatusInternalServerError}).With("", err.Error())
		}
	}

	return nil
}

// Get a group rule by name
func (rm *rulesManger) Get(logger *zap.Logger, name string, query *v1limiter.RuleQuery) *v1limiter.RuleResponse {
	logger = logger.Named("Get").With(zap.String("name", name))

	var rule *v1limiter.RuleResponse
	onFind := func(item any) {
		rule = item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule).Get(query)
	}

	if err := rm.rules.FindByAssociatedID(name, onFind); err != nil {
		logger.Error("failed to lokup rule.", zap.Error(err))
	}

	if rule == nil {
		logger.Warn("failed to find rule by AssociatedID.")
	}

	return rule
}

// list all group rules that match the provided key values
//
// Can also include the overrides
func (rm *rulesManger) List(logger *zap.Logger, query *v1limiter.RuleQuery) (v1limiter.Rules, *api.Error) {
	logger = logger.Named("List")
	foundRules := v1limiter.Rules{}

	onFindMatchingRule := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
		rule := associatedKeyValues.Value().(rules.Rule)
		foundRules = append(foundRules, rule.Get(query))

		return true
	}

	switch query.KeyValues {
	case nil:
		if err := rm.rules.Query(datatypes.AssociatedKeyValuesQuery{}, onFindMatchingRule); err != nil {
			logger.Error("faild to query for rules", zap.Error(err))
			return v1limiter.Rules{}, &api.Error{Message: "Internal server error", StatusCode: http.StatusInternalServerError}
		}
	default:
		// special match logic. we alwys need to look for empty strings as a 'Select All' operation
		keyValues := btreeassociated.KeyValues{}
		for key, _ := range *query.KeyValues {
			keyValues[datatypes.String(key)] = datatypes.String("")
		}

		// these need to be converted to empty string. duh
		if err := rm.rules.MatchPermutations(keyValues, onFindMatchingRule); err != nil {
			logger.Error("faild to match for rules", zap.Error(err))
			return v1limiter.Rules{}, &api.Error{Message: "Internal server error", StatusCode: http.StatusInternalServerError}
		}
	}

	return foundRules, nil
}

// Update a rule by name
func (rm *rulesManger) Update(logger *zap.Logger, name string, update *v1limiter.RuleUpdate) *api.Error {
	logger = logger.Named("Update").With(zap.String("rule_name", name))

	found := false
	onFind := func(item any) {
		found = true
		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		rule.Update(logger, update)
	}

	if err := rm.rules.FindByAssociatedID(name, onFind); err != nil {
		logger.Error("failed to find rule by AssociatedID because of an internal server error", zap.Error(err))
		return &api.Error{Message: "failed to find rule by name because of an internal server error", StatusCode: http.StatusInternalServerError}
	}

	if !found {
		logger.Warn("failed to find rule by AssociatedID")
		return (&api.Error{Message: "failed to find rule by name", StatusCode: http.StatusUnprocessableEntity}).With(fmt.Sprintf("name %s", name), "no rule found by that name")
	}

	return nil
}

// Delete a rule by name
func (rm *rulesManger) Delete(logger *zap.Logger, name string) *api.Error {
	logger = logger.Named("DeleteGroupRule").With(zap.String("rule_name", name))

	deleteCalled := false
	var cascadeError *api.Error
	canDelete := func(item any) bool {
		deleteCalled = true

		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		cascadeError = rule.CascadeDeletion(logger)

		if cascadeError == nil {
			logger.Debug("deleted rule")
			return true
		}

		logger.Warn("faled to deleted rule")

		return false
	}

	if err := rm.rules.DeleteByAssociatedID(name, canDelete); err != nil {
		logger.Error("failed to lookup rule for deletion", zap.String("name", name), zap.Error(err))
		return &api.Error{Message: "failed to delete rule by name", StatusCode: http.StatusInternalServerError}
	}

	if !deleteCalled {
		logger.Warn("failed to find rule for deletion")
	}

	return cascadeError
}

// Create an override for a rule by name
func (rm *rulesManger) CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *api.Error {
	logger = logger.Named("CreateOverride")

	var overrideErr *api.Error
	onFind := func(item any) {
		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		overrideErr = rule.SetOverride(logger, override)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, onFind); err != nil {
		logger.Error("failed to find rule by associatedid", zap.String("name", ruleName), zap.Error(err))
		return (&api.Error{Message: "failed to find rule by name", StatusCode: http.StatusUnprocessableEntity}).With(fmt.Sprintf("name %s", ruleName), "no rule found by that name")
	}

	return overrideErr
}

// Delete an override
func (rm *rulesManger) DeleteOverride(logger *zap.Logger, ruleName string, overrideName string) *api.Error {
	logger = logger.Named("DeleteOverride")

	var overrideErr *api.Error
	onFind := func(item any) {
		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		overrideErr = rule.DeleteOverride(logger, overrideName)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, onFind); err != nil {
		logger.Error("failed to find rule by AssociatedID", zap.String("name", ruleName), zap.Error(err))
		return (&api.Error{Message: "failed to find rule by name", StatusCode: http.StatusUnprocessableEntity}).With(fmt.Sprintf("name %s", ruleName), "no rule found by that name")
	}

	return overrideErr
}

// Increment trys to add to a group of key value pairs, and blocks if any rules have hit the limit
//
// NOTE: This is a stupid hard Horizontaly scaling issue without a 3rd party lock on things. How do we quickly know if a limit would succeed.
//
// This is NOT THREAD SAFE! There is some odd logic around needing to update the iterator + locking values that might already exist
//
// Adding a "Sorted [asc|dec]" in the BTreeAssociated I don't think fixes this, because we loop through values as soon as we find
// them, not after we grab all the IDs... This is also a stupid hard problem in how I am thinking about Horizontal Scaling
func (rm *rulesManger) IncrementKeyValues(logger *zap.Logger, increment *v1limiter.Counter) *api.Error {
	logger = logger.Named("IncrementKeyValues")

	// 1. query the rules that match the tags for our key values and record all the group by with their limit
	var foundRules []rules.Rule
	onFindMatchingRule := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
		rule := associatedKeyValues.Value().(rules.Rule)
		foundRules = append(foundRules, rule)

		return true
	}

	if err := rm.rules.Query(keyValuesToRuleQuery(increment.KeyValues), onFindMatchingRule); err != nil {
		return nil
	}

	// 2.a. if there are no rules, then we can just increment the counter on the key values and return
	if len(foundRules) == 0 {
		return nil
	}

	// 2.b. there are rules with limits. We need to grab a lock for all matched rules
	var lockerLocks []lockerclient.Lock
	defer func() { // ensure we release all the locks at the end
		for _, lockerLock := range lockerLocks {
			err := lockerLock.Release()
			if err != nil {
				logger.Error("failed to release locker service lock", zap.Error(err))
			}
		}
	}()

	// grab a lock for all the grouped key values we care about with the Increment's KeyValues
	for _, keys := range rules.SortRulesGroupBy(foundRules) {
		// setup the group to lock
		lockKeyValues := datatypes.KeyValues{}
		for _, key := range keys {
			lockKeyValues[key] = increment.KeyValues[key]
		}

		// obtain the required lock
		lockerLock, err := rm.lockerClient.ObtainLock(context.Background(), lockKeyValues, 10*time.Second)
		if err != nil {
			logger.Error("failed to obtain a lock from the locker service", zap.Strings("keys", keys), zap.Error(err))
			return &api.Error{Message: "Internal server error", StatusCode: http.StatusInternalServerError}
		}

		// don't need to add the lock 2x if we already obtained one with the same key values
		addLock := true
		for _, lock := range lockerLocks {
			if lock == lockerLock {
				addLock = false
				break
			}
		}

		if addLock {
			lockerLocks = append(lockerLocks, lockerLock)
		}
	}

	// 3. for each rule, find the limt and count all possible records
	for _, rule := range foundRules {
		limit, err := rule.FindLimit(logger, increment.KeyValues)
		if err != nil {
			return err
		}

		counter := uint64(0)
		onFindPagination := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
			counter += associatedKeyValues.Value().(counters.Counter).Count
			return true
		}

		query := datatypes.AssociatedKeyValuesQuery{}
		for _, key := range rule.GetGroupByKeys() {
			value := increment.KeyValues[key]
			query.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &value, ValueComparison: datatypes.EqualsPtr()}
		}

		if err := rm.counters.Query(query, onFindPagination); err != nil {
			logger.Error("failed to query counters", zap.Error(err))
			return &api.Error{Message: "internal server error", StatusCode: http.StatusInternalServerError}
		}

		if counter >= limit {
			return &api.Error{Message: fmt.Sprintf("Limit reached for rule: %s", rule.Name()), StatusCode: http.StatusConflict}
		}
	}

	// 4. we are under the limit, so update or create the requested tags
	createCounter := func() any {
		return counters.Counter{Count: 1}
	}

	incrementCounter := func(item any) {
		counter := item.(*btreeassociated.AssociatedKeyValues).Value().(counters.Counter)
		counter.Count += 1
	}

	rm.counters.CreateOrFind(btreeassociated.ConverDatatypesKeyValues(increment.KeyValues), createCounter, incrementCounter)

	return nil
}

// Increment trys to add to a group of key value pairs, and blocks if any rules have hit the limit
func (rm *rulesManger) DecrementKeyValues(logger *zap.Logger, decrement *v1limiter.Counter) {
	//logger = logger.Named("Decrement")
	//canDelete := func(item any) bool {
	//	counter := item.(*btreeassociated.AssociatedKeyValues).Value().(*atomic.Uint64)
	//	counter.Add(^uint64(0))
	//
	//	return counter.Load() == 0
	//}
	//
	//_ = rm.counters.Delete(btreeassociated.ConverDatatypesKeyValues(decrement.KeyValues), canDelete)
}

// TODO DSL: this can be optimized on the BtreeAssociated. Wher I could do a lookup all keys that these match for. then
// perform a filter for matching the values. But for now, this is fine to prove out the API.
func keyValuesToRuleQuery(keyValues datatypes.KeyValues) datatypes.AssociatedKeyValuesQuery {
	query := datatypes.AssociatedKeyValuesQuery{}

	// generate all possible key value tag pairs
	keyValueCombinations := keyValues.GenerateTagPairs()

	// setup a query to find any rules have the tags to match
	exists := true
	for _, keyValueCombo := range keyValueCombinations {
		orQuery := datatypes.AssociatedKeyValuesQuery{
			KeyValueSelection: &datatypes.KeyValueSelection{
				KeyValues: map[string]datatypes.Value{},
			},
		}

		for key, _ := range keyValueCombo {
			orQuery.KeyValueSelection.KeyValues[key] = datatypes.Value{Exists: &exists}
		}

		query.Or = append(query.Or, orQuery)
	}

	return query
}
