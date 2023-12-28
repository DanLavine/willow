package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/limiter"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/clients"
	lockerclient "github.com/DanLavine/willow/pkg/clients/locker_client"
	"github.com/DanLavine/willow/pkg/models/api"
	"github.com/DanLavine/willow/pkg/models/datatypes"
	"go.uber.org/zap"

	"github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
)

// Handles CRUD operations for Limit operations
//
//go:generate mockgen -destination=v1serverfakes/limiter_rules_mock.go -package=v1serverfakes github.com/DanLavine/willow/internal/server/versions/v1server LimitRuleHandler
type V1LimiterRuleHandler interface {
	// CRUD operations
	CreateRule(w http.ResponseWriter, r *http.Request)
	UpdateRule(w http.ResponseWriter, r *http.Request)
	DeleteRule(w http.ResponseWriter, r *http.Request)
	GetRule(w http.ResponseWriter, r *http.Request) // DSL TODO: not consistent with List Overrides. I think I want 2 apis. 1 for match. 1 for querry?
	ListRules(w http.ResponseWriter, r *http.Request)

	// overide operations
	ListOverrides(w http.ResponseWriter, r *http.Request)
	SetOverride(w http.ResponseWriter, r *http.Request)
	DeleteOverride(w http.ResponseWriter, r *http.Request)

	// counter operations
	UpdateCounters(w http.ResponseWriter, r *http.Request)
	ListCounters(w http.ResponseWriter, r *http.Request)
	SetCounters(w http.ResponseWriter, r *http.Request)
}

type groupRuleHandler struct {
	logger *zap.Logger

	shutdownContext context.Context

	// locker client to ensure that all locks are respected
	lockerClientConfig *clients.Config

	rulesManager limiter.RulesManager
}

func NewGroupRuleHandler(logger *zap.Logger, shutdownContext context.Context, lockerClientConfig *clients.Config, rulesManager limiter.RulesManager) *groupRuleHandler {
	return &groupRuleHandler{
		logger:             logger.Named("GroupRuleHandler"),
		shutdownContext:    shutdownContext,
		lockerClientConfig: lockerClientConfig,
		rulesManager:       rulesManager,
	}
}

// Create a new rule handler
func (grh *groupRuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Create"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	ruleRequest := &v1limiter.RuleCreateRequest{}
	if err := api.DecodeAndValidateHttpRequest(r, ruleRequest); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// create the group rule if it does not already exist
	if err := grh.rulesManager.Create(logger, ruleRequest); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Get a rule by name
func (grh *groupRuleHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Get"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the query
	query := &v1limiter.RuleQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// find the Rule and dessired overrides
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	rule := grh.rulesManager.Get(logger, namedParameters["rule_name"], query)
	if rule == nil {
		err := &errors.ServerError{Message: fmt.Sprintf("rule with name '%s' could not be found", namedParameters["rule_name"]), StatusCode: http.StatusUnprocessableEntity}
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, rule)
}

// List all rules that match particual KeyValues
func (grh *groupRuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("List"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	query := &v1limiter.RuleQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	rules, err := grh.rulesManager.List(logger, query)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, &rules)
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
	if err := grh.rulesManager.Update(logger, namedParameters["rule_name"], ruleUpdate); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// successfully updated the group rule
	w.WriteHeader(http.StatusOK)
}

// Delete a Rule by name
func (grh *groupRuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Delete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.rulesManager.Delete(logger, namedParameters["rule_name"]); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List a number of Overrides for a rule
func (grh *groupRuleHandler) ListOverrides(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("ListOverrides"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// read the query to run
	query := &v1common.AssociatedQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	overrides, err := grh.rulesManager.ListOverrides(logger, namedParameters["rule_name"], query)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, &overrides)
}

// Set an override for a specific rule
func (grh *groupRuleHandler) SetOverride(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("SetOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the override from the request
	override := &v1limiter.Override{}
	if err := api.DecodeAndValidateHttpRequest(r, override); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// create the override
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.rulesManager.CreateOverride(logger, namedParameters["rule_name"], override); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusCreated, nil)
}

// Delete an override for a specific rule
func (grh *groupRuleHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("DeleteOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if err := grh.rulesManager.DeleteOverride(logger, namedParameters["rule_name"], namedParameters["override_name"]); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusNoContent, nil)
}

// Query the counters to see what is already provided
func (grh *groupRuleHandler) ListCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("ListCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the query from the counters
	query := &v1common.AssociatedQuery{}
	if err := api.DecodeAndValidateHttpRequest(r, query); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	countersResp, err := grh.rulesManager.ListCounters(logger, query)
	if err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, &countersResp)
}

// Increment the Counters if they do not conflict with any rules
func (grh *groupRuleHandler) UpdateCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("UpdateCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the counter increment
	counters := &v1limiter.Counter{}
	if err := api.DecodeAndValidateHttpRequest(r, counters); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	if counters.Counters > 0 {
		// this is an increment request
		// create a locker client that will stop and close if a server shutdown is received
		logLockErr := func(kvs datatypes.KeyValues, err error) {
			logger.Error("failed to obtain lock", zap.Error(err), zap.Any("key_values", kvs))
		}
		lockerClient, lockerErr := lockerclient.NewLockClient(grh.shutdownContext, grh.lockerClientConfig, logLockErr)
		if lockerErr != nil {
			logger.Error("failed to create locker client on increment counter request", zap.Error(lockerErr))
			err := errors.InternalServerError
			_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
			return
		}

		// attempt to increment the counters for a particualr group of KeyValues
		if err := grh.rulesManager.IncrementCounters(logger, r.Context(), lockerClient, counters); err != nil {
			_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
			return
		}

		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
	} else {
		// this is a decrement request
		if err := grh.rulesManager.DecrementCounters(logger, counters); err != nil {
			_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
			return
		}

		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
	}
}

func (grh *groupRuleHandler) SetCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("SetCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the countrs set
	counter := &v1limiter.Counter{}
	if err := api.DecodeAndValidateHttpRequest(r, counter); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	// forcefully set the counters
	if err := grh.rulesManager.SetCounters(logger, counter); err != nil {
		_, _ = api.EncodeAndSendHttpResponse(r.Header, w, err.StatusCode, err)
		return
	}

	_, _ = api.EncodeAndSendHttpResponse(r.Header, w, http.StatusOK, nil)
}
