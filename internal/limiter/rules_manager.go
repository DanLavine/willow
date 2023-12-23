package limiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DanLavine/channelops"
	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/limiter/counters"
	"github.com/DanLavine/willow/internal/limiter/rules"
	v1limitermodels "github.com/DanLavine/willow/internal/limiter/v1_limiter_models"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	v1locker "github.com/DanLavine/willow/pkg/models/api/locker/v1"
)

var (
	LimitReached = &errors.ServerError{Message: "Limit has already been reached for requested key values", StatusCode: http.StatusConflict}
)

func errorMissingRuleName(name string) *errors.ServerError {
	return &errors.ServerError{Message: fmt.Sprintf("failed to find rule '%s' by name", name), StatusCode: http.StatusUnprocessableEntity}
}

// Handles CRUD backend logic for Limit operations
type RulesManager interface {
	// create
	Create(logger *zap.Logger, rule *v1limiter.RuleCreateRequest) *errors.ServerError

	// update
	Update(logger *zap.Logger, name string, update *v1limiter.RuleUpdateRquest) *errors.ServerError

	// read
	Get(logger *zap.Logger, name string, query *v1limiter.RuleQuery) *v1limiter.Rule
	List(logger *zap.Logger, query *v1limiter.RuleQuery) (v1limiter.Rules, *errors.ServerError)

	// delete operations
	Delete(logger *zap.Logger, name string) *errors.ServerError

	// override operations
	ListOverrides(logger *zap.Logger, ruleName string, query *v1common.AssociatedQuery) (v1limiter.Overrides, *errors.ServerError)
	CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *errors.ServerError
	DeleteOverride(logger *zap.Logger, ruleName string, overrideName string) *errors.ServerError

	// counter operations
	ListCounters(logger *zap.Logger, query *v1common.AssociatedQuery) (v1limiter.Counters, *errors.ServerError)
	IncrementCounters(logger *zap.Logger, requestContext context.Context, lockerClient lockerclient.LockerClient, increment *v1limiter.Counter) *errors.ServerError
	DecrementCounters(logger *zap.Logger, decrement *v1limiter.Counter) *errors.ServerError
	SetCounters(logger *zap.Logger, setCounters *v1limiter.Counter) *errors.ServerError
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
func (rm *rulesManger) Create(logger *zap.Logger, rule *v1limiter.RuleCreateRequest) *errors.ServerError {
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
			return &errors.ServerError{Message: "failed to create rule. GroupBy keys already in use by another rule", StatusCode: http.StatusUnprocessableEntity}
		case btreeassociated.ErrorAssociatedIDAlreadyExists:
			logger.Warn("failed to create new rule because name exist", zap.Error(err))
			return &errors.ServerError{Message: "failed to create rule. Name is already in use", StatusCode: http.StatusUnprocessableEntity}
		default:
			logger.Error("failed to create or find a new rule", zap.Error(err))
			return &errors.ServerError{Message: "failed to create, unexpected error.", StatusCode: http.StatusInternalServerError}
		}
	}

	return nil
}

// Get a group rule by name
func (rm *rulesManger) Get(logger *zap.Logger, name string, query *v1limiter.RuleQuery) *v1limiter.Rule {
	logger = logger.Named("Get").With(zap.String("name", name))

	var rule *v1limiter.Rule
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
func (rm *rulesManger) List(logger *zap.Logger, query *v1limiter.RuleQuery) (v1limiter.Rules, *errors.ServerError) {
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
			return nil, &errors.ServerError{Message: "Internal server error", StatusCode: http.StatusInternalServerError}
		}
	default:
		// special match logic. we alwys need to look for empty strings as a 'Select All' operation
		// these need to be converted to empty string. duh
		if err := rm.rules.MatchPermutations(keyValuesToRuleQuery(*query.KeyValues), onFindMatchingRule); err != nil {
			logger.Error("faild to match for rules", zap.Error(err))
			return nil, &errors.ServerError{Message: "Internal server error", StatusCode: http.StatusInternalServerError}
		}
	}

	return foundRules, nil
}

// Update a rule by name
func (rm *rulesManger) Update(logger *zap.Logger, name string, update *v1limiter.RuleUpdateRquest) *errors.ServerError {
	logger = logger.Named("Update").With(zap.String("rule_name", name))

	found := false
	onFind := func(item any) {
		found = true
		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		rule.Update(logger, update)
	}

	if err := rm.rules.FindByAssociatedID(name, onFind); err != nil {
		logger.Error("failed to find rule by AssociatedID because of an internal server error", zap.Error(err))
		return &errors.ServerError{Message: "failed to find rule by name because of an internal server error", StatusCode: http.StatusInternalServerError}
	}

	if !found {
		logger.Warn("failed to find rule by AssociatedID")
		return &errors.ServerError{Message: fmt.Sprintf("failed to find rule '%s' by name", name), StatusCode: http.StatusUnprocessableEntity}
	}

	return nil
}

// Delete a rule by name
func (rm *rulesManger) Delete(logger *zap.Logger, name string) *errors.ServerError {
	logger = logger.Named("DeleteGroupRule").With(zap.String("rule_name", name))

	deleteCalled := false
	var cascadeError *errors.ServerError
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
		return &errors.ServerError{Message: "failed to delete rule by name", StatusCode: http.StatusInternalServerError}
	}

	if !deleteCalled {
		logger.Warn("failed to find rule for deletion")
	}

	return cascadeError
}

// Create an override for a rule by name
func (rm *rulesManger) ListOverrides(logger *zap.Logger, ruleName string, query *v1common.AssociatedQuery) (v1limiter.Overrides, *errors.ServerError) {
	logger = logger.Named("ListOverrides")

	found := false
	overrides := v1limiter.Overrides{}
	var overrideErr *errors.ServerError
	onFind := func(item any) {
		found = true
		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		overrides, overrideErr = rule.QueryOverrides(logger, query)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, onFind); err != nil {
		logger.Error("failed to find rule by associatedid", zap.String("name", ruleName), zap.Error(err))
		return overrides, errorMissingRuleName(ruleName)
	}

	if !found {
		overrideErr = &errors.ServerError{Message: fmt.Sprintf("Rule %s not found", ruleName), StatusCode: http.StatusNotFound}
	}

	return overrides, overrideErr
}

// Create an override for a rule by name
func (rm *rulesManger) CreateOverride(logger *zap.Logger, ruleName string, override *v1limiter.Override) *errors.ServerError {
	logger = logger.Named("CreateOverride")

	found := false
	var overrideErr *errors.ServerError
	onFind := func(item any) {
		found = true
		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		overrideErr = rule.SetOverride(logger, override)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, onFind); err != nil {
		logger.Error("failed to find rule by associatedid", zap.String("name", ruleName), zap.Error(err))
		return errorMissingRuleName(ruleName)
	}

	if !found {
		overrideErr = &errors.ServerError{Message: fmt.Sprintf("Rule %s not found", ruleName), StatusCode: http.StatusNotFound}
	}

	return overrideErr
}

// Delete an override
func (rm *rulesManger) DeleteOverride(logger *zap.Logger, ruleName string, overrideName string) *errors.ServerError {
	logger = logger.Named("DeleteOverride")

	found := false
	var overrideErr *errors.ServerError
	onFind := func(item any) {
		found = true
		rule := item.(*btreeassociated.AssociatedKeyValues).Value().(rules.Rule)
		overrideErr = rule.DeleteOverride(logger, overrideName)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, onFind); err != nil {
		logger.Error("failed to find rule by AssociatedID", zap.String("name", ruleName), zap.Error(err))
		return errorMissingRuleName(ruleName)
	}

	if !found {
		overrideErr = &errors.ServerError{Message: fmt.Sprintf("Rule %s not found", ruleName), StatusCode: http.StatusNotFound}
	}

	return overrideErr
}

// List a all counters that match the query
func (rm *rulesManger) ListCounters(logger *zap.Logger, query *v1common.AssociatedQuery) (v1limiter.Counters, *errors.ServerError) {
	logger = logger.Named("ListCounters")

	countersResponse := v1limiter.Counters{}

	onFindPagination := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
		counter := associatedKeyValues.Value().(*counters.Counter)

		newCounter := &v1limiter.Counter{
			KeyValues: associatedKeyValues.KeyValues().StripAssociatedID().RetrieveStringDataType(),
			Counters:  counter.Load(),
		}

		countersResponse = append(countersResponse, newCounter)
		return true
	}

	if err := rm.counters.Query(query.AssociatedKeyValues, onFindPagination); err != nil {
		logger.Error("Failed to query key values", zap.Error(err))
		return countersResponse, errors.InternalServerError
	}

	return countersResponse, nil
}

// Increment trys to add to a group of key value pairs, and returns an error if any rules have hit the limit
func (rm *rulesManger) IncrementCounters(logger *zap.Logger, requestContext context.Context, lockerClient lockerclient.LockerClient, increment *v1limiter.Counter) *errors.ServerError {
	logger = logger.Named("IncrementCounters")

	var foundRules []rules.Rule
	var limitErr *errors.ServerError

	// 1. query the rules that match the tags for our key values and record all the group by with their limit
	allLimits := v1limitermodels.Limits{}
	onFindMatchingRule := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
		rule := associatedKeyValues.Value().(rules.Rule)
		foundRules = append(foundRules, rule)

		// for each rule, find all the limits that need to be checked
		limits, err := rule.FindLimits(logger, increment.KeyValues)
		if err != nil {
			limitErr = err
			return false
		}

		// get group of all the limits for all the rules
		allLimits = append(allLimits, limits...)

		// last index is 0, so just break early
		return limits[len(limits)-1].Limit != 0
	}

	// special match logic. we alwys need to look for empty strings as a 'Select All' operation
	// find all rules that match the permutation of the Increment's KeyValues
	if err := rm.rules.MatchPermutations(keyValuesToRuleQuery(increment.KeyValues), onFindMatchingRule); err != nil {
		logger.Error("Failed to query rules", zap.Error(err))
		return errors.InternalServerError
	}

	// there was an error finding lmits. This shouldn't happen
	if limitErr != nil {
		return limitErr
	}

	// there are no limits, so just accept
	if len(allLimits) != 0 {
		// the last limit has a limit of 0 so bail early
		if len(allLimits) != 0 && allLimits[len(allLimits)-1].Limit == 0 {
			return LimitReached
		}

		// 2. grab a lock for all key values
		lockerLocks := []lockerclient.Lock{}
		defer func() {
			for _, lock := range lockerLocks {
				if err := lock.Release(); err != nil {
					logger.Error("Failed to release lock", zap.Error(err))
				}
			}
		}()

		channelOps, chanReceiver := channelops.NewMergeRead[struct{}](true, requestContext)
		for _, key := range increment.KeyValues.SoretedKeys() {
			// setup the group to lock
			lockKeyValues := &v1locker.LockCreateRequest{
				KeyValues: datatypes.KeyValues{key: increment.KeyValues[key]},
				Timeout:   time.Second,
			}

			// obtain the required lock
			lockerLock, err := lockerClient.ObtainLock(requestContext, lockKeyValues)
			if err != nil {
				logger.Error("failed to obtain a lock from the locker service", zap.Any("key values", lockKeyValues), zap.Error(err))
				return errors.InternalServerError
			}

			// setup monitor for when a lock is released
			lockerLocks = append(lockerLocks, lockerLock)
			if err := channelOps.MergeOrToOne(lockerLock.Done()); err != nil {
				// in this case, something has already been lost
				break
			}
		}

		// add a channel to manually kick. This give a chance for any lost locks to process properly
		successChan := make(chan struct{}, 1)
		successChan <- struct{}{}
		defer close(successChan)

		if err := channelOps.MergeOrToOne(successChan); err != nil {
			// lock is already lost so bail early
			logger.Error("a lock was released unexpedily")
			return errors.InternalServerError
		}

		// ensure that we didn't cancel obtaining any locks by triggering a select. there is small chance that a lock was lost,
		// but that is such a rare race condition I don't see it happening for real.
		_, ok := <-chanReceiver
		if !ok {
			// lost a lock or canceled obtaining the locks
			logger.Error("a lock was released unexpedily")
			return errors.InternalServerError
		}

		// 3. for each limit, count the possible rules that match and ensure that they are under the current limites
		for _, singleLimit := range allLimits {
			query := datatypes.AssociatedKeyValuesQuery{
				KeyValueSelection: &datatypes.KeyValueSelection{KeyValues: map[string]datatypes.Value{}},
			}

			for key, value := range singleLimit.KeyValues {
				// This is spuer important to use, otherwise the address of value is used. so everything will point to the same value which is wrong!
				tmp := value
				query.KeyValueSelection.KeyValues[key] = datatypes.Value{Value: &tmp, ValueComparison: datatypes.EqualsPtr()}
			}

			counter := int64(0)
			onQuery := func(associatedKeyValues *btreeassociated.AssociatedKeyValues) bool {
				counter += associatedKeyValues.Value().(*counters.Counter).Load()
				return counter < int64(singleLimit.Limit) // check to exit query early if this fails
			}

			if err := rm.counters.Query(query, onQuery); err != nil {
				//if err := rm.counters.Query(datatypes.AssociatedKeyValuesQuery{}, onQuery); err != nil {
				logger.Error("Failed to query the current counters", zap.Error(err))
				return &errors.ServerError{Message: "Failed to query the current counters", StatusCode: http.StatusInternalServerError}
			}

			if counter >= int64(singleLimit.Limit) {
				logger.Info("Limit already reached", zap.String("rule name", singleLimit.Name))
				return &errors.ServerError{Message: fmt.Sprintf("Limit has already been reached for rule '%s'", singleLimit.Name), StatusCode: http.StatusConflict}
			}
		}
	}

	// 4. we are under the limit, so update or create the requested tags
	createCounter := func() any {
		return &counters.Counter{Count: atomic.NewInt64(increment.Counters)}
	}

	incrementCounter := func(item any) {
		item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Update(increment.Counters)
	}

	if _, err := rm.counters.CreateOrFind(btreeassociated.ConverDatatypesKeyValues(increment.KeyValues), createCounter, incrementCounter); err != nil {
		logger.Error("Failed to find or update the counter", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}

// Decrement removes a single instance from the key values group. If the total count would become 0, then the
// key values are removed entierly
//
// Decrement is muuch easier than increment because we don't need to ensure any rules validation. So no locks are required
// and we can just decrement the key values directly
func (rm *rulesManger) DecrementCounters(logger *zap.Logger, decrement *v1limiter.Counter) *errors.ServerError {
	logger = logger.Named("DecrementCounters")

	decrementCounter := func(item any) bool {
		count := item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Update(decrement.Counters)
		return count == 0
	}

	if err := rm.counters.Delete(btreeassociated.ConverDatatypesKeyValues(decrement.KeyValues), decrementCounter); err != nil {
		logger.Error("Failed to find or update the counter", zap.Error(err))
		return errors.InternalServerError
	}

	return nil
}

func (rm *rulesManger) SetCounters(logger *zap.Logger, counter *v1limiter.Counter) *errors.ServerError {
	logger = logger.Named("SetCounters")

	if counter.Counters <= 0 {

		// need to remove the key values
		decrementCounter := func(item any) bool {
			return true
		}

		if err := rm.counters.Delete(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), decrementCounter); err != nil {
			logger.Error("Failed to delete the set counters", zap.Error(err))
			return errors.InternalServerError
		}

		return nil
	} else {
		// need to create or set the key values
		createCounter := func() any {
			return &counters.Counter{Count: atomic.NewInt64(counter.Counters)}
		}

		incrementCounter := func(item any) {
			item.(*btreeassociated.AssociatedKeyValues).Value().(*counters.Counter).Set(counter.Counters)
		}

		if _, err := rm.counters.CreateOrFind(btreeassociated.ConverDatatypesKeyValues(counter.KeyValues), createCounter, incrementCounter); err != nil {
			logger.Error("Failed to find or update the set counter", zap.Error(err))
			return errors.InternalServerError
		}
	}

	return nil
}

func keyValuesToRuleQuery(query datatypes.KeyValues) btreeassociated.KeyValues {
	// special match logic. we alwys need to look for empty strings as a 'Select All' operation
	keyValues := btreeassociated.KeyValues{}
	for key, _ := range query {
		keyValues[datatypes.String(key)] = datatypes.String("")
	}

	return keyValues
}
