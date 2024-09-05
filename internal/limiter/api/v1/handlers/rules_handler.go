package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/middleware"
	"github.com/DanLavine/willow/pkg/models/api"
	"go.uber.org/zap"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// Create a new rule handler
func (grh *groupRuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "CreateRule")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the rule create request
	rule := &v1limiter.Rule{}
	if err := api.ObjectDecodeRequest(r, rule); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// create the group rule if it does not already exist
	ruleID, err := grh.ruleClient.CreateRule(ctx, rule)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	rule.State = &v1limiter.RuleState{
		ID: ruleID,
	}
	_, _ = api.ModelEncodeResponse(w, http.StatusCreated, rule)
}

// Query rules handler
func (grh *groupRuleHandler) QueryRules(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "QueryRules")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the matcher
	query := &queryassociatedaction.AssociatedActionQuery{}
	if err := api.ModelDecodeRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// find the Rule and desired overrides
	rules, err := grh.ruleClient.QueryRules(ctx, query)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, &rules)
}

// Match rules handler
func (grh *groupRuleHandler) MatchRules(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "MatchRules")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the matcher
	match := &querymatchaction.MatchActionQuery{}
	if err := api.ModelDecodeRequest(r, match); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// find the Rule and desired overrides
	rules, err := grh.ruleClient.MatchRules(ctx, match)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, &rules)
}

// Get a rule by name
func (grh *groupRuleHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "GetRule")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	rule, err := grh.ruleClient.GetRule(ctx, namedParameters["rule_id"])
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// if the error is nil, there will always be a rule
	_, _ = api.ModelEncodeResponse(w, http.StatusOK, rule)
}

// Update a rule by name
func (grh *groupRuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "UpdateRule")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the update request
	ruleUpdate := &v1limiter.RuleProperties{}
	if err := api.ModelDecodeRequest(r, ruleUpdate); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// Update the specific rule off of the name
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.UpdateRule(ctx, namedParameters["rule_id"], ruleUpdate); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// successfully updated the group rule
	_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
}

// Delete a Rule by name
func (grh *groupRuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "DeleteRule")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.DeleteRule(ctx, namedParameters["rule_id"]); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusNoContent, nil)
}
