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
)

type MultiUploaderTestSuite struct {
	suite.Suite
	controller *gomock.Controller
	mockClient *mocks.MockMultiClient
}

func TestMultiUploaderTestSuite(t *testing.T) {
	suite.Run(t, new(MultiUploaderTestSuite))
}

func (s *MultiUploaderTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockClient = mocks.NewMockMultiClient(s.controller)
}

func (s *MultiUploaderTestSuite) TestUpload() {
	empty := ""
	uploadID := "some-upload"
	body := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	bodySize := int64(len(body))
	key := "foo"
	maxChunkSize := 5
	emptyMD5 := "1B2M2Y8AsgTpgAmY7PhCfg=="
	eTags := []string{"one", "two"}

	createMultipartUploadInput := &s3.CreateMultipartUploadInput{
		Key:                  &key,
		Bucket:               &empty,
		SSECustomerKey:       &empty,
		SSECustomerKeyMD5:    &emptyMD5,
		SSECustomerAlgorithm: &empty,
		ChecksumAlgorithm:    types.ChecksumAlgorithm(""),
		StorageClass:         types.StorageClass(""),
	}
	createMultipartUploadOutput := &s3.CreateMultipartUploadOutput{
		UploadId: &uploadID,
	}
	completeUploadInput := &s3.CompleteMultipartUploadInput{
		Key:    &key,
		Bucket: &empty,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: []types.CompletedPart{
				{
					PartNumber: 1,
					ETag:       &eTags[0],
				},
				{
					PartNumber: 2,
					ETag:       &eTags[1],
				},
			},
		},
		SSECustomerKey:       &empty,
		SSECustomerKeyMD5:    &emptyMD5,
		SSECustomerAlgorithm: &empty,
		UploadId:             &uploadID,
	}

	s.Run("reads and uploads file", func() {
		upload := &fsmodels.FileUpload{
			Key:  key,
			Body: bytes.NewReader(body),
		}
		uploadPartInputs := []*s3.UploadPartInput{
			newTestUploadPartInput(upload, uploadID, 1, int64(maxChunkSize)),
			newTestUploadPartInput(upload, uploadID, 2, int64(3)),
		}

		s.mockClient.EXPECT().
			CreateMultipartUpload(context.Background(), createMultipartUploadInput).
			Return(createMultipartUploadOutput, nil)
		gomock.InOrder(
			s.mockClient.EXPECT().
				UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[0])).
				DoAndReturn(s.makeDoUploadPart(body[:maxChunkSize], maxChunkSize, eTags[0])),
			s.mockClient.EXPECT().
				UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[1])).
				DoAndReturn(s.makeDoUploadPart(body[maxChunkSize:], maxChunkSize, eTags[1])),
		)
		s.mockClient.EXPECT().
			CompleteMultipartUpload(context.Background(), completeUploadInput).
			Return(nil, nil)

		uploader := uploader.NewMultiUploader(bodySize, int64(maxChunkSize), s.mockClient)

		err := uploader.Upload(context.Background(), upload)

		s.NoError(err)
	})

	s.Run("returns error for zero chunk size", func() {
		upload := &fsmodels.FileUpload{Key: key}

		uploader := uploader.NewMultiUploader(bodySize, 0, nil)

		err := uploader.Upload(context.Background(), upload)

		s.ErrorContains(err, "invalid chunk size")
	})

	s.Run("wraps and returns error", func() {
		s.Run("from create multipart upload", func() {
			expectedErr := errors.New("oh no")
			upload := &fsmodels.FileUpload{Key: key}

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(nil, expectedErr)

			uploader := uploader.NewMultiUploader(bodySize, 1, s.mockClient)

			err := uploader.Upload(context.Background(), upload)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from upload part", func() {
			expectedErr := errors.New("oh no")
			upload := &fsmodels.FileUpload{Key: key}

			uploadPartInput := newTestUploadPartInput(
				upload, uploadID, 1, int64(maxChunkSize),
			)

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)
			s.mockClient.EXPECT().
				UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInput)).
				Return(nil, expectedErr)

			uploader := uploader.NewMultiUploader(bodySize, int64(maxChunkSize), s.mockClient)

			err := uploader.Upload(context.Background(), upload)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from complete multipart upload", func() {
			expectedErr := errors.New("oh no")
			upload := &fsmodels.FileUpload{
				Key:  key,
				Body: bytes.NewReader(body),
			}
			uploadPartInputs := []*s3.UploadPartInput{
				newTestUploadPartInput(upload, uploadID, 1, int64(maxChunkSize)),
				newTestUploadPartInput(upload, uploadID, 2, int64(3)),
			}

			s.mockClient.EXPECT().
				CreateMultipartUpload(context.Background(), createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)
			gomock.InOrder(
				s.mockClient.EXPECT().
					UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[0])).
					DoAndReturn(s.makeDoUploadPart(body[:maxChunkSize], maxChunkSize, eTags[0])),
				s.mockClient.EXPECT().
					UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInputs[1])).
					DoAndReturn(s.makeDoUploadPart(body[maxChunkSize:], maxChunkSize, eTags[1])),
			)
			s.mockClient.EXPECT().
				CompleteMultipartUpload(context.Background(), completeUploadInput).
				Return(nil, expectedErr)

			uploader := uploader.NewMultiUploader(bodySize, int64(maxChunkSize), s.mockClient)

			err := uploader.Upload(context.Background(), upload)

			s.ErrorIs(err, expectedErr)
		})
	})
}

func (s *MultiUploaderTestSuite) makeDoUploadPart(
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

func newTestUploadPartInput(
	upload *fsmodels.FileUpload,
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
