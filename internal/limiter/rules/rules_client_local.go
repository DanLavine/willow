package rules

import (
	"context"
	"fmt"
	"net/http"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/limiter/overrides"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

func errorMissingRuleName(name string) *errors.ServerError {
	return &errors.ServerError{Message: fmt.Sprintf("failed to find rule '%s' by name", name), StatusCode: http.StatusNotFound}
}

type localRulesCient struct {
	// rule constructor for creating and managing rules in a proper configuration
	ruleConstructor RuleConstructor

	// all possible rules a user had created
	rules btreeassociated.BTreeAssociated

	// client for rules to interact witth overrides
	overridesClient overrides.OverrideClient
}

func NewLocalRulesClient(ruleConstructor RuleConstructor, overridesClient overrides.OverrideClient) *localRulesCient {
	return &localRulesCient{
		ruleConstructor: ruleConstructor,
		rules:           btreeassociated.NewThreadSafe(),
		overridesClient: overridesClient,
	}
}

//	PARAMETERS:
//	- ctx - context that contains all the reporting tools
//	- rule - create reuest that defines the rule
//
//	RETURNS:
//	- *errors.ServerError - error response for the client if something unexpected happens
//
// Create new Rule
func (rm *localRulesCient) CreateRule(ctx context.Context, rule *v1limiter.RuleCreateRequest) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("CreateRule")

	onCreate := func() any {
		return rm.ruleConstructor.New(rule)
	}

	// record all GroupBy keys as empty strings in the associated tree
	keyValues := datatypes.KeyValues{}
	for _, groupByName := range rule.GroupBy {
		keyValues[groupByName] = datatypes.String("")
	}

	// create the rule only if the name is free
	if err := rm.rules.CreateWithID(rule.Name, keyValues, onCreate); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to create rule. rule is currently being destroy")
			return &errors.ServerError{Message: "failed to create rule. Previous rule is still in the process of destroying", StatusCode: http.StatusConflict}
		case btreeassociated.ErrorKeyValuesAlreadyExists:
			logger.Warn("failed to create new rule because keys exist", zap.Error(err))
			return &errors.ServerError{Message: "failed to create rule. GroupBy keys already in use by another rule", StatusCode: http.StatusConflict}
		case btreeassociated.ErrorAssociatedIDAlreadyExists:
			logger.Warn("failed to create new rule because name exist", zap.Error(err))
			return &errors.ServerError{Message: "failed to create rule. Name is already in use", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to create or find a new rule", zap.Error(err))
			return &errors.ServerError{Message: "failed to create, unexpected error.", StatusCode: http.StatusInternalServerError}
		}
	}

	return nil
}

// list all group rules that match the provided key values. Can also include the overrides
func (rm *localRulesCient) MatchRules(ctx context.Context, matchQuery *v1limiter.RuleMatch) (v1limiter.Rules, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("MatchRules")
	ctx = reporting.UpdateLogger(ctx, logger)

	foundRules := v1limiter.Rules{}

	var rulesErr *errors.ServerError
	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		rule := item.Value().(Rule)

		var overrides v1limiter.Overrides
		if matchQuery.OverridesToMatch != nil {
			overrides, rulesErr = rm.overridesClient.MatchOverrides(ctx, item.AssociatedID(), matchQuery.OverridesToMatch)
		}

		if rulesErr != nil {
			// found an error matchin overrides. stop paginating through all the rules
			return false
		}

		foundRules = append(foundRules, &v1limiter.Rule{
			Name:      item.AssociatedID(),
			GroupBy:   item.KeyValues().Keys(),
			Limit:     rule.Limit(),
			Overrides: overrides,
		})

		return true
	}

	var selectError error
	switch matchQuery.RulesToMatch.KeyValues {
	case nil: // select all
		selectError = rm.rules.Query(datatypes.AssociatedKeyValuesQuery{}, bTreeAssociatedOnIterate)
	default: // select specific values
		// special match logic. we alwys need to look for empty strings as a 'match' operation
		keyValues := datatypes.KeyValues{}
		for key, _ := range *matchQuery.RulesToMatch.KeyValues {
			keyValues[key] = datatypes.String("")
		}

		selectError = rm.rules.MatchPermutations(keyValues, bTreeAssociatedOnIterate)
	}

	switch selectError {
	case nil:
		// nothing to do here
	default:
		logger.Error("failed to lookup rule.", zap.Error(selectError))
		return nil, errors.InternalServerError
	}

	if rulesErr != nil {
		logger.Error("failed to lookup overrides", zap.Error(rulesErr))
		return nil, rulesErr
	}

	return foundRules, nil
}

//	PARAMETERS:
//	- ctx - context that contains all the reporting tools
//	- name - name of the Rule to obtain for
//	- getQuery - query for the overrides to obtain as well
//
//	RETURNS:
//	- *errors.ServerError - error response for the client if something unexpected happens
//
// Get a Rule by name
func (rm *localRulesCient) GetRule(ctx context.Context, ruleName string, getQuery *v1limiter.RuleGet) (*v1limiter.Rule, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("GetRule").With(zap.String("rule_name", ruleName))
	ctx = reporting.UpdateLogger(ctx, logger)

	var apiRule *v1limiter.Rule
	errorMissingRuleName := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) {
		rule := item.Value().(Rule)

		if getQuery.OverridesToMatch != nil {
			// match overrides as well
			overrides, overridesErr := rm.overridesClient.MatchOverrides(ctx, ruleName, getQuery.OverridesToMatch)

			if overridesErr != nil {
				errorMissingRuleName = overridesErr
			} else {
				errorMissingRuleName = nil
				apiRule = &v1limiter.Rule{
					Name:      ruleName,
					GroupBy:   item.KeyValues().Keys(),
					Limit:     rule.Limit(),
					Overrides: overrides,
				}
			}
		} else {
			// don't query any overrides
			errorMissingRuleName = nil
			apiRule = &v1limiter.Rule{
				Name:      ruleName,
				GroupBy:   item.KeyValues().Keys(),
				Limit:     rule.Limit(),
				Overrides: v1limiter.Overrides{},
			}
		}
	}

	if err := rm.rules.FindByAssociatedID(ruleName, bTreeAssociatedOnFind); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to get rule. rule is currently being destroy")
			return nil, &errors.ServerError{Message: "failed to get rule. Previous rule is still in the process of destroying", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to lookup rule.", zap.Error(err))
			return nil, errors.InternalServerError
		}
	}

	if errorMissingRuleName != nil {
		logger.Warn("failed to get rule", zap.Error(errorMissingRuleName))
	}

	return apiRule, errorMissingRuleName
}

//	PARAMETERS:
//	- ctx - context that contains all the reporting tools
//	- ruleName - name of the rule that is going to be updated
//	- update - update reuest that defines the new values
//
//	RETURNS:
//	- *errors.ServerError - error response for the client if something unexpected happens
//
// Create new Rule
func (rm *localRulesCient) UpdateRule(ctx context.Context, ruleName string, update *v1limiter.RuleUpdateRquest) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("UpdateRule").With(zap.String("rule_name", ruleName))

	updateErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		rule := item.Value().(Rule)
		updateErr = rule.Update(logger, update)

		return false
	}

	idValue := datatypes.String(ruleName)
	query := datatypes.AssociatedKeyValuesQuery{
		KeyValueSelection: &datatypes.KeyValueSelection{
			KeyValues: map[string]datatypes.Value{
				btreeassociated.ReservedID: datatypes.Value{Value: &idValue, ValueComparison: datatypes.EqualsPtr()},
			},
		},
	}

	if err := rm.rules.Query(query, bTreeAssociatedOnIterate); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to update rule. Rule is currently being destroy")
			return &errors.ServerError{Message: "failed to update Rule. Rule is being destroying", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find rule by AssociatedID because of an internal server error", zap.Error(err))
			return errors.InternalServerError
		}
	}

	if updateErr != nil {
		logger.Warn("failed to update rule by id", zap.Error(updateErr))
	}

	return updateErr
}

// Delete a rule by name
func (rm *localRulesCient) DeleteRule(ctx context.Context, ruleName string) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("DeleteRule").With(zap.String("rule_name", ruleName))
	ctx = reporting.UpdateLogger(ctx, logger)

	ruleErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnCanDelete := func(item btreeassociated.AssociatedKeyValues) bool {
		// 1. First before deleting the rule, we need to delete all overrides for the rule
		if ruleErr = rm.overridesClient.DestroyOverrides(ctx, ruleName); ruleErr != nil {
			return false
		}

		// 2. Can now delete the rule
		rule := item.Value().(Rule)
		if ruleErr = rule.Delete(); ruleErr != nil {
			return false
		}

		return true
	}

	// NOTE: important to call destroy here. This will make all other calles to the same associated id fail fast
	if err := rm.rules.DestroyByAssociatedID(ruleName, bTreeAssociatedOnCanDelete); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to delete rule. Rule is already being destroy")
			return &errors.ServerError{Message: "failed to delete rule. Previous request to delete Rule is still in the process of destroying", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to lookup rule for deletion", zap.String("name", ruleName), zap.Error(err))
			return errors.InternalServerError
		}
	}

	if ruleErr != nil {
		logger.Error("failed to delete rule", zap.Error(ruleErr))
	}

	return ruleErr
}

// Create an override for a rule by name
func (rm *localRulesCient) CreateOverride(ctx context.Context, ruleName string, override *v1limiter.Override) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("CreateOverride").With(zap.String("rule_name", ruleName))
	ctx = reporting.UpdateLogger(ctx, logger)

	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) {
		// 1.  ensure that the override has all the group by keys
		for key, _ := range item.KeyValues() {
			if _, ok := override.KeyValues[key]; !ok {
				overrideErr = &errors.ServerError{Message: fmt.Sprintf("Missing Rule's GroubBy keys in the override: %s", key), StatusCode: http.StatusBadRequest}
				return
			}
		}

		// 2. create the override
		overrideErr = rm.overridesClient.CreateOverride(ctx, ruleName, override)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, bTreeAssociatedOnFind); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to create override. Rule is being destroy")
			return &errors.ServerError{Message: "Rule is being destroyed. Refusing to create Override", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find rule by associatedid", zap.String("name", ruleName), zap.Error(err))
			return errorMissingRuleName(ruleName)
		}
	}

	if overrideErr != nil {
		logger.Error("failed to crate Override", zap.Error(overrideErr))
	}

	return overrideErr
}

// match all overrides for a given Rule
func (rm *localRulesCient) MatchOverrides(ctx context.Context, ruleName string, matchQuery *v1common.MatchQuery) (v1limiter.Overrides, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("MatchOverrides").With(zap.String("rule_name", ruleName))
	ctx = reporting.UpdateLogger(ctx, logger)

	var overrides v1limiter.Overrides
	overridesErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) {
		overrides, overridesErr = rm.overridesClient.MatchOverrides(ctx, ruleName, matchQuery)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, bTreeAssociatedOnFind); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to match overrides. Rule is currently being destroy")
			return nil, &errors.ServerError{Message: "failed to match overrides. Rule is being destroying", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find rule by associatedid", zap.Error(err))
			return nil, errorMissingRuleName(ruleName)
		}
	}

	if overridesErr != nil {
		logger.Error("failed to match overrides", zap.Error(overridesErr))
		return nil, overridesErr
	}

	return overrides, nil
}

// get a single override by name
func (rm *localRulesCient) GetOverride(ctx context.Context, ruleName string, overrideName string) (*v1limiter.Override, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("GetOverride").With(zap.String("rule_name", ruleName))
	ctx = reporting.UpdateLogger(ctx, logger)

	var override *v1limiter.Override
	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) {
		override, overrideErr = rm.overridesClient.GetOverride(ctx, ruleName, overrideName)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, bTreeAssociatedOnFind); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to get rule. rule is currently being destroy")
			return nil, &errors.ServerError{Message: "Rule is being destroyed. Refusing to obtain Override", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find rule by associatedid", zap.Error(err))
			return nil, errorMissingRuleName(ruleName)
		}
	}

	if overrideErr != nil {
		logger.Error("failed to get override", zap.Error(overrideErr))
	}

	return override, overrideErr
}

// Update an override by its name
func (rm *localRulesCient) UpdateOverride(ctx context.Context, ruleName string, overrideName string, override *v1limiter.OverrideUpdate) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("UpdateOverried").With(zap.String("rule_name", ruleName))
	ctx = reporting.UpdateLogger(ctx, logger)

	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) {
		overrideErr = rm.overridesClient.UpdateOverride(ctx, ruleName, overrideName, override)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, bTreeAssociatedOnFind); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to update override. Rule is being destroyed")
			return &errors.ServerError{Message: "Rule is being destroyed. Refusing to update Override", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find rule by associatedid", zap.Error(err))
			return errorMissingRuleName(ruleName)
		}
	}

	if overrideErr != nil {
		logger.Error("failed to update Override", zap.Error(overrideErr))
	}

	return overrideErr
}

// Delete an override
func (rm *localRulesCient) DeleteOverride(ctx context.Context, ruleName string, overrideName string) *errors.ServerError {
	logger := reporting.GetLogger(ctx).Named("DeleteOverride").With(zap.String("rule_name", ruleName))
	ctx = reporting.UpdateLogger(ctx, logger)

	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) {
		overrideErr = rm.overridesClient.DestroyOverride(ctx, ruleName, overrideName)
	}

	if err := rm.rules.FindByAssociatedID(ruleName, bTreeAssociatedOnFind); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("refusing to destroy override. Rule is already being destroy")
			return &errors.ServerError{Message: "Rule is already being destroyed. Refusing to delete override again", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find rule by AssociatedID", zap.Error(err))
			return errorMissingRuleName(ruleName)
		}
	}

	if overrideErr != nil {
		logger.Error("failed to delete override", zap.Error(overrideErr))
	}

	return overrideErr
}

// Find the limits for each rule and the overrides for a give key values
func (rm localRulesCient) FindLimits(ctx context.Context, keyValue datatypes.KeyValues) (v1limiter.Rules, *errors.ServerError) {
	logger := reporting.GetLogger(ctx).Named("FindLimits")
	ctx = reporting.UpdateLogger(ctx, logger)

	rules := v1limiter.Rules{}
	var limitErr *errors.ServerError

	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		// 1. find all the overrides for the rule
		overrides, err := rm.overridesClient.FindOverrideLimits(ctx, item.AssociatedID(), keyValue)
		if err != nil {
			limitErr = err
			return false
		}

		// 2. append the rule
		newRule := &v1limiter.Rule{
			Name:      item.AssociatedID(),
			GroupBy:   item.KeyValues().Keys(),
			Limit:     item.Value().(Rule).Limit(),
			Overrides: overrides,
		}
		rules = append(rules, newRule)

		// 3. check if we can stop early
		if len(newRule.Overrides) == 0 {
			if newRule.Limit == 0 {
				return false
			}
		} else if newRule.Limit == 0 {
			if newRule.Overrides[len(newRule.Overrides)-1].Limit == 0 {
				return false
			}
		}

		return true
	}

	// special match logic. we always need to look for empty strings as a 'match' operation
	ruleKeyValues := datatypes.KeyValues{}
	for key, _ := range keyValue {
		ruleKeyValues[key] = datatypes.String("")
	}

	if err := rm.rules.MatchPermutations(ruleKeyValues, bTreeAssociatedOnIterate); err != nil {
		switch err {
		default:
			logger.Error("failed to match rule permutations", zap.Error(err))
			return nil, errors.InternalServerError
		}
	}

	if limitErr != nil {
		logger.Error("failed to find the limts", zap.Error(limitErr))
		return nil, limitErr
	}

	return rules, nil
}
