// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/internal/willow/brokers/queue_channels/constructor (interfaces: QueueChannel)

// Package constructorfakes is a generated GoMock package.
package constructorfakes

import (
	context "context"
	reflect "reflect"

	errors "github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/willow/v1"
	gomock "go.uber.org/mock/gomock"
)

// MockQueueChannel is a mock of QueueChannel interface.
type MockQueueChannel struct {
	ctrl     *gomock.Controller
	recorder *MockQueueChannelMockRecorder
}

// MockQueueChannelMockRecorder is the mock recorder for MockQueueChannel.
type MockQueueChannelMockRecorder struct {
	mock *MockQueueChannel
}

// NewMockQueueChannel creates a new mock instance.
func NewMockQueueChannel(ctrl *gomock.Controller) *MockQueueChannel {
	mock := &MockQueueChannel{ctrl: ctrl}
	mock.recorder = &MockQueueChannelMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueueChannel) EXPECT() *MockQueueChannelMockRecorder {
	return m.recorder
}

// ACK mocks base method.
func (m *MockQueueChannel) ACK(arg0 context.Context, arg1 *v1.ACK) (bool, *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ACK", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(*errors.ServerError)
	return ret0, ret1
}

// ACK indicates an expected call of ACK.
func (mr *MockQueueChannelMockRecorder) ACK(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ACK", reflect.TypeOf((*MockQueueChannel)(nil).ACK), arg0, arg1)
}

// Delete mocks base method.
func (m *MockQueueChannel) Delete() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockQueueChannelMockRecorder) Delete() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockQueueChannel)(nil).Delete))
}

// Dequeue mocks base method.
func (m *MockQueueChannel) Dequeue() <-chan func(context.Context) (*v1.DequeueQueueItem, func(), func()) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dequeue")
	ret0, _ := ret[0].(<-chan func(context.Context) (*v1.DequeueQueueItem, func(), func()))
	return ret0
}

// Dequeue indicates an expected call of Dequeue.
func (mr *MockQueueChannelMockRecorder) Dequeue() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dequeue", reflect.TypeOf((*MockQueueChannel)(nil).Dequeue))
}

// Enqueue mocks base method.
func (m *MockQueueChannel) Enqueue(arg0 context.Context, arg1 *v1.EnqueueQueueItem) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enqueue", arg0, arg1)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// Enqueue indicates an expected call of Enqueue.
func (mr *MockQueueChannelMockRecorder) Enqueue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enqueue", reflect.TypeOf((*MockQueueChannel)(nil).Enqueue), arg0, arg1)
}

// Execute mocks base method.
func (m *MockQueueChannel) Execute(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute.
func (mr *MockQueueChannelMockRecorder) Execute(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockQueueChannel)(nil).Execute), arg0)
}

// ForceDelete mocks base method.
func (m *MockQueueChannel) ForceDelete(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ForceDelete", arg0)
}

// ForceDelete indicates an expected call of ForceDelete.
func (mr *MockQueueChannelMockRecorder) ForceDelete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceDelete", reflect.TypeOf((*MockQueueChannel)(nil).ForceDelete), arg0)
}

// Heartbeat mocks base method.
func (m *MockQueueChannel) Heartbeat(arg0 context.Context, arg1 *v1.Heartbeat) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Heartbeat", arg0, arg1)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// Heartbeat indicates an expected call of Heartbeat.
func (mr *MockQueueChannelMockRecorder) Heartbeat(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Heartbeat", reflect.TypeOf((*MockQueueChannel)(nil).Heartbeat), arg0, arg1)
}
