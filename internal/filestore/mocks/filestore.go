// Code generated by MockGen. DO NOT EDIT.
// Source: filestore.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	io "io"
	fs "io/fs"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	filestore "github.com/mspraggs/hoard/internal/filestore"
	models "github.com/mspraggs/hoard/internal/filestore/models"
	models0 "github.com/mspraggs/hoard/internal/models"
)

// MockBucketSelector is a mock of BucketSelector interface.
type MockBucketSelector struct {
	ctrl     *gomock.Controller
	recorder *MockBucketSelectorMockRecorder
}

// MockBucketSelectorMockRecorder is the mock recorder for MockBucketSelector.
type MockBucketSelectorMockRecorder struct {
	mock *MockBucketSelector
}

// NewMockBucketSelector creates a new mock instance.
func NewMockBucketSelector(ctrl *gomock.Controller) *MockBucketSelector {
	mock := &MockBucketSelector{ctrl: ctrl}
	mock.recorder = &MockBucketSelectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBucketSelector) EXPECT() *MockBucketSelectorMockRecorder {
	return m.recorder
}

// SelectBucket mocks base method.
func (m *MockBucketSelector) SelectBucket(fileUpload *models0.FileUpload) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectBucket", fileUpload)
	ret0, _ := ret[0].(string)
	return ret0
}

// SelectBucket indicates an expected call of SelectBucket.
func (mr *MockBucketSelectorMockRecorder) SelectBucket(fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectBucket", reflect.TypeOf((*MockBucketSelector)(nil).SelectBucket), fileUpload)
}

// MockChecksummer is a mock of Checksummer interface.
type MockChecksummer struct {
	ctrl     *gomock.Controller
	recorder *MockChecksummerMockRecorder
}

// MockChecksummerMockRecorder is the mock recorder for MockChecksummer.
type MockChecksummerMockRecorder struct {
	mock *MockChecksummer
}

// NewMockChecksummer creates a new mock instance.
func NewMockChecksummer(ctrl *gomock.Controller) *MockChecksummer {
	mock := &MockChecksummer{ctrl: ctrl}
	mock.recorder = &MockChecksummerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChecksummer) EXPECT() *MockChecksummerMockRecorder {
	return m.recorder
}

// Algorithm mocks base method.
func (m *MockChecksummer) Algorithm() models0.ChecksumAlgorithm {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Algorithm")
	ret0, _ := ret[0].(models0.ChecksumAlgorithm)
	return ret0
}

// Algorithm indicates an expected call of Algorithm.
func (mr *MockChecksummerMockRecorder) Algorithm() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Algorithm", reflect.TypeOf((*MockChecksummer)(nil).Algorithm))
}

// Checksum mocks base method.
func (m *MockChecksummer) Checksum(reader io.Reader) (models0.Checksum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Checksum", reader)
	ret0, _ := ret[0].(models0.Checksum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Checksum indicates an expected call of Checksum.
func (mr *MockChecksummerMockRecorder) Checksum(reader interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Checksum", reflect.TypeOf((*MockChecksummer)(nil).Checksum), reader)
}

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

// Algorithm mocks base method.
func (m *MockEncryptionKeyGenerator) Algorithm() models0.EncryptionAlgorithm {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Algorithm")
	ret0, _ := ret[0].(models0.EncryptionAlgorithm)
	return ret0
}

// Algorithm indicates an expected call of Algorithm.
func (mr *MockEncryptionKeyGeneratorMockRecorder) Algorithm() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Algorithm", reflect.TypeOf((*MockEncryptionKeyGenerator)(nil).Algorithm))
}

// GenerateKey mocks base method.
func (m *MockEncryptionKeyGenerator) GenerateKey(fileUpload *models0.FileUpload) models0.EncryptionKey {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateKey", fileUpload)
	ret0, _ := ret[0].(models0.EncryptionKey)
	return ret0
}

// GenerateKey indicates an expected call of GenerateKey.
func (mr *MockEncryptionKeyGeneratorMockRecorder) GenerateKey(fileUpload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateKey", reflect.TypeOf((*MockEncryptionKeyGenerator)(nil).GenerateKey), fileUpload)
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
func (m *MockUploader) Upload(ctx context.Context, file fs.File, cs filestore.Checksummer, upload *models.FileUpload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", ctx, file, cs, upload)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upload indicates an expected call of Upload.
func (mr *MockUploaderMockRecorder) Upload(ctx, file, cs, upload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockUploader)(nil).Upload), ctx, file, cs, upload)
}

// MockUploaderSelector is a mock of UploaderSelector interface.
type MockUploaderSelector struct {
	ctrl     *gomock.Controller
	recorder *MockUploaderSelectorMockRecorder
}

// MockUploaderSelectorMockRecorder is the mock recorder for MockUploaderSelector.
type MockUploaderSelectorMockRecorder struct {
	mock *MockUploaderSelector
}

// NewMockUploaderSelector creates a new mock instance.
func NewMockUploaderSelector(ctrl *gomock.Controller) *MockUploaderSelector {
	mock := &MockUploaderSelector{ctrl: ctrl}
	mock.recorder = &MockUploaderSelectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUploaderSelector) EXPECT() *MockUploaderSelectorMockRecorder {
	return m.recorder
}

// SelectUploader mocks base method.
func (m *MockUploaderSelector) SelectUploader(file fs.File) (filestore.Uploader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectUploader", file)
	ret0, _ := ret[0].(filestore.Uploader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectUploader indicates an expected call of SelectUploader.
func (mr *MockUploaderSelectorMockRecorder) SelectUploader(file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectUploader", reflect.TypeOf((*MockUploaderSelector)(nil).SelectUploader), file)
}
