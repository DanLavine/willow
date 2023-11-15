package limiter

import (
	"fmt"
	"net/http"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

type nameRule struct {
	name string
}

func (nr nameRule) Less(item any) bool {
	return nr.name < item.(nameRule).name
}

// Handles CRUD backend logic for Limit operations
type RulesManager interface {
	// TODO: add an API for listing current limits

	// create
	Create(logger *zap.Logger, rule *v1limiter.Rule) *api.Error

	// update
	Update(logger *zap.Logger, name string, update *v1limiter.RuleUpdate) *api.Error

	// read
	Get(logger *zap.Logger, name string, includeOverrides bool) *v1limiter.Rule
	FindRule(logger *zap.Logger, name string) rules.Rule
	ListRules(logger *zap.Logger) []rules.Rule

	// delete operations
	Delete(logger *zap.Logger, name string) *api.Error

	// override operations
	CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *api.Error
	DeleteOverride(logger *zap.Logger, ruleName string, overrideName string) *api.Error

	// increment a partiular group of tags
	Increment(logger *zap.Logger, increment *v1limiter.RuleCounterRequest) *api.Error

	// decrement a particular group of tags
	Decrement(logger *zap.Logger, decrement *v1limiter.RuleCounterRequest)
}

type rulesManger struct {
	// rule constructor for creating and managing rules in a proper configuration
	ruleConstructor rules.RuleConstructor

	// all possible rules a user had created
	rules btreeassociated.BTreeAssociated

	// all possible tag groups and their counters
	counters btreeassociated.BTreeAssociated
}

func NewRulesManger(ruleConstructor rules.RuleConstructor) *rulesManger {
	return &rulesManger{
		ruleConstructor: ruleConstructor,
		rules:           btreeassociated.NewThreadSafe(),
		counters:        btreeassociated.NewThreadSafe(),
	}
}

// Create new group rule operation
func (rm *rulesManger) Create(logger *zap.Logger, rule *v1limiter.Rule) *api.Error {
	logger = logger.Named("CreateGroupRule")
	onCreate := func() any {
		return rm.ruleConstructor.New(rule)
	}

	// record the name as a value. This will make the group by + name unique keys
	keyValues := btreeassociated.KeyValues{
		datatypes.Custom(nameRule{name: "name"}): datatypes.String(rule.Name),
	}
	for _, groupByName := range rule.GroupBy {
		keyValues[datatypes.String(groupByName)] = datatypes.Nil()
	}

	// create the rule only if the name is free
	if err := rm.rules.CreateWithID(rule.Name, keyValues, onCreate); err != nil {
		switch err {
		case btreeassociated.ErrorCreateFailureKeyValuesExist, btreeassociated.ErrorAssociatedIDAlreadyExists:
			logger.Warn("failed to create or find a new rule", zap.Error(err))
			return (&api.Error{Message: "failed to create rule.", StatusCode: http.StatusUnprocessableEntity}).With("name to not be in use", "name is already in use by another rule")
		default:
			logger.Error("failed to create or find a new rule", zap.Error(err))
			return (&api.Error{Message: "failed to create, unexpected error.", StatusCode: http.StatusInternalServerError}).With("", err.Error())
		}
	}

	return nil
}

// Get a group rule by name
func (rm *rulesManger) Get(logger *zap.Logger, name string, includeOverrides bool) *v1limiter.Rule {
	logger = logger.Named("Get").With(zap.String("name", name))

	var rule *v1limiter.Rule
	onFind := func(item any) {
		rule = item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule).Get(includeOverrides)
	}

	if err := rm.rules.FindByAssociatedID(name, onFind); err != nil {
		logger.Error("failed to lokup rule.", zap.Error(err))
	}

	if rule == nil {
		logger.Warn("failed to find rule by AssociatedID.")
	}

	return rule
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

func (rm *rulesManger) FindRule(logger *zap.Logger, name string) rules.Rule {
	//logger = logger.Named("FindGroupRule")
	var limiterRules rules.Rule

	//onFind := func(item any) {
	//	limiterRules = item.(Rule)
	//}
	//
	//_ = rm.rules.Find(datatypes.String(name), onFind)
	return limiterRules
}

// list all group rules
func (rm *rulesManger) ListRules(logger *zap.Logger) []rules.Rule {
	//logger = logger.Named("ListGroupRules")
	//limiterRules := []Rule{}
	//
	//onFind := func(item any) bool {
	//	rule := item.(Rule)
	//	limiterRules = append(limiterRules, rule)
	//	return true
	//}
	//
	//_ = rm.rules.Iterate(onFind)
	//
	//return limiterRules

	return nil
}

// Increment trys to add to a group of key value pairs, and blocks if any rules have hit the limit
//
// NOTE: This is a stupid hard Horizontaly scaling issue without a 3rd party lock on things. How do we quickly know if a limit would succeed.
//
// This is NOT THREAD SAFE! There is some odd logic around needing to update the iterator + locking values that might already exist
//
// Adding a "Sorted [asc|dec]" in the BTreeAssociated I don't think fixes this, because we loop through values as soon as we find
// them, not after we grab all the IDs... This is also a stupid hard problem in how I am thinking about Horizontal Scaling
func (rm *rulesManger) Increment(logger *zap.Logger, increment *v1limiter.RuleCounterRequest) *api.Error {
	//logger = logger.Named("Increment")
	//
	//// This is better since there are no locks, but there is still an issue, where 2+ different requests trying to update different values
	//// cause any number of requests to fail when some should succeed.
	////
	//// I.E. have a "group_by = 'key1'"
	//// * in parallel 20 requests all add a different 'key = [index]' value.
	//// * the searches on the 'OnFindPagination' will all fail if all inserts succeed before any queryes are made.
	////
	//// This is why I initialy had the lock on the 'rule'. But unless all the rule's IDs are found at the same time,
	//// doing any sorting is pointless. This problem will come back around when trying to make any number of tag groups that
	//// are split across different nodes and we want to limit the # of items in a single queue.
	////
	//// What I had actually would work before since I'm locking on the "iterate"  operation for "rules" which is in a guranteed order
	//// But it is slow as shit since we do an exlusive lock for all updates which is insane... This needs to be thought about more
	//
	///*
	//	Rule 1: Group_By ["key1"], Limit: 1
	//	Rule 2: Group_By ["key3"], Limit: 5
	//
	//	if a request has ["key1", "key2", "key3"] -> both rules
	//	if a request has ["key1"] -> first rule
	//	if a request has ["key2", "key3"] -> 2nd rule
	//
	//	Issue is I don't know what rules a key value grouping belongs to.
	//
	//	...
	//
	//	Solutions
	//
	//	I could grab "locks" for all possible key value combination? This would gurantess that any other possible rules for the
	//	same tags are waiting till one collection goes through?
	//
	//*/
	//
	//// 1. always perform an insert/update to the counter
	//var initialCount uint64
	//var initialCounter *atomic.Uint64
	//
	//// there are no rules, so just record the value incrementation
	//onCreate := func() any {
	//	counter := new(atomic.Uint64)
	//	initialCounter = counter
	//	initialCount = counter.Add(1)
	//
	//	return counter
	//}
	//
	//// update the counter
	//onFind := func(item any) {
	//	counter := item.(*btreeassociated.AssociatedKeyValues).Value().(*atomic.Uint64)
	//	initialCounter = counter
	//	initialCount = counter.Add(1)
	//}
	//
	//_, _ = rm.counters.CreateOrFind(btreeassociated.ConverDatatypesKeyValues(increment.KeyValues), onCreate, onFind)
	//
	//// 2. sort through all the rules to know if we are at the limit
	//limitReached := false
	//OnFindPagination := func(item any) bool {
	//	rule := item.(Rule)
	//	totalCount := initialCount
	//
	//	// if the rule matches the tags we are incrementing, we need to check that the limits are not already reached
	//	if rule.TagsMatch(logger, increment.KeyValues) {
	//		ruleLimit := rule.FindLimit(logger, increment.KeyValues)
	//
	//		// callback to count all known values
	//		countPagination := func(item any) bool {
	//			counter := item.(*btreeassociated.AssociatedKeyValues).Value().(*atomic.Uint64)
	//			if initialCounter != counter {
	//				totalCount += counter.Load()
	//			}
	//
	//			// chek if we reached the rule limit
	//			return totalCount < ruleLimit
	//		}
	//
	//		_ = rm.counters.Query(rule.GenerateQuery(increment.KeyValues), countPagination)
	//
	//		// the rule we are on has failed, need to setup a trigger for the decrement case
	//		if totalCount > ruleLimit {
	//			limitReached = true
	//			return false
	//		}
	//	}
	//
	//	return true
	//}
	//
	//_ = rm.rules.Iterate(OnFindPagination)
	//
	//// at this point, the last rule is blocking because the limit has been reached
	//if limitReached {
	//	// need to decrement the inital value since we failed
	//	canDelete := func(item any) bool {
	//		counter := item.(*btreeassociated.AssociatedKeyValues).Value().(*atomic.Uint64)
	//		counter.Add(^uint64(0))
	//
	//		return counter.Load() == 0
	//	}
	//
	//	_ = rm.counters.Delete(btreeassociated.ConverDatatypesKeyValues(increment.KeyValues), canDelete)
	//
	//	return &api.Error{Message: "Unable to process limit request. The limits are already reached", StatusCode: http.StatusLocked}
	//}

	return nil
}

// Increment trys to add to a group of key value pairs, and blocks if any rules have hit the limit
func (rm *rulesManger) Decrement(logger *zap.Logger, decrement *v1limiter.RuleCounterRequest) {
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
