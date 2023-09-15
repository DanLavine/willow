package v1limiter

import "net/http"

type GroupRuleHandler interface {
	// group rule operations
	CreateGroupRule(w http.ResponseWriter, r *http.Request)
	FindGroupRule(w http.ResponseWriter, r *http.Request)
	UpdateGroupRule(w http.ResponseWriter, r *http.Request)
	DeleteGroupRule(w http.ResponseWriter, r *http.Request)
}

type groupRuleHandler struct {
}

func NewGroupRuleHandler() *groupRuleHandler {
	return &groupRuleHandler{}
}

func (grh *groupRuleHandler) CreateGroupRule(w http.ResponseWriter, r *http.Request) {

}
