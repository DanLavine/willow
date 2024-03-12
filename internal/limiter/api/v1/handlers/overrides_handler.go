package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/pkg/models/api"
	"go.uber.org/zap"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// Create an Override for a specific Rule
func (grh *groupRuleHandler) CreateOverride(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), grh.logger.Named("CreateOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the override from the request
	override := &v1limiter.Override{}
	if err := api.DecodeAndValidateHttpRequest(r, override); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// create the override
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.CreateOverride(ctx, namedParameters["rule_name"], override); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusCreated, nil)
}

// Get a particular override
func (grh *groupRuleHandler) GetOverride(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), grh.logger.Named("GetOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	override, err := grh.ruleClient.GetOverride(ctx, namedParameters["rule_name"], namedParameters["override_name"])
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if override != nil {
		// found an override
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, override)
	} else {
		// override not found
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusNotFound, nil)
	}
}

// Match a number of Overrides for a rule
func (grh *groupRuleHandler) MatchOverrides(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), grh.logger.Named("MatchOverrides"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// read the query to run
	query := &v1common.MatchQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	overrides, err := grh.ruleClient.MatchOverrides(ctx, namedParameters["rule_name"], query)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, &overrides)
}

// Update a particular override
func (grh *groupRuleHandler) UpdateOverride(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), grh.logger.Named("UpdateOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the update parameters
	overrideUpdate := &v1limiter.OverrideUpdate{}
	if err := api.DecodeAndValidateHttpRequest(r, overrideUpdate); err != nil {
		logger.Warn("failed to decode and validate request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.UpdateOverride(ctx, namedParameters["rule_name"], namedParameters["override_name"], overrideUpdate); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
}

// Delete an override for a specific rule
func (grh *groupRuleHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	ctx, logger := reporting.SetupContextWithLoggerFromRequest(r.Context(), grh.logger.Named("DeleteOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.DeleteOverride(ctx, namedParameters["rule_name"], namedParameters["override_name"]); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusNoContent, nil)
}
