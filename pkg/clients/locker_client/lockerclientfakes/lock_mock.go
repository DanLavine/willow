// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/pkg/clients/locker_client (interfaces: Locker)

// Package lockerclientfakes is a generated GoMock package.
package lockerclientfakes

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockLocker is a mock of Locker interface.
type MockLocker struct {
	ctrl     *gomock.Controller
	recorder *MockLockerMockRecorder
}

// MockLockerMockRecorder is the mock recorder for MockLocker.
type MockLockerMockRecorder struct {
	mock *MockLocker
}

// NewMockLocker creates a new mock instance.
func NewMockLocker(ctrl *gomock.Controller) *MockLocker {
	mock := &MockLocker{ctrl: ctrl}
	mock.recorder = &MockLockerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLocker) EXPECT() *MockLockerMockRecorder {
	return m.recorder
}

// Done mocks base method.
func (m *MockLocker) Done() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Done indicates an expected call of Done.
func (mr *MockLockerMockRecorder) Done() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Done", reflect.TypeOf((*MockLocker)(nil).Done))
}

// Release mocks base method.
func (m *MockLocker) Release() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Release")
	ret0, _ := ret[0].(error)
	return ret0
}

// Release indicates an expected call of Release.
func (mr *MockLockerMockRecorder) Release() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Release", reflect.TypeOf((*MockLocker)(nil).Release))
}
