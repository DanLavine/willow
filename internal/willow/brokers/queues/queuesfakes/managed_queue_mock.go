// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/internal/willow/brokers/queues (interfaces: ManagedQueue)

// Package queuesfakes is a generated GoMock package.
package queuesfakes

import (
	context "context"
	reflect "reflect"

	errors "github.com/DanLavine/willow/pkg/models/api/common/errors"
	v1 "github.com/DanLavine/willow/pkg/models/api/willow/v1"
	datatypes "github.com/DanLavine/willow/pkg/models/datatypes"
	gomock "go.uber.org/mock/gomock"
	zap "go.uber.org/zap"
)

// MockManagedQueue is a mock of ManagedQueue interface.
type MockManagedQueue struct {
	ctrl     *gomock.Controller
	recorder *MockManagedQueueMockRecorder
}

// MockManagedQueueMockRecorder is the mock recorder for MockManagedQueue.
type MockManagedQueueMockRecorder struct {
	mock *MockManagedQueue
}

// NewMockManagedQueue creates a new mock instance.
func NewMockManagedQueue(ctrl *gomock.Controller) *MockManagedQueue {
	mock := &MockManagedQueue{ctrl: ctrl}
	mock.recorder = &MockManagedQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockManagedQueue) EXPECT() *MockManagedQueueMockRecorder {
	return m.recorder
}

// ACK mocks base method.
func (m *MockManagedQueue) ACK(arg0 *zap.Logger, arg1 *v1.ACK) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ACK", arg0, arg1)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// ACK indicates an expected call of ACK.
func (mr *MockManagedQueueMockRecorder) ACK(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ACK", reflect.TypeOf((*MockManagedQueue)(nil).ACK), arg0, arg1)
}

// Dequeue mocks base method.
func (m *MockManagedQueue) Dequeue(arg0 *zap.Logger, arg1 context.Context, arg2 datatypes.AssociatedKeyValuesQuery) (*v1.DequeueItemResponse, func(), func(), *errors.ServerError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dequeue", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.DequeueItemResponse)
	ret1, _ := ret[1].(func())
	ret2, _ := ret[2].(func())
	ret3, _ := ret[3].(*errors.ServerError)
	return ret0, ret1, ret2, ret3
}

// Dequeue indicates an expected call of Dequeue.
func (mr *MockManagedQueueMockRecorder) Dequeue(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dequeue", reflect.TypeOf((*MockManagedQueue)(nil).Dequeue), arg0, arg1, arg2)
}

// Enqueue mocks base method.
func (m *MockManagedQueue) Enqueue(arg0 *zap.Logger, arg1 *v1.EnqueueItemRequest) *errors.ServerError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enqueue", arg0, arg1)
	ret0, _ := ret[0].(*errors.ServerError)
	return ret0
}

// Enqueue indicates an expected call of Enqueue.
func (mr *MockManagedQueueMockRecorder) Enqueue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enqueue", reflect.TypeOf((*MockManagedQueue)(nil).Enqueue), arg0, arg1)
}

// Execute mocks base method.
func (m *MockManagedQueue) Execute(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute.
func (mr *MockManagedQueueMockRecorder) Execute(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockManagedQueue)(nil).Execute), arg0)
}

// Metrics mocks base method.
func (m *MockManagedQueue) Metrics() *v1.QueueMetricsResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Metrics")
	ret0, _ := ret[0].(*v1.QueueMetricsResponse)
	return ret0
}

// Metrics indicates an expected call of Metrics.
func (mr *MockManagedQueueMockRecorder) Metrics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Metrics", reflect.TypeOf((*MockManagedQueue)(nil).Metrics))
}
