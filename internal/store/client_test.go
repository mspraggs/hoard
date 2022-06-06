package store_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/store/mocks"
	"github.com/mspraggs/hoard/internal/store/models"
)

type AWSClientTestSuite struct {
	suite.Suite
	controller *gomock.Controller
	mockClient *mocks.MockBackendClient
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(AWSClientTestSuite))
}

func (s *AWSClientTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockClient = mocks.NewMockBackendClient(s.controller)
}

func (s *AWSClientTestSuite) makeDoUploadPart(
	expectedBytes []byte,
	maxChunkSize int,
	eTag string,
) func(context.Context, *s3.UploadPartInput, ...func(*s3.Options)) (*s3.UploadPartOutput, error) {

	return func(
		ctx context.Context,
		upi *s3.UploadPartInput,
		optFns ...func(*s3.Options),
	) (*s3.UploadPartOutput, error) {

		bs := make([]byte, maxChunkSize)
		n, err := upi.Body.Read(bs)
		s.Require().Equal(expectedBytes, bs[:n])
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			return nil, err
		}
		return &s3.UploadPartOutput{ETag: &eTag}, nil
	}
}

func newTestPutObjectInput(upload *models.FileUpload, body []byte) *s3.PutObjectInput {
	empty := ""
	emptyMD5 := "1B2M2Y8AsgTpgAmY7PhCfg=="
	return &s3.PutObjectInput{
		Key:                  &upload.Key,
		Bucket:               &empty,
		SSECustomerKey:       &empty,
		SSECustomerKeyMD5:    &emptyMD5,
		SSECustomerAlgorithm: &empty,
		ChecksumAlgorithm:    upload.ChecksumAlgorithm,
		Body:                 bytes.NewReader(body),
	}
}

type putObjectInputMatcher struct {
	expected *s3.PutObjectInput
}

func newPutObjectInputMatcher(expected *s3.PutObjectInput) *putObjectInputMatcher {
	return &putObjectInputMatcher{expected}
}

func (m *putObjectInputMatcher) Matches(actual interface{}) bool {
	actualInput, ok := actual.(*s3.PutObjectInput)
	if !ok {
		return false
	}

	actualBody, err := io.ReadAll(actualInput.Body)
	if err != nil {
		return false
	}

	expectedBody, err := io.ReadAll(m.expected.Body)
	if err != nil {
		return false
	}

	if bytes.Compare(expectedBody, actualBody) != 0 {
		return false
	}

	return cmp.Equal(
		m.expected, actualInput,
		cmpopts.IgnoreUnexported(s3.PutObjectInput{}),
		cmpopts.IgnoreFields(s3.PutObjectInput{}, "Body"),
	)
}

func (m *putObjectInputMatcher) String() string {
	return fmt.Sprintf("is equal to %v", m.expected)
}

func newTestUploadPartInput(
	upload *models.FileUpload,
	uploadID string,
	chunkNum int32,
	chunkSize int64,
) *s3.UploadPartInput {

	empty := ""
	emptyMD5 := "1B2M2Y8AsgTpgAmY7PhCfg=="
	return &s3.UploadPartInput{
		Key:                  &upload.Key,
		UploadId:             &uploadID,
		Bucket:               &empty,
		PartNumber:           chunkNum,
		ContentLength:        chunkSize,
		SSECustomerKey:       &empty,
		SSECustomerKeyMD5:    &emptyMD5,
		SSECustomerAlgorithm: &empty,
		ChecksumAlgorithm:    upload.ChecksumAlgorithm,
		Body:                 &io.LimitedReader{R: upload.Body, N: chunkSize},
	}
}

type uploadPartInputMatcher struct {
	expected *s3.UploadPartInput
}

func newUploadPartInputMatcher(expected *s3.UploadPartInput) *uploadPartInputMatcher {
	return &uploadPartInputMatcher{expected}
}

func (m *uploadPartInputMatcher) Matches(actual interface{}) bool {
	actualInput, ok := actual.(*s3.UploadPartInput)
	if !ok {
		return false
	}

	return cmp.Equal(
		m.expected, actualInput,
		cmpopts.IgnoreUnexported(s3.UploadPartInput{}),
		cmpopts.IgnoreFields(s3.UploadPartInput{}, "Body"),
	)
}

func (m *uploadPartInputMatcher) String() string {
	return fmt.Sprintf("is equal to %v", m.expected)
}
