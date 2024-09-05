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

// Create an Override for a specific Rule
func (grh *groupRuleHandler) CreateOverride(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "CreateOverride")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the override from the request
	override := &v1limiter.Override{}
	if err := api.ObjectDecodeRequest(r, override); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// create the override
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	overrideID, err := grh.ruleClient.CreateOverride(ctx, namedParameters["rule_id"], override)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	override.State = &v1limiter.OverrideState{
		ID: overrideID,
	}
	_, _ = api.ModelEncodeResponse(w, http.StatusCreated, override)
}

// query a number of Overrides for a rule
func (grh *groupRuleHandler) QueryOverrides(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "QueryOverrides")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// read the query to run
	query := &queryassociatedaction.AssociatedActionQuery{}
	if err := api.ModelDecodeRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	overrides, err := grh.ruleClient.QueryOverrides(ctx, namedParameters["rule_id"], query)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, &overrides)
}

// match a number of Overrides for a rule
func (grh *groupRuleHandler) MatchOverrides(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "MatchOVerrides")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// read the query to run
	match := &querymatchaction.MatchActionQuery{}
	if err := api.ModelDecodeRequest(r, match); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	overrides, err := grh.ruleClient.MatchOverrides(ctx, namedParameters["rule_id"], match)
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, &overrides)
}

// Get a particular override
func (grh *groupRuleHandler) GetOverride(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "GetOVerride")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	override, err := grh.ruleClient.GetOverride(ctx, namedParameters["rule_id"], namedParameters["override_id"])
	if err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	if override != nil {
		// found an override
		_, _ = api.ModelEncodeResponse(w, http.StatusOK, override)
	} else {
		// override not found
		_, _ = api.ModelEncodeResponse(w, http.StatusNotFound, nil)
	}
}

// Update a particular override
func (grh *groupRuleHandler) UpdateOverride(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "UpdateOverride")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the update parameters
	overrideUpdate := &v1limiter.OverrideProperties{}
	if err := api.ModelDecodeRequest(r, overrideUpdate); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.UpdateOverride(ctx, namedParameters["rule_id"], namedParameters["override_id"], overrideUpdate); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusOK, nil)
}

// Delete an override for a specific rule
func (grh *groupRuleHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	// grab the request middleware objects
	ctx, logger := middleware.GetNamedMiddlewareLogger(r.Context(), "DeleteOverride")
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.DeleteOverride(ctx, namedParameters["rule_id"], namedParameters["override_id"]); err != nil {
		_, _ = api.ModelEncodeResponse(w, err.StatusCode, err)
		return
	}

	_, _ = api.ModelEncodeResponse(w, http.StatusNoContent, nil)
}
