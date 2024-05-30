// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/pkg/clients/limiter_client (interfaces: LimiterClient)
//
// Generated by this command:
//
//	mockgen -destination=limiterclientfakes/limiter_client_mock.go -package=limiterclientfakes github.com/DanLavine/willow/pkg/clients/limiter_client LimiterClient
//

// Package limiterclientfakes is a generated GoMock package.
package limiterclientfakes

import (
	context "context"
	reflect "reflect"

	queryassociatedaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_associated_action"
	querymatchaction "github.com/DanLavine/willow/pkg/models/api/common/v1/query_match_action"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	gomock "go.uber.org/mock/gomock"
)

// MockLimiterClient is a mock of LimiterClient interface.
type MockLimiterClient struct {
	ctrl     *gomock.Controller
	recorder *MockLimiterClientMockRecorder
}

// MockLimiterClientMockRecorder is the mock recorder for MockLimiterClient.
type MockLimiterClientMockRecorder struct {
	mock *MockLimiterClient
}

// NewMockLimiterClient creates a new mock instance.
func NewMockLimiterClient(ctrl *gomock.Controller) *MockLimiterClient {
	mock := &MockLimiterClient{ctrl: ctrl}
	mock.recorder = &MockLimiterClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLimiterClient) EXPECT() *MockLimiterClientMockRecorder {
	return m.recorder
}

// CreateOverride mocks base method.
func (m *MockLimiterClient) CreateOverride(arg0 context.Context, arg1 string, arg2 *v1.Override) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOverride", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateOverride indicates an expected call of CreateOverride.
func (mr *MockLimiterClientMockRecorder) CreateOverride(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOverride", reflect.TypeOf((*MockLimiterClient)(nil).CreateOverride), arg0, arg1, arg2)
}

// CreateRule mocks base method.
func (m *MockLimiterClient) CreateRule(arg0 context.Context, arg1 *v1.Rule) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRule", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRule indicates an expected call of CreateRule.
func (mr *MockLimiterClientMockRecorder) CreateRule(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRule", reflect.TypeOf((*MockLimiterClient)(nil).CreateRule), arg0, arg1)
}

// DeleteOverride mocks base method.
func (m *MockLimiterClient) DeleteOverride(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteOverride", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteOverride indicates an expected call of DeleteOverride.
func (mr *MockLimiterClientMockRecorder) DeleteOverride(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOverride", reflect.TypeOf((*MockLimiterClient)(nil).DeleteOverride), arg0, arg1, arg2)
}

// DeleteRule mocks base method.
func (m *MockLimiterClient) DeleteRule(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRule", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRule indicates an expected call of DeleteRule.
func (mr *MockLimiterClientMockRecorder) DeleteRule(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRule", reflect.TypeOf((*MockLimiterClient)(nil).DeleteRule), arg0, arg1)
}

// GetOverride mocks base method.
func (m *MockLimiterClient) GetOverride(arg0 context.Context, arg1, arg2 string) (*v1.Override, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOverride", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.Override)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOverride indicates an expected call of GetOverride.
func (mr *MockLimiterClientMockRecorder) GetOverride(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOverride", reflect.TypeOf((*MockLimiterClient)(nil).GetOverride), arg0, arg1, arg2)
}

// GetRule mocks base method.
func (m *MockLimiterClient) GetRule(arg0 context.Context, arg1 string) (*v1.Rule, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRule", arg0, arg1)
	ret0, _ := ret[0].(*v1.Rule)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRule indicates an expected call of GetRule.
func (mr *MockLimiterClientMockRecorder) GetRule(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRule", reflect.TypeOf((*MockLimiterClient)(nil).GetRule), arg0, arg1)
}

// Healthy mocks base method.
func (m *MockLimiterClient) Healthy() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Healthy")
	ret0, _ := ret[0].(error)
	return ret0
}

// Healthy indicates an expected call of Healthy.
func (mr *MockLimiterClientMockRecorder) Healthy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Healthy", reflect.TypeOf((*MockLimiterClient)(nil).Healthy))
}

// MatchOverrides mocks base method.
func (m *MockLimiterClient) MatchOverrides(arg0 context.Context, arg1 string, arg2 *querymatchaction.MatchActionQuery) (v1.Overrides, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MatchOverrides", arg0, arg1, arg2)
	ret0, _ := ret[0].(v1.Overrides)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MatchOverrides indicates an expected call of MatchOverrides.
func (mr *MockLimiterClientMockRecorder) MatchOverrides(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MatchOverrides", reflect.TypeOf((*MockLimiterClient)(nil).MatchOverrides), arg0, arg1, arg2)
}

// MatchRules mocks base method.
func (m *MockLimiterClient) MatchRules(arg0 context.Context, arg1 *querymatchaction.MatchActionQuery) (v1.Rules, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MatchRules", arg0, arg1)
	ret0, _ := ret[0].(v1.Rules)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MatchRules indicates an expected call of MatchRules.
func (mr *MockLimiterClientMockRecorder) MatchRules(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MatchRules", reflect.TypeOf((*MockLimiterClient)(nil).MatchRules), arg0, arg1)
}

// QueryCounters mocks base method.
func (m *MockLimiterClient) QueryCounters(arg0 context.Context, arg1 *queryassociatedaction.AssociatedActionQuery) (v1.Counters, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryCounters", arg0, arg1)
	ret0, _ := ret[0].(v1.Counters)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryCounters indicates an expected call of QueryCounters.
func (mr *MockLimiterClientMockRecorder) QueryCounters(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryCounters", reflect.TypeOf((*MockLimiterClient)(nil).QueryCounters), arg0, arg1)
}

// QueryOverrides mocks base method.
func (m *MockLimiterClient) QueryOverrides(arg0 context.Context, arg1 string, arg2 *queryassociatedaction.AssociatedActionQuery) (v1.Overrides, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryOverrides", arg0, arg1, arg2)
	ret0, _ := ret[0].(v1.Overrides)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryOverrides indicates an expected call of QueryOverrides.
func (mr *MockLimiterClientMockRecorder) QueryOverrides(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryOverrides", reflect.TypeOf((*MockLimiterClient)(nil).QueryOverrides), arg0, arg1, arg2)
}

// QueryRules mocks base method.
func (m *MockLimiterClient) QueryRules(arg0 context.Context, arg1 *queryassociatedaction.AssociatedActionQuery) (v1.Rules, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryRules", arg0, arg1)
	ret0, _ := ret[0].(v1.Rules)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryRules indicates an expected call of QueryRules.
func (mr *MockLimiterClientMockRecorder) QueryRules(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRules", reflect.TypeOf((*MockLimiterClient)(nil).QueryRules), arg0, arg1)
}

// SetCounters mocks base method.
func (m *MockLimiterClient) SetCounters(arg0 context.Context, arg1 *v1.Counter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetCounters", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetCounters indicates an expected call of SetCounters.
func (mr *MockLimiterClientMockRecorder) SetCounters(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCounters", reflect.TypeOf((*MockLimiterClient)(nil).SetCounters), arg0, arg1)
}

// UpdateCounter mocks base method.
func (m *MockLimiterClient) UpdateCounter(arg0 context.Context, arg1 *v1.Counter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCounter", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCounter indicates an expected call of UpdateCounter.
func (mr *MockLimiterClientMockRecorder) UpdateCounter(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCounter", reflect.TypeOf((*MockLimiterClient)(nil).UpdateCounter), arg0, arg1)
}

// UpdateOverride mocks base method.
func (m *MockLimiterClient) UpdateOverride(arg0 context.Context, arg1, arg2 string, arg3 *v1.OverrideUpdate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOverride", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOverride indicates an expected call of UpdateOverride.
func (mr *MockLimiterClientMockRecorder) UpdateOverride(arg0, arg1, arg2, arg3 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOverride", reflect.TypeOf((*MockLimiterClient)(nil).UpdateOverride), arg0, arg1, arg2, arg3)
}

// UpdateRule mocks base method.
func (m *MockLimiterClient) UpdateRule(arg0 context.Context, arg1 string, arg2 *v1.RuleUpdateRquest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateRule", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateRule indicates an expected call of UpdateRule.
func (mr *MockLimiterClientMockRecorder) UpdateRule(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRule", reflect.TypeOf((*MockLimiterClient)(nil).UpdateRule), arg0, arg1, arg2)
}
