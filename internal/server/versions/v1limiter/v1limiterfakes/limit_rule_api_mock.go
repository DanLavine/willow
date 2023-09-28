// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/internal/server/versions/v1limiter (interfaces: LimitRuleHandler)

// Package v1limiterfakes is a generated GoMock package.
package v1limiterfakes

import (
	http "net/http"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockLimitRuleHandler is a mock of LimitRuleHandler interface.
type MockLimitRuleHandler struct {
	ctrl     *gomock.Controller
	recorder *MockLimitRuleHandlerMockRecorder
}

// MockLimitRuleHandlerMockRecorder is the mock recorder for MockLimitRuleHandler.
type MockLimitRuleHandlerMockRecorder struct {
	mock *MockLimitRuleHandler
}

// NewMockLimitRuleHandler creates a new mock instance.
func NewMockLimitRuleHandler(ctrl *gomock.Controller) *MockLimitRuleHandler {
	mock := &MockLimitRuleHandler{ctrl: ctrl}
	mock.recorder = &MockLimitRuleHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLimitRuleHandler) EXPECT() *MockLimitRuleHandlerMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockLimitRuleHandler) Create(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Create", arg0, arg1)
}

// Create indicates an expected call of Create.
func (mr *MockLimitRuleHandlerMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockLimitRuleHandler)(nil).Create), arg0, arg1)
}

// Decrement mocks base method.
func (m *MockLimitRuleHandler) Decrement(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Decrement", arg0, arg1)
}

// Decrement indicates an expected call of Decrement.
func (mr *MockLimitRuleHandlerMockRecorder) Decrement(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decrement", reflect.TypeOf((*MockLimitRuleHandler)(nil).Decrement), arg0, arg1)
}

// Delete mocks base method.
func (m *MockLimitRuleHandler) Delete(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Delete", arg0, arg1)
}

// Delete indicates an expected call of Delete.
func (mr *MockLimitRuleHandlerMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockLimitRuleHandler)(nil).Delete), arg0, arg1)
}

// Find mocks base method.
func (m *MockLimitRuleHandler) Find(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Find", arg0, arg1)
}

// Find indicates an expected call of Find.
func (mr *MockLimitRuleHandlerMockRecorder) Find(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockLimitRuleHandler)(nil).Find), arg0, arg1)
}

// Increment mocks base method.
func (m *MockLimitRuleHandler) Increment(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Increment", arg0, arg1)
}

// Increment indicates an expected call of Increment.
func (mr *MockLimitRuleHandlerMockRecorder) Increment(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Increment", reflect.TypeOf((*MockLimitRuleHandler)(nil).Increment), arg0, arg1)
}

// List mocks base method.
func (m *MockLimitRuleHandler) List(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "List", arg0, arg1)
}

// List indicates an expected call of List.
func (mr *MockLimitRuleHandlerMockRecorder) List(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockLimitRuleHandler)(nil).List), arg0, arg1)
}

// SetOverride mocks base method.
func (m *MockLimitRuleHandler) SetOverride(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetOverride", arg0, arg1)
}

// SetOverride indicates an expected call of SetOverride.
func (mr *MockLimitRuleHandlerMockRecorder) SetOverride(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetOverride", reflect.TypeOf((*MockLimitRuleHandler)(nil).SetOverride), arg0, arg1)
}

// Update mocks base method.
func (m *MockLimitRuleHandler) Update(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Update", arg0, arg1)
}

// Update indicates an expected call of Update.
func (mr *MockLimitRuleHandlerMockRecorder) Update(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockLimitRuleHandler)(nil).Update), arg0, arg1)
}