package uploader_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/suite"

	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
	"github.com/mspraggs/hoard/internal/filestore/uploader"
	"github.com/mspraggs/hoard/internal/filestore/uploader/mocks"
	"github.com/mspraggs/hoard/internal/models"
)

type SingleUploaderTestSuite struct {
	suite.Suite
	controller      *gomock.Controller
	mockClient      *mocks.MockSingleClient
	mockChecksummer *mocks.MockChecksummer
}

type mockReader func([]byte) (int, error)

func (f mockReader) Read(b []byte) (int, error) {
	return f(b)
}

func TestSingleUploaderTestSuite(t *testing.T) {
	suite.Run(t, new(SingleUploaderTestSuite))
}

func (s *SingleUploaderTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockClient = mocks.NewMockSingleClient(s.controller)
	s.mockChecksummer = mocks.NewMockChecksummer(s.controller)
}

func (s *SingleUploaderTestSuite) TestUpload() {
	body := []byte{0, 1, 2, 3}
	checksum := "5678"
	upload := &fsmodels.FileUpload{
		Key:               "foo",
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
	}

	reader := func(bs []byte) (int, error) {
		for i, b := range body {
			bs[i] = b
		}
		return len(body), io.EOF
	}

	s.Run("reads and uploads file with checksum", func() {
		putObjectInput := newTestPutObjectInput(upload, body, checksum)

		s.mockChecksummer.EXPECT().
			Checksum(bytes.NewReader(body)).Return(models.Checksum(checksum), nil)
		s.mockClient.EXPECT().
			PutObject(context.Background(), newPutObjectInputMatcher(putObjectInput)).
			Return(nil, nil)

		uploader := uploader.NewSingleUploader(s.mockClient)

		err := uploader.Upload(context.Background(), mockReader(reader), s.mockChecksummer, upload)

		s.Require().NoError(err)
	})

	s.Run("wraps and returns error", func() {
		s.Run("from reader", func() {
			expectedErr := errors.New("fail")

			reader := func(bs []byte) (int, error) {
				return 0, expectedErr
			}

			uploader := uploader.NewSingleUploader(s.mockClient)

			err := uploader.Upload(context.Background(), mockReader(reader), nil, nil)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from checksummer", func() {
			expectedErr := errors.New("fail")

			s.mockChecksummer.EXPECT().
				Checksum(bytes.NewReader(body)).Return(models.Checksum(""), expectedErr)

			uploader := uploader.NewSingleUploader(s.mockClient)

			err := uploader.Upload(context.Background(), mockReader(reader), s.mockChecksummer, nil)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from put object", func() {
			expectedErr := errors.New("fail")
			putObjectInput := newTestPutObjectInput(upload, body, checksum)

			s.mockChecksummer.EXPECT().
				Checksum(bytes.NewReader(body)).Return(models.Checksum(checksum), nil)
			s.mockClient.EXPECT().
				PutObject(context.Background(), newPutObjectInputMatcher(putObjectInput)).
				Return(nil, expectedErr)

			uploader := uploader.NewSingleUploader(s.mockClient)

			err := uploader.Upload(context.Background(), mockReader(reader), s.mockChecksummer, upload)

			s.ErrorIs(err, expectedErr)
		})
	})
}

func newTestPutObjectInput(
	upload *fsmodels.FileUpload,
	body []byte,
	checksum string,
) *s3.PutObjectInput {

	empty := ""
	return &s3.PutObjectInput{
		Key:                  &upload.Key,
		Bucket:               &empty,
		SSECustomerKey:       &empty,
		SSECustomerAlgorithm: &empty,
		ChecksumAlgorithm:    upload.ChecksumAlgorithm,
		Body:                 bytes.NewBuffer(body),
		ChecksumSHA256:       &checksum,
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
