package handlers

import (
	"context"
	"net/http"

	"github.com/DanLavine/willow/internal/limiter/counters"
	"github.com/DanLavine/willow/internal/limiter/rules"
	"github.com/DanLavine/willow/pkg/clients"
)

// Handles CRUD operations for Limit operations
//
//go:generate mockgen -destination=v1serverfakes/limiter_rules_mock.go -package=v1serverfakes github.com/DanLavine/willow/internal/server/versions/v1server LimitRuleHandler
type V1LimiterRuleHandler interface {
	// RULES
	CreateRule(w http.ResponseWriter, r *http.Request)
	QueryRules(w http.ResponseWriter, r *http.Request)
	MatchRules(w http.ResponseWriter, r *http.Request)
	GetRule(w http.ResponseWriter, r *http.Request)
	UpdateRule(w http.ResponseWriter, r *http.Request)
	DeleteRule(w http.ResponseWriter, r *http.Request)

	// OVERRIDES
	CreateOverride(w http.ResponseWriter, r *http.Request)
	QueryOverrides(w http.ResponseWriter, r *http.Request)
	MatchOverrides(w http.ResponseWriter, r *http.Request)
	GetOverride(w http.ResponseWriter, r *http.Request)
	UpdateOverride(w http.ResponseWriter, r *http.Request)
	DeleteOverride(w http.ResponseWriter, r *http.Request)

	// COUNTERS
	UpsertCounters(w http.ResponseWriter, r *http.Request)
	QueryCounters(w http.ResponseWriter, r *http.Request)
	SetCounters(w http.ResponseWriter, r *http.Request)
}

type groupRuleHandler struct {
	shutdownContext context.Context

	// locker client to ensure that all locks are respected
	lockerClientConfig *clients.Config

	// clients for the service
	ruleClient    rules.RuleClient
	counterClient counters.CounterClient
}

func NewGroupRuleHandler(shutdownContext context.Context, lockerClientConfig *clients.Config, rulesClient rules.RuleClient, countersClient counters.CounterClient) *groupRuleHandler {
	return &groupRuleHandler{
		shutdownContext:    shutdownContext,
		lockerClientConfig: lockerClientConfig,
		ruleClient:         rulesClient,
		counterClient:      countersClient,
	}
}
