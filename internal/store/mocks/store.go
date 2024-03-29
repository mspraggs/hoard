// Code generated by MockGen. DO NOT EDIT.
// Source: store.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	gomock "github.com/golang/mock/gomock"
)

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

// CompleteMultipartUpload mocks base method.
func (m *MockClient) CompleteMultipartUpload(ctx context.Context, input *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, input}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CompleteMultipartUpload", varargs...)
	ret0, _ := ret[0].(*s3.CompleteMultipartUploadOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CompleteMultipartUpload indicates an expected call of CompleteMultipartUpload.
func (mr *MockClientMockRecorder) CompleteMultipartUpload(ctx, input interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, input}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompleteMultipartUpload", reflect.TypeOf((*MockClient)(nil).CompleteMultipartUpload), varargs...)
}

// CreateMultipartUpload mocks base method.
func (m *MockClient) CreateMultipartUpload(ctx context.Context, input *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, input}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateMultipartUpload", varargs...)
	ret0, _ := ret[0].(*s3.CreateMultipartUploadOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMultipartUpload indicates an expected call of CreateMultipartUpload.
func (mr *MockClientMockRecorder) CreateMultipartUpload(ctx, input interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, input}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMultipartUpload", reflect.TypeOf((*MockClient)(nil).CreateMultipartUpload), varargs...)
}

// PutObject mocks base method.
func (m *MockClient) PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, input}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PutObject", varargs...)
	ret0, _ := ret[0].(*s3.PutObjectOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PutObject indicates an expected call of PutObject.
func (mr *MockClientMockRecorder) PutObject(ctx, input interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, input}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutObject", reflect.TypeOf((*MockClient)(nil).PutObject), varargs...)
}

// UploadPart mocks base method.
func (m *MockClient) UploadPart(ctx context.Context, input *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, input}
	for _, a := range optFns {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UploadPart", varargs...)
	ret0, _ := ret[0].(*s3.UploadPartOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadPart indicates an expected call of UploadPart.
func (mr *MockClientMockRecorder) UploadPart(ctx, input interface{}, optFns ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, input}, optFns...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadPart", reflect.TypeOf((*MockClient)(nil).UploadPart), varargs...)
}
