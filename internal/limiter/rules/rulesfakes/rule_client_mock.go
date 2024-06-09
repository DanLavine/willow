// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/internal/limiter/rules (interfaces: RuleClient)
//
// Generated by this command:
//
//	mockgen -destination=rulesfakes/rule_client_mock.go -package=rulesfakes github.com/DanLavine/willow/internal/limiter/rules RuleClient
//

// Package rulesfakes is a generated GoMock package.
package rulesfakes

import (
	context "context"
	reflect "reflect"

	errors "github.com/DanLavine/willow/pkg/models/api/common/errors"
	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	datatypes "github.com/DanLavine/willow/pkg/models/datatypes"
	gomock "go.uber.org/mock/gomock"
)

// MockRuleClient is a mock of RuleClient interface.
type MockRuleClient struct {
	ctrl     *gomock.Controller
	recorder *MockRuleClientMockRecorder
}

// MockRuleClientMockRecorder is the mock recorder for MockRuleClient.
type MockRuleClientMockRecorder struct {
	mock *MockRuleClient
}

// NewMockRuleClient creates a new mock instance.
func NewMockRuleClient(ctrl *gomock.Controller) *MockRuleClient {
	mock := &MockRuleClient{ctrl: ctrl}
	mock.recorder = &MockRuleClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRuleClient) EXPECT() *MockRuleClientMockRecorder {
	return m.recorder
}

// CreateOverride mocks base method.
func (m *MockRuleClient) CreateOverride(arg0 context.Context, arg1 string, arg2 *v1.Override) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOverride", arg0, arg1, arg2)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// CreateOverride indicates an expected call of CreateOverride.
func (mr *MockRuleClientMockRecorder) CreateOverride(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOverride", reflect.TypeOf((*MockRuleClient)(nil).CreateOverride), arg0, arg1, arg2)
}

// CreateRule mocks base method.
func (m *MockRuleClient) CreateRule(arg0 context.Context, arg1 *v1.Rule) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRule", arg0, arg1)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// CreateRule indicates an expected call of CreateRule.
func (mr *MockRuleClientMockRecorder) CreateRule(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRule", reflect.TypeOf((*MockRuleClient)(nil).CreateRule), arg0, arg1)
}

// DeleteOverride mocks base method.
func (m *MockRuleClient) DeleteOverride(arg0 context.Context, arg1, arg2 string) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteOverride", arg0, arg1, arg2)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// DeleteOverride indicates an expected call of DeleteOverride.
func (mr *MockRuleClientMockRecorder) DeleteOverride(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOverride", reflect.TypeOf((*MockRuleClient)(nil).DeleteOverride), arg0, arg1, arg2)
}

// DeleteRule mocks base method.
func (m *MockRuleClient) DeleteRule(arg0 context.Context, arg1 string) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRule", arg0, arg1)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// DeleteRule indicates an expected call of DeleteRule.
func (mr *MockRuleClientMockRecorder) DeleteRule(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRule", reflect.TypeOf((*MockRuleClient)(nil).DeleteRule), arg0, arg1)
}

// FindLimits mocks base method.
func (m *MockRuleClient) FindLimits(arg0 context.Context, arg1 datatypes.KeyValues) (v1.Rules, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindLimits", arg0, arg1)
	ret0, _ := ret[0].(v1.Rules)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// FindLimits indicates an expected call of FindLimits.
func (mr *MockRuleClientMockRecorder) FindLimits(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindLimits", reflect.TypeOf((*MockRuleClient)(nil).FindLimits), arg0, arg1)
}

// GetOverride mocks base method.
func (m *MockRuleClient) GetOverride(arg0 context.Context, arg1, arg2 string) (*v1.Override, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOverride", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Override)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// GetOverride indicates an expected call of GetOverride.
func (mr *MockRuleClientMockRecorder) GetOverride(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOverride", reflect.TypeOf((*MockRuleClient)(nil).GetOverride), arg0, arg1, arg2)
}

// GetRule mocks base method.
func (m *MockRuleClient) GetRule(arg0 context.Context, arg1 string) (*v1.Rule, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRule", arg0, arg1)
	ret0, _ := ret[0].(*v1.Rule)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// GetRule indicates an expected call of GetRule.
func (mr *MockRuleClientMockRecorder) GetRule(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRule", reflect.TypeOf((*MockRuleClient)(nil).GetRule), arg0, arg1)
}

// MatchOverrides mocks base method.
func (m *MockRuleClient) MatchOverrides(arg0 context.Context, arg1 string, arg2 *querymatchaction.MatchActionQuery) (v1.Overrides, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MatchOverrides", arg0, arg1, arg2)
	ret0, _ := ret[0].(v1.Overrides)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// MatchOverrides indicates an expected call of MatchOverrides.
func (mr *MockRuleClientMockRecorder) MatchOverrides(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MatchOverrides", reflect.TypeOf((*MockRuleClient)(nil).MatchOverrides), arg0, arg1, arg2)
}

// MatchRules mocks base method.
func (m *MockRuleClient) MatchRules(arg0 context.Context, arg1 *querymatchaction.MatchActionQuery) (v1.Rules, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MatchRules", arg0, arg1)
	ret0, _ := ret[0].(v1.Rules)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// MatchRules indicates an expected call of MatchRules.
func (mr *MockRuleClientMockRecorder) MatchRules(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MatchRules", reflect.TypeOf((*MockRuleClient)(nil).MatchRules), arg0, arg1)
}

// QueryOverrides mocks base method.
func (m *MockRuleClient) QueryOverrides(arg0 context.Context, arg1 string, arg2 *queryassociatedaction.AssociatedActionQuery) (v1.Overrides, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryOverrides", arg0, arg1, arg2)
	ret0, _ := ret[0].(v1.Overrides)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// QueryOverrides indicates an expected call of QueryOverrides.
func (mr *MockRuleClientMockRecorder) QueryOverrides(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryOverrides", reflect.TypeOf((*MockRuleClient)(nil).QueryOverrides), arg0, arg1, arg2)
}

// QueryRules mocks base method.
func (m *MockRuleClient) QueryRules(arg0 context.Context, arg1 *queryassociatedaction.AssociatedActionQuery) (v1.Rules, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryRules", arg0, arg1)
	ret0, _ := ret[0].(v1.Rules)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// QueryRules indicates an expected call of QueryRules.
func (mr *MockRuleClientMockRecorder) QueryRules(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRules", reflect.TypeOf((*MockRuleClient)(nil).QueryRules), arg0, arg1)
}

// UpdateOverride mocks base method.
func (m *MockRuleClient) UpdateOverride(arg0 context.Context, arg1, arg2 string, arg3 *v1.OverrideProperties) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOverride", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// UpdateOverride indicates an expected call of UpdateOverride.
func (mr *MockRuleClientMockRecorder) UpdateOverride(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOverride", reflect.TypeOf((*MockRuleClient)(nil).UpdateOverride), arg0, arg1, arg2, arg3)
}

// UpdateRule mocks base method.
func (m *MockRuleClient) UpdateRule(arg0 context.Context, arg1 string, arg2 *v1.RuleProperties) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateRule", arg0, arg1, arg2)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// UpdateRule indicates an expected call of UpdateRule.
func (mr *MockRuleClientMockRecorder) UpdateRule(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRule", reflect.TypeOf((*MockRuleClient)(nil).UpdateRule), arg0, arg1, arg2)
}
