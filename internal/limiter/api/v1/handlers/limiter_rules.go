package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/limiter"
	"github.com/DanLavine/willow/internal/logger"
	servererrors "github.com/DanLavine/willow/internal/server_errors"
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
	ListCounters(w http.ResponseWriter, r *http.Request)
	Increment(w http.ResponseWriter, r *http.Request)
	Decrement(w http.ResponseWriter, r *http.Request)
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

	ruleRequest := &v1limiter.RuleRequest{}
	if err := ruleRequest.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// create the group rule if it does not already exist
	if serverErr := grh.rulesManager.Create(logger, ruleRequest); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
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
	if err := query.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// find the Rule and dessired overrides
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	rule := grh.rulesManager.Get(logger, namedParameters["rule_name"], query)
	if rule == nil {
		serverErr := &servererrors.ApiError{Message: fmt.Sprintf("rule with name '%s' could not be found", namedParameters["rule_name"])}
		_, _ = api.HttpResponse(r, w, http.StatusUnprocessableEntity, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusOK, rule)
}

// List all rules that match particual KeyValues
func (grh *groupRuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("List"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	query := &v1limiter.RuleQuery{}
	if err := query.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	rules, serverErr := grh.rulesManager.List(logger, query)
	if serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusOK, rules)
}

// Update a rule by name
func (grh *groupRuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Update"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the update request
	ruleUpdate := &v1limiter.RuleUpdate{}
	if err := ruleUpdate.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// Update the specific rule off of the name
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if serverErr := grh.rulesManager.Update(logger, namedParameters["rule_name"], ruleUpdate); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
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
	if serverErr := grh.rulesManager.Delete(logger, namedParameters["rule_name"]); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
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
	if err := query.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// find all the overrides for the particular rule
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	overrides, serverErr := grh.rulesManager.ListOverrides(logger, namedParameters["rule_name"], query)
	if serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusOK, overrides)
}

// Set an override for a specific rule
func (grh *groupRuleHandler) SetOverride(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("SetOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the override from the request
	override := &v1limiter.Override{}
	if err := override.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// create the override
	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if serverErr := grh.rulesManager.CreateOverride(logger, namedParameters["rule_name"], override); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusCreated, nil)
}

// Delete an override for a specific rule
func (grh *groupRuleHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("DeleteOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())
	if serverErr := grh.rulesManager.DeleteOverride(logger, namedParameters["rule_name"], namedParameters["override_name"]); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusNoContent, nil)
}

// Query the counters to see what is already provided
func (grh *groupRuleHandler) ListCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("ListCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the query from the counters
	query := &v1common.AssociatedQuery{}
	if err := query.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	countersResp, serverErr := grh.rulesManager.ListCounters(logger, query)
	if serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusOK, countersResp)
}

// Increment the Counters if they do not conflict with any rules
func (grh *groupRuleHandler) Increment(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Increment"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the counter increment
	counters := &v1limiter.Counter{}
	if err := counters.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// create a locker client that will stop and close if a server shutdown is received
	logLockErr := func(kvs datatypes.KeyValues, err error) {
		logger.Error("failed to obtain lock", zap.Error(err), zap.Any("key_values", kvs))
	}
	lockerClient, lockerErr := lockerclient.NewLockerClient(grh.shutdownContext, grh.lockerClientConfig, logLockErr)
	if lockerErr != nil {
		logger.Error("failed to create locker client on increment counter request", zap.Error(lockerErr))
		serverErr := servererrors.InternalServerError
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	// attempt to increment the counters for a particualr group of KeyValues
	if serverErr := grh.rulesManager.IncrementCounters(logger, r.Context(), lockerClient, counters); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusOK, nil)
}

func (grh *groupRuleHandler) Decrement(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Decrement"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the counter decerement
	counters := &v1limiter.Counter{}
	if err := counters.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// decrement the counter
	if serverErr := grh.rulesManager.DecrementCounters(logger, counters); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusNoContent, nil)
}

func (grh *groupRuleHandler) SetCounters(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("SetCounters"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	// parse the countrs set
	setCounter := &v1limiter.CounterSet{}
	if err := setCounter.Decode(api.ContentTypeFromRequest(r), r.Body); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		_, _ = api.HttpResponse(r, w, http.StatusBadRequest, errors.ServerError(err))
		return
	}

	// forcefully set the counters
	if serverErr := grh.rulesManager.SetCounters(logger, setCounter); serverErr != nil {
		_, _ = api.HttpResponse(r, w, serverErr.StatusCode, serverErr)
		return
	}

	_, _ = api.HttpResponse(r, w, http.StatusOK, nil)
}
