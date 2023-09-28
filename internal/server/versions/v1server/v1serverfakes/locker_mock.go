// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/internal/server/versions/v1server (interfaces: LockerHandler)

// Package v1serverfakes is a generated GoMock package.
package v1serverfakes

import (
	http "net/http"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockLockerHandler is a mock of LockerHandler interface.
type MockLockerHandler struct {
	ctrl     *gomock.Controller
	recorder *MockLockerHandlerMockRecorder
}

// MockLockerHandlerMockRecorder is the mock recorder for MockLockerHandler.
type MockLockerHandlerMockRecorder struct {
	mock *MockLockerHandler
}

// NewMockLockerHandler creates a new mock instance.
func NewMockLockerHandler(ctrl *gomock.Controller) *MockLockerHandler {
	mock := &MockLockerHandler{ctrl: ctrl}
	mock.recorder = &MockLockerHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLockerHandler) EXPECT() *MockLockerHandlerMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockLockerHandler) Create(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Create", arg0, arg1)
}

// Create indicates an expected call of Create.
func (mr *MockLockerHandlerMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockLockerHandler)(nil).Create), arg0, arg1)
}

// Delete mocks base method.
func (m *MockLockerHandler) Delete(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Delete", arg0, arg1)
}

// Delete indicates an expected call of Delete.
func (mr *MockLockerHandlerMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockLockerHandler)(nil).Delete), arg0, arg1)
}

// List mocks base method.
func (m *MockLockerHandler) List(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "List", arg0, arg1)
}

// List indicates an expected call of List.
func (mr *MockLockerHandlerMockRecorder) List(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockLockerHandler)(nil).List), arg0, arg1)
}
