package v1server

import (
	"fmt"
	"net/http"

	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/limiter"
	"github.com/DanLavine/willow/internal/logger"
	"github.com/DanLavine/willow/pkg/models/api"
	v1limiter "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"go.uber.org/zap"
)

// Handles CRUD operations for Limit operations
//
//go:generate mockgen -destination=v1serverfakes/limiter_rules_mock.go -package=v1serverfakes github.com/DanLavine/willow/internal/server/versions/v1server LimitRuleHandler
type LimitRuleHandler interface {
	// CRUD operations
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)

	// overide operations
	SetOverride(w http.ResponseWriter, r *http.Request)
	DeleteOverride(w http.ResponseWriter, r *http.Request)

	// counter operations
	Increment(w http.ResponseWriter, r *http.Request)
	Decrement(w http.ResponseWriter, r *http.Request)
}

type groupRuleHandler struct {
	logger *zap.Logger

	rulesManager limiter.RulesManager
}

func NewGroupRuleHandler(logger *zap.Logger, rulesManager limiter.RulesManager) *groupRuleHandler {
	return &groupRuleHandler{
		logger:       logger.Named("GroupRuleHandler"),
		rulesManager: rulesManager,
	}
}

// Create a new rule handler
func (grh *groupRuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Create"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	limiterCreateRequest, err := v1limiter.ParseRuleRequest(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())

		return
	}

	// create the group rule. On find, return an error that the rule already exists
	if err = grh.rulesManager.Create(logger, limiterCreateRequest); err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Get a rule by name hendler
func (grh *groupRuleHandler) Get(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Get"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())

	query, err := v1limiter.ParseRuleQuery(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	// find the group rule
	rule := grh.rulesManager.Get(logger, namedParameters["rule_name"], query)
	if rule == nil {
		err := &api.Error{Message: fmt.Sprintf("rule with name '%s' could not be found", namedParameters["rule_name"])}
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rule.ToBytes())
}

// Update a rule by name handelr
func (grh *groupRuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Update"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	limiterUpdateRequest, err := v1limiter.ParseRuleUpdateRequest(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	namedParameters := urlrouter.GetNamedParamters(r.Context())

	// find the specific limiter group rule
	err = grh.rulesManager.Update(logger, namedParameters["rule_name"], limiterUpdateRequest)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	// successfully updated the group rule
	w.WriteHeader(http.StatusOK)
}

// Delete a rule by name handler
func (grh *groupRuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Delete"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())

	if err := grh.rulesManager.Delete(logger, namedParameters["rule_name"]); err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (grh *groupRuleHandler) List(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("List"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	query, err := v1limiter.ParseRuleQuery(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	rules, err := grh.rulesManager.List(logger, query)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rules.ToBytes())
}

func (grh *groupRuleHandler) SetOverride(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("SetOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	ruleOverrideRequest, err := v1limiter.ParseOverrideRequest(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	namedParameters := urlrouter.GetNamedParamters(r.Context())

	if err := grh.rulesManager.CreateOverride(logger, namedParameters["rule_name"], ruleOverrideRequest); err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (grh *groupRuleHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("DeleteOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	namedParameters := urlrouter.GetNamedParamters(r.Context())

	if err := grh.rulesManager.DeleteOverride(logger, namedParameters["rule_name"], namedParameters["override_name"]); err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (grh *groupRuleHandler) Increment(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Increment"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	counterRequest, err := v1limiter.ParseCounterRequest(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())

		return
	}

	// create the group rule. On find, return an error that the rule already exists
	if err = grh.rulesManager.IncrementKeyValues(logger, counterRequest); err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (grh *groupRuleHandler) Decrement(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Increment"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	w.WriteHeader(http.StatusNotImplemented)
}
