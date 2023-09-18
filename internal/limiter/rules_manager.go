package limiter

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/DanLavine/willow/internal/datastructures/btree"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/errors"
	"github.com/DanLavine/willow/internal/limiter/memory"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/api/v1limiter"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"
)

// TODO (quality of life improvement): On the iterations, it would be nice to get all the "keys" that make up a collection.
// this way we can easily report which rule "name" is blocking a request. Also the "overrides key values" in the rule's themselves
// wouldn't need to save that information which is just a lot of extra space...

// Handles CRUD backend logic for Limit operations
type RulesManager interface {
	// TODO: add an API for listing current limits

	// create
	CreateGroupRule(logger *zap.Logger, createRequest *v1limiter.RuleRequest) *api.Error

	// read
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
	rules btree.BTree

	// all possible tag groups and their counters
	counters btreeassociated.BTreeAssociated
}

func NewRulesManger() *rulesManger {
	tree, err := btree.NewThreadSafe(2)
	if err != nil {
		panic(err)
	}

	return &rulesManger{
		rules:    tree,
		counters: btreeassociated.NewThreadSafe(),
	}
}

// Create new group rule operation
func (rm *rulesManger) CreateGroupRule(logger *zap.Logger, createRequest *v1limiter.RuleRequest) *api.Error {
	logger = logger.Named("CreateGroupRule")
	var returnErr *api.Error

	onCreate := func() any {
		return memory.NewRule(createRequest)
	}

	onFind := func(_ any) {
		logger.Error("failed to create new rule set", zap.String("name", createRequest.Name))
		returnErr = (&api.Error{Message: "rule already exists", StatusCode: http.StatusConflict}).With(fmt.Sprintf("name %s to not be in use", createRequest.Name), "")
	}

	if err := rm.rules.CreateOrFind(datatypes.String(createRequest.Name), onCreate, onFind); err != nil {
		logger.Error("failed to create or find a new rule", zap.Error(err))
		return errors.InternalServerError
	}

	return returnErr
}

func (rm *rulesManger) FindRule(logger *zap.Logger, name string) Rule {
	//logger = logger.Named("FindGroupRule")
	var limiterRules Rule

	onFind := func(item any) {
		limiterRules = item.(Rule)
	}

	_ = rm.rules.Find(datatypes.String(name), onFind)
	return limiterRules
}

// list all group rules
func (rm *rulesManger) ListRules(logger *zap.Logger) []Rule {
	//logger = logger.Named("ListGroupRules")
	limiterRules := []Rule{}

	onFind := func(item any) bool {
		rule := item.(Rule)
		limiterRules = append(limiterRules, rule)
		return true
	}

	_ = rm.rules.Iterate(onFind)

	return limiterRules
}

func (rm *rulesManger) DeleteGroupRule(logger *zap.Logger, name string) {
	//logger = logger.Named("DeleteGroupRule")
	canDelte := func(item any) bool {
		// nothing to block deleting a group rule. So just automatically delete it
		return true
	}

	_ = rm.rules.Delete(datatypes.String(name), canDelte)
}

// Increment trys to add to a group of key value pairs, and blocks if any rules have hit the limit
func (rm *rulesManger) Increment(logger *zap.Logger, increment *v1limiter.RuleCounterRequest) *api.Error {
	logger = logger.Named("Increment")
	var rules []Rule
	defer func() {
		// unlock all the rules
		for _, rule := range rules {
			rule.Unlock()
		}
	}()

	limitReached := false
	OnFindPagination := func(item any) bool {
		rule := item.(Rule)
		ruleLimit := rule.FindLimit(logger, increment.KeyValues)
		var totalCount uint64

		// if the rule matches the tags we are incrementing, we need to check that the limits are not already reached
		if rule.TagsMatch(logger, increment.KeyValues) {
			rule.Lock()
			rules = append(rules, rule)

			// callback to count all known values
			countPagination := func(item any) bool {
				counter := item.(*atomic.Uint64)
				totalCount += counter.Load()

				// chek if we reached the rule limit
				return totalCount <= ruleLimit
			}

			_ = rm.counters.Query(rule.GenerateQuery(increment.KeyValues), countPagination)
		}

		// the rule we are on has failed, need to setup a trigger for the decrement case
		if totalCount >= ruleLimit {
			limitReached = true
			return false
		}

		return true
	}

	_ = rm.rules.Iterate(OnFindPagination)

	// at this point, the last rule is blocking because the limit has been reached
	if limitReached {
		return &api.Error{Message: "Unable to process limit request. The limits are already reached", StatusCode: http.StatusLocked}
	}

	// there are no rules, so just record the value incrementation
	onCreate := func() any {
		counter := new(atomic.Uint64)
		counter.Add(1)
		return counter
	}

	// update the counter
	onFind := func(item any) {
		counter := item.(*atomic.Uint64)
		counter.Add(1)
	}

	_ = rm.counters.CreateOrFind(increment.KeyValues, onCreate, onFind)

	return nil
}

// Increment trys to add to a group of key value pairs, and blocks if any rules have hit the limit
func (rm *rulesManger) Decrement(logger *zap.Logger, decrement *v1limiter.RuleCounterRequest) {
	//logger = logger.Named("Decrement")
	canDelete := func(item any) bool {
		counter := item.(*atomic.Uint64)
		counter.Add(^uint64(0))

		return counter.Load() == 0
	}

	_ = rm.counters.Delete(decrement.KeyValues, canDelete)
}
