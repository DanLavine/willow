package handlers

import (
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/models/api"
	"go.uber.org/zap"

	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// Create a new rule handler
func (grh *groupRuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("CreateRule"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the rule create request
	ruleRequest := &v1limiter.RuleCreateRequest{}
	if err := api.DecodeAndValidateHttpRequest(r, ruleRequest); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// create the group rule if it does not already exist
	if err := grh.ruleClient.CreateRule(logger, ruleRequest); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Match rules handler
func (grh *groupRuleHandler) MatchRules(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("MatchRules"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the matcher
	query := &v1limiter.RuleMatch{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// find the Rule and desired overrides
	rules, err := grh.ruleClient.MatchRules(logger, query)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, &rules)
}

// Get a rule by name
func (grh *groupRuleHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Get"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the rule and overides to match for
	ruleGet := &v1limiter.RuleGet{}
	if err := api.DecodeAndValidateHttpRequest(r, ruleGet); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	rule, err := grh.ruleClient.GetRule(logger, namedParameters["rule_name"], ruleGet)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// if the error is nil, there will always be a rule
	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, rule)
}

// Update a rule by name
func (grh *groupRuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Update"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the update request
	ruleUpdate := &v1limiter.RuleUpdateRquest{}
	if err := api.DecodeAndValidateHttpRequest(r, ruleUpdate); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// Update the specific rule off of the name
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.UpdateRule(logger, namedParameters["rule_name"], ruleUpdate); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// successfully updated the group rule
	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
}

// Delete a Rule by name
func (grh *groupRuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Delete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.ruleClient.DeleteRule(logger, namedParameters["rule_name"]); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusNoContent, nil)
}
