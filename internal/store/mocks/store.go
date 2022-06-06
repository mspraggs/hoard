// Code generated by MockGen. DO NOT EDIT.
// Source: store.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/mspraggs/hoard/internal/models"
	models0 "github.com/mspraggs/hoard/internal/store/models"
)

// MockEncryptionKeyGenerator is a mock of EncryptionKeyGenerator interface.
type MockEncryptionKeyGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockEncryptionKeyGeneratorMockRecorder
}

// MockEncryptionKeyGeneratorMockRecorder is the mock recorder for MockEncryptionKeyGenerator.
type MockEncryptionKeyGeneratorMockRecorder struct {
	mock *MockEncryptionKeyGenerator
}

// NewMockEncryptionKeyGenerator creates a new mock instance.
func NewMockEncryptionKeyGenerator(ctrl *gomock.Controller) *MockEncryptionKeyGenerator {
	mock := &MockEncryptionKeyGenerator{ctrl: ctrl}
	mock.recorder = &MockEncryptionKeyGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEncryptionKeyGenerator) EXPECT() *MockEncryptionKeyGeneratorMockRecorder {
	return m.recorder
}

// GenerateKey mocks base method.
func (m *MockEncryptionKeyGenerator) GenerateKey(fileUpload *models.FileUpload) (models.EncryptionKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateKey", fileUpload)
	ret0, _ := ret[0].(models.EncryptionKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateKey indicates an expected call of GenerateKey.
func (mr *MockEncryptionKeyGeneratorMockRecorder) GenerateKey(fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateKey", reflect.TypeOf((*MockEncryptionKeyGenerator)(nil).GenerateKey), fileUpload)
}

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockClient) Delete(ctx context.Context, upload *models0.FileUpload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, upload)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockClientMockRecorder) Delete(ctx, upload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockClient)(nil).Delete), ctx, upload)
}

// Upload mocks base method.
func (m *MockClient) Upload(ctx context.Context, upload *models0.FileUpload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", ctx, upload)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upload indicates an expected call of Upload.
func (mr *MockClientMockRecorder) Upload(ctx, upload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockClient)(nil).Upload), ctx, upload)
}
