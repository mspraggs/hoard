// Code generated by MockGen. DO NOT EDIT.
// Source: processor.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/mspraggs/hoard/internal/models"
)

// MockRegistry is a mock of Registry interface.
type MockRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockRegistryMockRecorder
}

// MockRegistryMockRecorder is the mock recorder for MockRegistry.
type MockRegistryMockRecorder struct {
	mock *MockRegistry
}

// NewMockRegistry creates a new mock instance.
func NewMockRegistry(ctrl *gomock.Controller) *MockRegistry {
	mock := &MockRegistry{ctrl: ctrl}
	mock.recorder = &MockRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegistry) EXPECT() *MockRegistryMockRecorder {
	return m.recorder
}

// GetUploadedFileUpload mocks base method.
func (m *MockRegistry) GetUploadedFileUpload(ctx context.Context, ID string) (*models.FileUpload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUploadedFileUpload", ctx, ID)
	ret0, _ := ret[0].(*models.FileUpload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUploadedFileUpload indicates an expected call of GetUploadedFileUpload.
func (mr *MockRegistryMockRecorder) GetUploadedFileUpload(ctx, ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUploadedFileUpload", reflect.TypeOf((*MockRegistry)(nil).GetUploadedFileUpload), ctx, ID)
}

// MarkFileUploadDeleted mocks base method.
func (m *MockRegistry) MarkFileUploadDeleted(ctx context.Context, fileUpload *models.FileUpload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkFileUploadDeleted", ctx, fileUpload)
	ret0, _ := ret[0].(error)
	return ret0
}

// MarkFileUploadDeleted indicates an expected call of MarkFileUploadDeleted.
func (mr *MockRegistryMockRecorder) MarkFileUploadDeleted(ctx, fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkFileUploadDeleted", reflect.TypeOf((*MockRegistry)(nil).MarkFileUploadDeleted), ctx, fileUpload)
}

// MarkFileUploadUploaded mocks base method.
func (m *MockRegistry) MarkFileUploadUploaded(ctx context.Context, fileUpload *models.FileUpload) (*models.FileUpload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkFileUploadUploaded", ctx, fileUpload)
	ret0, _ := ret[0].(*models.FileUpload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarkFileUploadUploaded indicates an expected call of MarkFileUploadUploaded.
func (mr *MockRegistryMockRecorder) MarkFileUploadUploaded(ctx, fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkFileUploadUploaded", reflect.TypeOf((*MockRegistry)(nil).MarkFileUploadUploaded), ctx, fileUpload)
}

// RegisterFileUpload mocks base method.
func (m *MockRegistry) RegisterFileUpload(ctx context.Context, fileUpload *models.FileUpload) (*models.FileUpload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterFileUpload", ctx, fileUpload)
	ret0, _ := ret[0].(*models.FileUpload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterFileUpload indicates an expected call of RegisterFileUpload.
func (mr *MockRegistryMockRecorder) RegisterFileUpload(ctx, fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterFileUpload", reflect.TypeOf((*MockRegistry)(nil).RegisterFileUpload), ctx, fileUpload)
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

// EraseFileUpload mocks base method.
func (m *MockStore) EraseFileUpload(ctx context.Context, FileUpload *models.FileUpload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EraseFileUpload", ctx, FileUpload)
	ret0, _ := ret[0].(error)
	return ret0
}

// EraseFileUpload indicates an expected call of EraseFileUpload.
func (mr *MockStoreMockRecorder) EraseFileUpload(ctx, FileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EraseFileUpload", reflect.TypeOf((*MockStore)(nil).EraseFileUpload), ctx, FileUpload)
}

// StoreFileUpload mocks base method.
func (m *MockStore) StoreFileUpload(ctx context.Context, FileUpload *models.FileUpload) (*models.FileUpload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreFileUpload", ctx, FileUpload)
	ret0, _ := ret[0].(*models.FileUpload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StoreFileUpload indicates an expected call of StoreFileUpload.
func (mr *MockStoreMockRecorder) StoreFileUpload(ctx, FileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreFileUpload", reflect.TypeOf((*MockStore)(nil).StoreFileUpload), ctx, FileUpload)
}
