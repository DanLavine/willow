// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/internal/limiter/rules (interfaces: RuleConstructor)
//
// Generated by this command:
//
//	mockgen -destination=rulefakes/constructor_mock.go -package=rulefakes github.com/DanLavine/willow/internal/limiter/rules RuleConstructor
//

// Package rulefakes is a generated GoMock package.
package rulefakes

import (
	reflect "reflect"

	rules "github.com/DanLavine/willow/internal/limiter/rules"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	gomock "go.uber.org/mock/gomock"
)

// MockRuleConstructor is a mock of RuleConstructor interface.
type MockRuleConstructor struct {
	ctrl     *gomock.Controller
	recorder *MockRuleConstructorMockRecorder
}

// MockRuleConstructorMockRecorder is the mock recorder for MockRuleConstructor.
type MockRuleConstructorMockRecorder struct {
	mock *MockRuleConstructor
}

// NewMockRuleConstructor creates a new mock instance.
func NewMockRuleConstructor(ctrl *gomock.Controller) *MockRuleConstructor {
	mock := &MockRuleConstructor{ctrl: ctrl}
	mock.recorder = &MockRuleConstructorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRuleConstructor) EXPECT() *MockRuleConstructorMockRecorder {
	return m.recorder
}

// New mocks base method.
func (m *MockRuleConstructor) New(arg0 *v1.RuleProperties) rules.Rule {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", arg0)
	ret0, _ := ret[0].(rules.Rule)
	return ret0
}

// New indicates an expected call of New.
func (mr *MockRuleConstructorMockRecorder) New(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockRuleConstructor)(nil).New), arg0)
}
