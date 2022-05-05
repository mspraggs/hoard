package uploader_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/suite"

	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
	"github.com/mspraggs/hoard/internal/filestore/uploader"
	"github.com/mspraggs/hoard/internal/filestore/uploader/mocks"
)

type SingleUploaderTestSuite struct {
	suite.Suite
	controller *gomock.Controller
	mockClient *mocks.MockSingleClient
}

func TestSingleUploaderTestSuite(t *testing.T) {
	suite.Run(t, new(SingleUploaderTestSuite))
}

func (s *SingleUploaderTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockClient = mocks.NewMockSingleClient(s.controller)
}

func (s *SingleUploaderTestSuite) TestUpload() {
	body := []byte{0, 1, 2, 3}

	s.Run("reads and uploads file with checksum", func() {
		upload := &fsmodels.FileUpload{
			Key:  "foo",
			Body: bytes.NewReader(body),
		}
		putObjectInput := newTestPutObjectInput(upload, body)

		s.mockClient.EXPECT().
			PutObject(context.Background(), newPutObjectInputMatcher(putObjectInput)).
			Return(nil, nil)

		uploader := uploader.NewSingleUploader(s.mockClient)

		err := uploader.Upload(context.Background(), upload)

		s.Require().NoError(err)
	})

	s.Run("wraps and returns error", func() {
		s.Run("from put object", func() {
			expectedErr := errors.New("fail")
			upload := &fsmodels.FileUpload{
				Key:  "foo",
				Body: bytes.NewReader(body),
			}
			putObjectInput := newTestPutObjectInput(upload, body)

			s.mockClient.EXPECT().
				PutObject(context.Background(), newPutObjectInputMatcher(putObjectInput)).
				Return(nil, expectedErr)

			uploader := uploader.NewSingleUploader(s.mockClient)

			err := uploader.Upload(context.Background(), upload)

			s.ErrorIs(err, expectedErr)
		})
	})
}

func newTestPutObjectInput(upload *fsmodels.FileUpload, body []byte) *s3.PutObjectInput {
	empty := ""
	return &s3.PutObjectInput{
		Key:                  &upload.Key,
		Bucket:               &empty,
		SSECustomerKey:       &empty,
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
