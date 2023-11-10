package v1server

import (
	"net/http"

	"github.com/DanLavine/willow/internal/limiter"
	"github.com/DanLavine/willow/internal/logger"
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
	Find(w http.ResponseWriter, r *http.Request)

	// overide operations
	SetOverride(w http.ResponseWriter, r *http.Request)
	DeleteOverride(w http.ResponseWriter, r *http.Request)

	// Increment a group of tags
	Increment(w http.ResponseWriter, r *http.Request)

	// Decrement a group of tags
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

func (grh *groupRuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Create"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	limiterCreateReqquest, err := v1limiter.ParseRuleRequest(r.Body)
	if err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())

		return
	}

	// create the group rule. On find, return an error that the rule already exists
	if err = grh.rulesManager.CreateGroupRule(logger, limiterCreateReqquest); err != nil {
		w.WriteHeader(err.StatusCode)
		w.Write(err.ToBytes())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (grh *groupRuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Update"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	w.WriteHeader(http.StatusNotImplemented)

	//limiterUpdateRequest, err := v1limiter.ParseRuleUpdateRequest(r.Body)
	//if err != nil {
	//	w.WriteHeader(err.StatusCode)
	//	w.Write(err.ToBytes())
	//	return
	//}
	//
	//namedParameters := urlrouter.GetNamedParamters(r.Context())
	//
	//// find the specific limiter group rule
	//limiterRule := grh.rulesManager.FindRule(logger, namedParameters["name"])
	//if limiterRule == nil {
	//	err = &api.Error{Message: "rule not found", StatusCode: http.StatusUnprocessableEntity}
	//	w.WriteHeader(err.StatusCode)
	//	w.Write(err.ToBytes())
	//	return
	//}
	//
	//// create the group rule. On find, return an error that the rule already exists
	//limiterRule.Update(logger, limiterUpdateRequest.Limit)
	//w.WriteHeader(http.StatusOK)
}

func (grh *groupRuleHandler) Find(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Find"), r)
	logger.Debug("starting CreateGroupRule request")
	w.WriteHeader(http.StatusNotImplemented)
	//defer logger.Debug("processed CreateGroupRule request")
	//
	//ruleName := r.URL.Query().Get("name")
	//if ruleName == "" {
	//	err := api.InvalidRequestBody.With("Name to be provided", "recieved empty string")
	//	w.WriteHeader(err.StatusCode)
	//	w.Write(err.ToBytes())
	//	return
	//}
	//
	//// find the specific limiter group rule
	//limiterRule := grh.rulesManager.FindRule(logger, ruleName)
	//if limiterRule == nil {
	//	err := &api.Error{Message: "rule not found", StatusCode: http.StatusBadRequest}
	//	w.WriteHeader(err.StatusCode)
	//	w.Write(err.ToBytes())
	//	return
	//}
	//
	//data, jsonErr := json.Marshal(limiterRule)
	//if jsonErr != nil {
	//	logger.Error("Failed to JSON marshel response", zap.Error(jsonErr))
	//	w.WriteHeader(http.StatusInternalServerError)
	//	w.Write(errors.InternalServerError.With("", jsonErr.Error()).ToBytes())
	//	return
	//}
	//
	//if data == nil {
	//	w.WriteHeader(http.StatusNoContent)
	//	return
	//}
	//
	//w.WriteHeader(http.StatusOK)
	//w.Write(data)
}

func (grh *groupRuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Delete"), r)
	logger.Debug("starting CreateGroupRule request")
	defer logger.Debug("processed CreateGroupRule request")

	w.WriteHeader(http.StatusNotImplemented)

	// switch method := r.Method; method {
	// case "DELETE":
	// 	ruleName := r.URL.Query().Get("name")
	// 	if ruleName == "" {
	// 		err := api.InvalidRequestBody.With("Name to be provided", "recieved empty string")
	// 		w.WriteHeader(err.StatusCode)
	// 		w.Write(err.ToBytes())
	// 		return
	// 	}

	// 	grh.rulesManager.DeleteGroupRule(logger, ruleName)
	// 	w.WriteHeader(http.StatusNoContent)
	// default:
	// 	w.WriteHeader(http.StatusNotFound)
	// }
}

func (grh *groupRuleHandler) SetOverride(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("SetOverride"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	w.WriteHeader(http.StatusNotImplemented)

	//ruleOverrideRequest, err := v1limiter.ParseOverrideRequest(r.Body)
	//if err != nil {
	//	w.WriteHeader(err.StatusCode)
	//	w.Write(err.ToBytes())
	//	return
	//}
	//
	//namedParameters := urlrouter.GetNamedParamters(r.Context())
	//
	//rule := grh.rulesManager.FindRule(logger, namedParameters["_associated_id"])
	//if rule == nil {
	//	err = &api.Error{Message: "rule not found", StatusCode: http.StatusBadRequest}
	//	w.WriteHeader(err.StatusCode)
	//	w.Write(err.ToBytes())
	//	return
	//}
	//
	//rule.Update(logger, ruleOverrideRequest.Limit)
	//w.WriteHeader(http.StatusCreated)
}

func (grh *groupRuleHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (grh *groupRuleHandler) Increment(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Increment"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	w.WriteHeader(http.StatusNotImplemented)
}

func (grh *groupRuleHandler) Decrement(w http.ResponseWriter, r *http.Request) {
	logger := logger.AddRequestID(grh.logger.Named("Increment"), r)
	logger.Debug("starting request")
	defer logger.Debug("processed request")

	w.WriteHeader(http.StatusNotImplemented)
}
