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

type MultiUploaderTestSuite struct {
	suite.Suite
	controller      *gomock.Controller
	mockClient      *mocks.MockMultiClient
	mockChecksummer *mocks.MockChecksummer
}

type fakeReader struct {
	current int
	max     int
	err     error
}

func (r *fakeReader) Read(bs []byte) (int, error) {
	if r.err != nil {
		return 1, r.err
	}
	for i := range bs {
		if r.current >= r.max {
			return i, nil
		}
		bs[i] = byte(r.current)
		r.current++
	}
	return len(bs), nil
}

func TestMultiUploaderTestSuite(t *testing.T) {
	suite.Run(t, new(MultiUploaderTestSuite))
}

func (s *MultiUploaderTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockClient = mocks.NewMockMultiClient(s.controller)
	s.mockChecksummer = mocks.NewMockChecksummer(s.controller)
}

func (s *MultiUploaderTestSuite) TestUpload() {
	empty := ""
	uploadID := "some-upload"
	upload := &fsmodels.FileUpload{
		Key:               "foo",
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
	}
	body := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	checksums := []string{"one", "two"}
	maxChunkSize := 5

	createMultipartUploadInput := &s3.CreateMultipartUploadInput{
		Key:                  &upload.Key,
		Bucket:               &empty,
		SSECustomerKey:       &empty,
		SSECustomerAlgorithm: &empty,
		ChecksumAlgorithm:    upload.ChecksumAlgorithm,
		StorageClass:         upload.StorageClass,
	}
	createMultipartUploadOutput := &s3.CreateMultipartUploadOutput{
		UploadId: &uploadID,
	}
	completeUploadInput := &s3.CompleteMultipartUploadInput{
		Key:                  &upload.Key,
		Bucket:               &empty,
		SSECustomerKey:       &empty,
		SSECustomerAlgorithm: &empty,
		UploadId:             &uploadID,
	}

	s.Run("reads and uploads file with checksum", func() {
		reader := &fakeReader{0, 8, nil}
		uploadPartInputs := []*s3.UploadPartInput{
			newTestUploadPartInput(upload, uploadID, body[:maxChunkSize], checksums[0]),
			newTestUploadPartInput(upload, uploadID, body[maxChunkSize:], checksums[1]),
		}

		s.mockClient.EXPECT().
			CreateMultipartUpload(context.Background(), createMultipartUploadInput).
			Return(createMultipartUploadOutput, nil)
		gomock.InOrder(
			s.mockChecksummer.EXPECT().
				Checksum(bytes.NewReader(body[:maxChunkSize])).
				Return(models.Checksum(checksums[0]), nil),
			s.mockChecksummer.EXPECT().
				Checksum(bytes.NewReader(body[maxChunkSize:])).
				Return(models.Checksum(checksums[1]), nil),
		)
		gomock.InOrder(
			s.mockClient.EXPECT().
				UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[0])).
				Return(nil, nil),
			s.mockClient.EXPECT().
				UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[1])).
				Return(nil, nil),
		)
		s.mockClient.EXPECT().
			CompleteMultipartUpload(context.Background(), completeUploadInput).
			Return(nil, nil)

		uploader := uploader.NewMultiUploader(int64(maxChunkSize), s.mockClient)

		err := uploader.Upload(context.Background(), reader, s.mockChecksummer, upload)

		s.NoError(err)
	})

	s.Run("wraps and returns error", func() {
		s.Run("from create multipart upload", func() {
			expectedErr := errors.New("oh no")

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(nil, expectedErr)

			uploader := uploader.NewMultiUploader(0, s.mockClient)

			err := uploader.Upload(context.Background(), nil, s.mockChecksummer, upload)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from reader", func() {
			expectedErr := errors.New("oh no")
			reader := &fakeReader{0, 8, expectedErr}

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)

			uploader := uploader.NewMultiUploader(int64(maxChunkSize), s.mockClient)

			err := uploader.Upload(context.Background(), reader, s.mockChecksummer, upload)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from checksum", func() {
			expectedErr := errors.New("oh no")
			reader := &fakeReader{0, 8, nil}

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)
			s.mockChecksummer.EXPECT().
				Checksum(bytes.NewReader(body[:maxChunkSize])).
				Return(models.Checksum(""), expectedErr)

			uploader := uploader.NewMultiUploader(int64(maxChunkSize), s.mockClient)

			err := uploader.Upload(context.Background(), reader, s.mockChecksummer, upload)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from upload part", func() {
			expectedErr := errors.New("oh no")
			reader := &fakeReader{0, 8, nil}

			uploadPartInput := newTestUploadPartInput(
				upload, uploadID, body[:maxChunkSize], checksums[0],
			)

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)
			s.mockChecksummer.EXPECT().
				Checksum(bytes.NewReader(body[:maxChunkSize])).
				Return(models.Checksum(checksums[0]), nil)
			s.mockClient.EXPECT().
				UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInput)).
				Return(nil, expectedErr)

			uploader := uploader.NewMultiUploader(int64(maxChunkSize), s.mockClient)

			err := uploader.Upload(context.Background(), reader, s.mockChecksummer, upload)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from complete multipart upload", func() {
			expectedErr := errors.New("oh no")
			reader := &fakeReader{0, 8, nil}
			uploadPartInputs := []*s3.UploadPartInput{
				newTestUploadPartInput(upload, uploadID, body[:maxChunkSize], checksums[0]),
				newTestUploadPartInput(upload, uploadID, body[maxChunkSize:], checksums[1]),
			}

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)
			gomock.InOrder(
				s.mockChecksummer.EXPECT().
					Checksum(bytes.NewReader(body[:maxChunkSize])).
					Return(models.Checksum(checksums[0]), nil),
				s.mockChecksummer.EXPECT().
					Checksum(bytes.NewReader(body[maxChunkSize:])).
					Return(models.Checksum(checksums[1]), nil),
			)
			gomock.InOrder(
				s.mockClient.EXPECT().
					UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[0])).
					Return(nil, nil),
				s.mockClient.EXPECT().
					UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[1])).
					Return(nil, nil),
			)
			s.mockClient.EXPECT().
				CompleteMultipartUpload(context.Background(), completeUploadInput).
				Return(nil, expectedErr)

			uploader := uploader.NewMultiUploader(int64(maxChunkSize), s.mockClient)

			err := uploader.Upload(context.Background(), reader, s.mockChecksummer, upload)

			s.ErrorIs(err, expectedErr)
		})
	})
}

func newTestUploadPartInput(
	upload *fsmodels.FileUpload,
	uploadID string,
	body []byte,
	checksum string,
) *s3.UploadPartInput {

	empty := ""
	return &s3.UploadPartInput{
		Key:                  &upload.Key,
		UploadId:             &uploadID,
		Bucket:               &empty,
		SSECustomerKey:       &empty,
		SSECustomerAlgorithm: &empty,
		ChecksumAlgorithm:    upload.ChecksumAlgorithm,
		ChecksumSHA256:       &checksum,
		Body:                 bytes.NewReader(body),
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
		cmpopts.IgnoreUnexported(s3.UploadPartInput{}),
		cmpopts.IgnoreFields(s3.UploadPartInput{}, "Body"),
	)
}

func (m *uploadPartInputMatcher) String() string {
	return fmt.Sprintf("is equal to %v", m.expected)
}
