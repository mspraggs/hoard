// Code generated by MockGen. DO NOT EDIT.
// Source: processor.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	processor "github.com/mspraggs/hoard/internal/processor"
)

// MockKeyGenerator is a mock of KeyGenerator interface.
type MockKeyGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockKeyGeneratorMockRecorder
}

// MockKeyGeneratorMockRecorder is the mock recorder for MockKeyGenerator.
type MockKeyGeneratorMockRecorder struct {
	mock *MockKeyGenerator
}

// NewMockKeyGenerator creates a new mock instance.
func NewMockKeyGenerator(ctrl *gomock.Controller) *MockKeyGenerator {
	mock := &MockKeyGenerator{ctrl: ctrl}
	mock.recorder = &MockKeyGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeyGenerator) EXPECT() *MockKeyGeneratorMockRecorder {
	return m.recorder
}

// GenerateKey mocks base method.
func (m *MockKeyGenerator) GenerateKey() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateKey")
	ret0, _ := ret[0].(string)
	return ret0
}

// GenerateKey indicates an expected call of GenerateKey.
func (mr *MockKeyGeneratorMockRecorder) GenerateKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateKey", reflect.TypeOf((*MockKeyGenerator)(nil).GenerateKey))
}

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

// Create mocks base method.
func (m *MockRegistry) Create(ctx context.Context, file *processor.File) (*processor.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, file)
	ret0, _ := ret[0].(*processor.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockRegistryMockRecorder) Create(ctx, file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockRegistry)(nil).Create), ctx, file)
}

// FetchLatest mocks base method.
func (m *MockRegistry) FetchLatest(ctx context.Context, path string) (*processor.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchLatest", ctx, path)
	ret0, _ := ret[0].(*processor.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchLatest indicates an expected call of FetchLatest.
func (mr *MockRegistryMockRecorder) FetchLatest(ctx, path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchLatest", reflect.TypeOf((*MockRegistry)(nil).FetchLatest), ctx, path)
}

// MockUploader is a mock of Uploader interface.
type MockUploader struct {
	ctrl     *gomock.Controller
	recorder *MockUploaderMockRecorder
}

// MockUploaderMockRecorder is the mock recorder for MockUploader.
type MockUploaderMockRecorder struct {
	mock *MockUploader
}

// NewMockUploader creates a new mock instance.
func NewMockUploader(ctrl *gomock.Controller) *MockUploader {
	mock := &MockUploader{ctrl: ctrl}
	mock.recorder = &MockUploaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUploader) EXPECT() *MockUploaderMockRecorder {
	return m.recorder
}

// Upload mocks base method.
func (m *MockUploader) Upload(ctx context.Context, file *processor.File) (*processor.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", ctx, file)
	ret0, _ := ret[0].(*processor.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upload indicates an expected call of Upload.
func (mr *MockUploaderMockRecorder) Upload(ctx, file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockUploader)(nil).Upload), ctx, file)
}
