// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DanLavine/willow/internal/datastructures/btree_one_to_many (interfaces: BTreeOneToMany)

// Package btreeonetomanyfakes is a generated GoMock package.
package btreeonetomanyfakes

import (
	reflect "reflect"

	btreeonetomany "github.com/DanLavine/willow/internal/datastructures/btree_one_to_many"
	datatypes "github.com/DanLavine/willow/pkg/models/datatypes"
	gomock "go.uber.org/mock/gomock"
)

// MockBTreeOneToMany is a mock of BTreeOneToMany interface.
type MockBTreeOneToMany struct {
	ctrl     *gomock.Controller
	recorder *MockBTreeOneToManyMockRecorder
}

// MockBTreeOneToManyMockRecorder is the mock recorder for MockBTreeOneToMany.
type MockBTreeOneToManyMockRecorder struct {
	mock *MockBTreeOneToMany
}

// NewMockBTreeOneToMany creates a new mock instance.
func NewMockBTreeOneToMany(ctrl *gomock.Controller) *MockBTreeOneToMany {
	mock := &MockBTreeOneToMany{ctrl: ctrl}
	mock.recorder = &MockBTreeOneToManyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBTreeOneToMany) EXPECT() *MockBTreeOneToManyMockRecorder {
	return m.recorder
}

// CreateOrFind mocks base method.
func (m *MockBTreeOneToMany) CreateOrFind(arg0 string, arg1 datatypes.KeyValues, arg2 btreeonetomany.OneToManyTreeOnCreate, arg3 btreeonetomany.OneToManyTreeOnFind) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrFind", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrFind indicates an expected call of CreateOrFind.
func (mr *MockBTreeOneToManyMockRecorder) CreateOrFind(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrFind", reflect.TypeOf((*MockBTreeOneToMany)(nil).CreateOrFind), arg0, arg1, arg2, arg3)
}

// CreateWithID mocks base method.
func (m *MockBTreeOneToMany) CreateWithID(arg0, arg1 string, arg2 datatypes.KeyValues, arg3 btreeonetomany.OneToManyTreeOnCreate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWithID", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateWithID indicates an expected call of CreateWithID.
func (mr *MockBTreeOneToManyMockRecorder) CreateWithID(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWithID", reflect.TypeOf((*MockBTreeOneToMany)(nil).CreateWithID), arg0, arg1, arg2, arg3)
}

// DeleteOneOfManyByKeyValues mocks base method.
func (m *MockBTreeOneToMany) DeleteOneOfManyByKeyValues(arg0 string, arg1 datatypes.KeyValues, arg2 btreeonetomany.OneToManyTreeRemove) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteOneOfManyByKeyValues", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteOneOfManyByKeyValues indicates an expected call of DeleteOneOfManyByKeyValues.
func (mr *MockBTreeOneToManyMockRecorder) DeleteOneOfManyByKeyValues(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOneOfManyByKeyValues", reflect.TypeOf((*MockBTreeOneToMany)(nil).DeleteOneOfManyByKeyValues), arg0, arg1, arg2)
}

// DestroyOne mocks base method.
func (m *MockBTreeOneToMany) DestroyOne(arg0 string, arg1 btreeonetomany.OneToManyTreeRemove) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DestroyOne", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DestroyOne indicates an expected call of DestroyOne.
func (mr *MockBTreeOneToManyMockRecorder) DestroyOne(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DestroyOne", reflect.TypeOf((*MockBTreeOneToMany)(nil).DestroyOne), arg0, arg1)
}

// DestroyOneOfManyByID mocks base method.
func (m *MockBTreeOneToMany) DestroyOneOfManyByID(arg0, arg1 string, arg2 btreeonetomany.OneToManyTreeRemove) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DestroyOneOfManyByID", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DestroyOneOfManyByID indicates an expected call of DestroyOneOfManyByID.
func (mr *MockBTreeOneToManyMockRecorder) DestroyOneOfManyByID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DestroyOneOfManyByID", reflect.TypeOf((*MockBTreeOneToMany)(nil).DestroyOneOfManyByID), arg0, arg1, arg2)
}

// MatchPermutations mocks base method.
func (m *MockBTreeOneToMany) MatchPermutations(arg0 string, arg1 datatypes.KeyValues, arg2 btreeonetomany.OneToManyTreeIterate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MatchPermutations", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// MatchPermutations indicates an expected call of MatchPermutations.
func (mr *MockBTreeOneToManyMockRecorder) MatchPermutations(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MatchPermutations", reflect.TypeOf((*MockBTreeOneToMany)(nil).MatchPermutations), arg0, arg1, arg2)
}

// Query mocks base method.
func (m *MockBTreeOneToMany) Query(arg0 string, arg1 datatypes.AssociatedKeyValuesQuery, arg2 btreeonetomany.OneToManyTreeIterate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Query indicates an expected call of Query.
func (mr *MockBTreeOneToManyMockRecorder) Query(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockBTreeOneToMany)(nil).Query), arg0, arg1, arg2)
}
