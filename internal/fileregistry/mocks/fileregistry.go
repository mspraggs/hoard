// Code generated by MockGen. DO NOT EDIT.
// Source: fileregistry.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	fileregistry "github.com/mspraggs/hoard/internal/fileregistry"
	models "github.com/mspraggs/hoard/internal/models"
)

// MockClock is a mock of Clock interface.
type MockClock struct {
	ctrl     *gomock.Controller
	recorder *MockClockMockRecorder
}

// MockClockMockRecorder is the mock recorder for MockClock.
type MockClockMockRecorder struct {
	mock *MockClock
}

// NewMockClock creates a new mock instance.
func NewMockClock(ctrl *gomock.Controller) *MockClock {
	mock := &MockClock{ctrl: ctrl}
	mock.recorder = &MockClockMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClock) EXPECT() *MockClockMockRecorder {
	return m.recorder
}

// Now mocks base method.
func (m *MockClock) Now() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Now")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// Now indicates an expected call of Now.
func (mr *MockClockMockRecorder) Now() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Now", reflect.TypeOf((*MockClock)(nil).Now))
}

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// GetFileUploadByChangeRequestID mocks base method.
func (m *MockStore) GetFileUploadByChangeRequestID(ctx context.Context, requestID string) (*models.FileUpload, models.ChangeType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileUploadByChangeRequestID", ctx, requestID)
	ret0, _ := ret[0].(*models.FileUpload)
	ret1, _ := ret[1].(models.ChangeType)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetFileUploadByChangeRequestID indicates an expected call of GetFileUploadByChangeRequestID.
func (mr *MockStoreMockRecorder) GetFileUploadByChangeRequestID(ctx, requestID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileUploadByChangeRequestID", reflect.TypeOf((*MockStore)(nil).GetFileUploadByChangeRequestID), ctx, requestID)
}

// InsertFileUpload mocks base method.
func (m *MockStore) InsertFileUpload(ctx context.Context, requestID string, fileUpload *models.FileUpload) (*models.FileUpload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertFileUpload", ctx, requestID, fileUpload)
	ret0, _ := ret[0].(*models.FileUpload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertFileUpload indicates an expected call of InsertFileUpload.
func (mr *MockStoreMockRecorder) InsertFileUpload(ctx, requestID, fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertFileUpload", reflect.TypeOf((*MockStore)(nil).InsertFileUpload), ctx, requestID, fileUpload)
}

// UpdateFileUpload mocks base method.
func (m *MockStore) UpdateFileUpload(ctx context.Context, requestID string, fileUpload *models.FileUpload) (*models.FileUpload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFileUpload", ctx, requestID, fileUpload)
	ret0, _ := ret[0].(*models.FileUpload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateFileUpload indicates an expected call of UpdateFileUpload.
func (mr *MockStoreMockRecorder) UpdateFileUpload(ctx, requestID, fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFileUpload", reflect.TypeOf((*MockStore)(nil).UpdateFileUpload), ctx, requestID, fileUpload)
}

// MockInTransactioner is a mock of InTransactioner interface.
type MockInTransactioner struct {
	ctrl     *gomock.Controller
	recorder *MockInTransactionerMockRecorder
}

// MockInTransactionerMockRecorder is the mock recorder for MockInTransactioner.
type MockInTransactionerMockRecorder struct {
	mock *MockInTransactioner
}

// NewMockInTransactioner creates a new mock instance.
func NewMockInTransactioner(ctrl *gomock.Controller) *MockInTransactioner {
	mock := &MockInTransactioner{ctrl: ctrl}
	mock.recorder = &MockInTransactionerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInTransactioner) EXPECT() *MockInTransactionerMockRecorder {
	return m.recorder
}

// InTransaction mocks base method.
func (m *MockInTransactioner) InTransaction(ctx context.Context, fn fileregistry.TxnFunc) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InTransaction", ctx, fn)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InTransaction indicates an expected call of InTransaction.
func (mr *MockInTransactionerMockRecorder) InTransaction(ctx, fn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InTransaction", reflect.TypeOf((*MockInTransactioner)(nil).InTransaction), ctx, fn)
}
