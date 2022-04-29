// Code generated by MockGen. DO NOT EDIT.
// Source: db.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockTx is a mock of Tx interface.
type MockTx struct {
	ctrl     *gomock.Controller
	recorder *MockTxMockRecorder
}

// MockTxMockRecorder is the mock recorder for MockTx.
type MockTxMockRecorder struct {
	mock *MockTx
}

// NewMockTx creates a new mock instance.
func NewMockTx(ctrl *gomock.Controller) *MockTx {
	mock := &MockTx{ctrl: ctrl}
	mock.recorder = &MockTxMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTx) EXPECT() *MockTxMockRecorder {
	return m.recorder
}

// ExecContext mocks base method.
func (m *MockTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ExecContext", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecContext indicates an expected call of ExecContext.
func (mr *MockTxMockRecorder) ExecContext(ctx, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecContext", reflect.TypeOf((*MockTx)(nil).ExecContext), varargs...)
}

// ScanStructContext mocks base method.
func (m *MockTx) ScanStructContext(ctx context.Context, i interface{}, query string, args ...interface{}) (bool, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, i, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ScanStructContext", varargs...)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ScanStructContext indicates an expected call of ScanStructContext.
func (mr *MockTxMockRecorder) ScanStructContext(ctx, i, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, i, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScanStructContext", reflect.TypeOf((*MockTx)(nil).ScanStructContext), varargs...)
}
