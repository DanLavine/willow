package limiter

import (
	"fmt"
	"net/http"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/limiter/memory"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

// TODO (quality of life improvement): On the iterations, it would be nice to get all the "keys" that make up a collection.
// this way we can easily report which rule "name" is blocking a request. Also the "overrides key values" in the rule's themselves
// wouldn't need to save that information which is just a lot of extra space...

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
	CreateGroupRule(logger *zap.Logger, rule *v1limiter.Rule) *api.Error

	// read
	Get(logger *zap.Logger, name string, includeOverrides bool) *v1limiter.Rule
	FindRule(logger *zap.Logger, name string) Rule
	ListRules(logger *zap.Logger) []Rule

	// delete operations
	DeleteGroupRule(logger *zap.Logger, name string)

	// increment a partiular group of tags
	Increment(logger *zap.Logger, increment *v1limiter.RuleCounterRequest) *api.Error

	// decrement a particular group of tags
	Decrement(logger *zap.Logger, decrement *v1limiter.RuleCounterRequest)
}

type rulesManger struct {
	// all possible rules a user had created
	rules btreeassociated.BTreeAssociated

	// all possible tag groups and their counters
	counters btreeassociated.BTreeAssociated
}

func NewRulesManger() *rulesManger {
	return &rulesManger{
		rules:    btreeassociated.NewThreadSafe(),
		counters: btreeassociated.NewThreadSafe(),
	}
}

// Create new group rule operation
func (rm *rulesManger) CreateGroupRule(logger *zap.Logger, rule *v1limiter.Rule) *api.Error {
	logger = logger.Named("CreateGroupRule")
	onCreate := func() any {
		return memory.NewRule(rule)
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
		logger.Error("failed to create or find a new rule", zap.Error(err))
		return (&api.Error{Message: "rule already exists", StatusCode: http.StatusUnprocessableEntity}).With(fmt.Sprintf("name %s to not be in use", rule.Name), "")

	}

	return nil
}

// Get a group rule by name
func (rm *rulesManger) Get(logger *zap.Logger, name string, includeOverrides bool) *v1limiter.Rule {
	logger = logger.Named("Get")

	var rule Rule
	onFind := func(item any) {
		rule = item.(*btreeassociated.AssociatedKeyValues).Value().(Rule)
	}

	if err := rm.rules.FindByAssociatedID(name, onFind); err != nil {
		logger.Error("failed to find rule by associatedid", zap.String("name", name), zap.Error(err))
	}

	if rule != nil {
		return rule.GetRuleResponse(includeOverrides)
	}

	return nil
}

func (rm *rulesManger) FindRule(logger *zap.Logger, name string) Rule {
	//logger = logger.Named("FindGroupRule")
	var limiterRules Rule

	//onFind := func(item any) {
	//	limiterRules = item.(Rule)
	//}
	//
	//_ = rm.rules.Find(datatypes.String(name), onFind)
	return limiterRules
}

// list all group rules
func (rm *rulesManger) ListRules(logger *zap.Logger) []Rule {
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

func (rm *rulesManger) DeleteGroupRule(logger *zap.Logger, name string) {
	//logger = logger.Named("DeleteGroupRule")
	//canDelte := func(item any) bool {
	//	// nothing to block deleting a group rule. So just automatically delete it
	//	return true
	//}
	//
	//_ = rm.rules.Delete(datatypes.String(name), canDelte)
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
