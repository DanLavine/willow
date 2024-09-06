package rules

import (
	"context"
	"fmt"
	"net/http"

	btreeassociated "github.com/DanLavine/willow/internal/datastructures/btree_associated"
	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/limiter/overrides"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
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
func (rm *localRulesCient) CreateRule(ctx context.Context, rule *v1limiter.Rule) (string, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "CreateRule")

	onCreate := func() any {
		return rm.ruleConstructor.New(rule.Spec.Properties)
	}

	// create the rule only if the name is free
	id, err := rm.rules.Create(rule.Spec.DBDefinition.GroupByKeyValues, onCreate)
	if err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to create rule. rule is currently being destroy")
			return "", &errors.ServerError{Message: "failed to create rule. Previous rule is still in the process of destroying", StatusCode: http.StatusConflict}
		case btreeassociated.ErrorKeyValuesAlreadyExists:
			logger.Warn("failed to create new rule because keys exist", zap.Error(err))
			return "", &errors.ServerError{Message: "failed to create rule. GroupBy keys already in use by another rule", StatusCode: http.StatusConflict}
		case btreeassociated.ErrorAssociatedIDAlreadyExists:
			logger.Warn("failed to create new rule because name exist", zap.Error(err))
			return "", &errors.ServerError{Message: "failed to create rule. Name is already in use", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to create or find a new rule", zap.Error(err))
			return "", &errors.ServerError{Message: "failed to create, unexpected error.", StatusCode: http.StatusInternalServerError}
		}
	}

	return id, nil
}

// list all group rules that the query matches
func (rm *localRulesCient) QueryRules(ctx context.Context, ruleQuery *queryassociatedaction.AssociatedActionQuery) (v1limiter.Rules, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "QueryRules")
	foundRules := v1limiter.Rules{}

	bTreeAssociatedOnIterate := func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		rule := associatedKeyValues.Value().(Rule)

		foundRules = append(foundRules, &v1limiter.Rule{
			Spec: &v1limiter.RuleSpec{
				DBDefinition: &v1limiter.RuleDBDefinition{
					GroupByKeyValues: associatedKeyValues.KeyValues(),
				},
				Properties: &v1limiter.RuleProperties{
					Limit: helpers.PointerOf(rule.Limit()),
				},
			},
			State: &v1limiter.RuleState{
				ID:        associatedKeyValues.AssociatedID(),
				Overrides: v1limiter.Overrides{},
			},
		})

		return true
	}

	if err := rm.rules.QueryAction(ruleQuery, bTreeAssociatedOnIterate); err != nil {
		logger.Error("failed to query rules", zap.Error(err))
		return nil, errors.InternalServerError
	}

	return foundRules, nil
}

// list all group rules that match the provided key values
func (rm *localRulesCient) MatchRules(ctx context.Context, ruleMatch *querymatchaction.MatchActionQuery) (v1limiter.Rules, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "MatchRules")
	foundRules := v1limiter.Rules{}

	bTreeAssociatedOnIterate := func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		rule := associatedKeyValues.Value().(Rule)

		foundRules = append(foundRules, &v1limiter.Rule{
			Spec: &v1limiter.RuleSpec{
				DBDefinition: &v1limiter.RuleDBDefinition{
					GroupByKeyValues: associatedKeyValues.KeyValues(),
				},
				Properties: &v1limiter.RuleProperties{
					Limit: helpers.PointerOf(rule.Limit()),
				},
			},
			State: &v1limiter.RuleState{
				ID:        associatedKeyValues.AssociatedID(),
				Overrides: v1limiter.Overrides{},
			},
		})

		return true
	}

	if err := rm.rules.MatchAction(ruleMatch, bTreeAssociatedOnIterate); err != nil {
		logger.Error("failed to query rules", zap.Error(err))
		return nil, errors.InternalServerError
	}

	return foundRules, nil
}

//	PARAMETERS:
//	- ctx - context that contains all the reporting tools
//	- name - name of the Rule to obtain for
//
//	RETURNS:
//	- *errors.ServerError - error response for the client if something unexpected happens
//
// Get a Rule by name
func (rm *localRulesCient) GetRule(ctx context.Context, ruleName string) (*v1limiter.Rule, *errors.ServerError) {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "GetRule")

	var apiRule *v1limiter.Rule
	errorMissingRuleName := errorMissingRuleName(ruleName)

	bTreeAssociatedOnIterate := func(associatedKeyValues btreeassociated.AssociatedKeyValues) bool {
		rule := associatedKeyValues.Value().(Rule)
		errorMissingRuleName = nil

		apiRule = &v1limiter.Rule{
			Spec: &v1limiter.RuleSpec{
				DBDefinition: &v1limiter.RuleDBDefinition{
					GroupByKeyValues: associatedKeyValues.KeyValues(),
				},
				Properties: &v1limiter.RuleProperties{
					Limit: helpers.PointerOf(rule.Limit()),
				},
			},
			State: &v1limiter.RuleState{
				ID:        associatedKeyValues.AssociatedID(),
				Overrides: v1limiter.Overrides{},
			},
		}

		return true
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnIterate); err != nil {
		logger.Error("failed to query rules", zap.Error(err))
		return nil, errors.InternalServerError
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
func (rm *localRulesCient) UpdateRule(ctx context.Context, ruleName string, update *v1limiter.RuleProperties) *errors.ServerError {
	_, logger := middleware.GetNamedMiddlewareLogger(ctx, "UpdateRule")

	updateErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		rule := item.Value().(Rule)
		updateErr = rule.Update(update)

		return false
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnIterate); err != nil {
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
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "DeleteRule")

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
func (rm *localRulesCient) CreateOverride(ctx context.Context, ruleName string, override *v1limiter.Override) (string, *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "CreateOverride")

	var overrideID string
	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) bool {
		// 1.  ensure that the override has all the group by keys
		for key, _ := range item.KeyValues() {
			if _, ok := override.Spec.DBDefinition.GroupByKeyValues[key]; !ok {
				overrideErr = &errors.ServerError{Message: fmt.Sprintf("Missing Rule's GroubBy keys in the override: %s", key), StatusCode: http.StatusBadRequest}
				return false
			}
		}

		// 2. create the override
		overrideID, overrideErr = rm.overridesClient.CreateOverride(ctx, ruleName, override)

		return false
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnFind); err != nil {
		switch err {
		case btreeassociated.ErrorTreeItemDestroying:
			logger.Warn("failed to create override. Rule is being destroy")
			return "", &errors.ServerError{Message: "Rule is being destroyed. Refusing to create Override", StatusCode: http.StatusConflict}
		default:
			logger.Error("failed to find rule by associatedid", zap.String("name", ruleName), zap.Error(err))
			return "", errorMissingRuleName(ruleName)
		}
	}

	if overrideErr != nil {
		logger.Error("failed to crate Override", zap.Error(overrideErr))
	}

	return overrideID, overrideErr
}

// query all overrides for a given Rule
func (rm *localRulesCient) QueryOverrides(ctx context.Context, ruleName string, overrideQuery *queryassociatedaction.AssociatedActionQuery) (v1limiter.Overrides, *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "QueryOverrides")

	var overrides v1limiter.Overrides
	overridesErr := errorMissingRuleName(ruleName)

	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) bool {
		overrides, overridesErr = rm.overridesClient.QueryOverrides(ctx, ruleName, overrideQuery)
		return overridesErr == nil
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnFind); err != nil {
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

// match all overrides for a given Rule
func (rm *localRulesCient) MatchOverrides(ctx context.Context, ruleName string, overrideMatch *querymatchaction.MatchActionQuery) (v1limiter.Overrides, *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "MAtchOverrides")

	var overrides v1limiter.Overrides
	overridesErr := errorMissingRuleName(ruleName)

	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) bool {
		overrides, overridesErr = rm.overridesClient.MatchOverrides(ctx, ruleName, overrideMatch)
		return overridesErr == nil
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnFind); err != nil {
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
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "GetOverride")

	var override *v1limiter.Override
	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) bool {
		override, overrideErr = rm.overridesClient.GetOverride(ctx, ruleName, overrideName)

		if overrideErr != nil {
			return false
		}

		return true
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnFind); err != nil {
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
func (rm *localRulesCient) UpdateOverride(ctx context.Context, ruleName string, overrideName string, override *v1limiter.OverrideProperties) *errors.ServerError {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "UpdateOverride")

	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) bool {
		overrideErr = rm.overridesClient.UpdateOverride(ctx, ruleName, overrideName, override)
		if overrideErr != nil {
			return false
		}

		return true
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnFind); err != nil {
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
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "DeleteOverride")

	overrideErr := errorMissingRuleName(ruleName)
	bTreeAssociatedOnFind := func(item btreeassociated.AssociatedKeyValues) bool {
		overrideErr = rm.overridesClient.DestroyOverride(ctx, ruleName, overrideName)
		return false
	}

	if err := rm.rules.QueryAction(queryassociatedaction.StringToAssociatedActionQuery(ruleName), bTreeAssociatedOnFind); err != nil {
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
func (rm localRulesCient) FindLimits(ctx context.Context, keyValues datatypes.KeyValues) (v1limiter.Rules, *errors.ServerError) {
	ctx, logger := middleware.GetNamedMiddlewareLogger(ctx, "FindLimits")

	rules := v1limiter.Rules{}
	var limitErr *errors.ServerError

	bTreeAssociatedOnIterate := func(item btreeassociated.AssociatedKeyValues) bool {
		// 1. find all the overrides for the rule
		overrides, err := rm.overridesClient.FindOverrideLimits(ctx, item.AssociatedID(), keyValues)
		if err != nil {
			limitErr = err
			return false
		}

		// 2. append the rule
		newRule := &v1limiter.Rule{
			Spec: &v1limiter.RuleSpec{
				DBDefinition: &v1limiter.RuleDBDefinition{
					GroupByKeyValues: item.KeyValues(),
				},
				Properties: &v1limiter.RuleProperties{
					Limit: helpers.PointerOf(item.Value().(Rule).Limit()),
				},
			},
			State: &v1limiter.RuleState{
				ID:        item.AssociatedID(),
				Overrides: overrides,
			},
		}
		rules = append(rules, newRule)

		// 3. check if we can stop early
		if len(newRule.State.Overrides) == 0 {
			if *newRule.Spec.Properties.Limit == 0 {
				return false
			}
		} else if *newRule.Spec.Properties.Limit == 0 {
			if *newRule.State.Overrides[len(newRule.State.Overrides)-1].Spec.Properties.Limit == 0 {
				return false
			}
		}

		return true
	}

	if err := rm.rules.MatchAction(querymatchaction.KeyValuesToAnyMatchActionQuery(keyValues), bTreeAssociatedOnIterate); err != nil {
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
