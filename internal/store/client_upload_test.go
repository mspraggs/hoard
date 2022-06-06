package store_test

import (
	"bytes"
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"

	"github.com/mspraggs/hoard/internal/store"
	"github.com/mspraggs/hoard/internal/store/models"
)

func (s *AWSClientTestSuite) TestUpload() {
	chunksize := int64(5)

	s.Run("given file smaller than chunk size", func() {
		body := []byte{0, 1, 2, 3}

		s.Run("reads and uploads file", func() {
			upload := &models.FileUpload{
				Key:  "foo",
				Size: int64(len(body)),
				Body: bytes.NewReader(body),
			}
			putObjectInput := newTestPutObjectInput(upload, body)

			s.mockClient.EXPECT().
				PutObject(context.Background(), newPutObjectInputMatcher(putObjectInput)).
				Return(nil, nil)

			uploader := store.NewAWSClient(chunksize, s.mockClient)

			err := uploader.Upload(context.Background(), upload)

			s.Require().NoError(err)
		})

		s.Run("wraps and returns error", func() {
			s.Run("from put object", func() {
				expectedErr := errors.New("fail")
				upload := &models.FileUpload{
					Key:  "foo",
					Size: int64(len(body)),
					Body: bytes.NewReader(body),
				}
				putObjectInput := newTestPutObjectInput(upload, body)

				s.mockClient.EXPECT().
					PutObject(context.Background(), newPutObjectInputMatcher(putObjectInput)).
					Return(nil, expectedErr)

				uploader := store.NewAWSClient(chunksize, s.mockClient)

				err := uploader.Upload(context.Background(), upload)

				s.ErrorIs(err, expectedErr)
			})
		})
	})

	s.Run("given file larger than chunk size", func() {
		empty := ""
		uploadID := "some-upload"
		body := []byte{0, 1, 2, 3, 4, 5, 6, 7}
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
			upload := &models.FileUpload{
				Key:  key,
				Size: int64(len(body)),
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

			uploader := store.NewAWSClient(chunksize, s.mockClient)

			err := uploader.Upload(context.Background(), upload)

			s.NoError(err)
		})

		s.Run("returns error for zero chunk size", func() {
			upload := &models.FileUpload{
				Key:  key,
				Size: int64(len(body)),
			}

			uploader := store.NewAWSClient(0, nil)

			err := uploader.Upload(context.Background(), upload)

			s.ErrorContains(err, "invalid chunk size")
		})

		s.Run("wraps and returns error", func() {
			s.Run("from create multipart upload", func() {
				expectedErr := errors.New("oh no")
				upload := &models.FileUpload{
					Key:  key,
					Size: int64(len(body)),
				}

				s.mockClient.EXPECT().
					CreateMultipartUpload(context.Background(), createMultipartUploadInput).
					Return(nil, expectedErr)

				uploader := store.NewAWSClient(1, s.mockClient)

				err := uploader.Upload(context.Background(), upload)

				s.ErrorIs(err, expectedErr)
			})
			s.Run("from upload part", func() {
				expectedErr := errors.New("oh no")
				upload := &models.FileUpload{
					Key:  key,
					Size: int64(len(body)),
				}

				uploadPartInput := newTestUploadPartInput(
					upload, uploadID, 1, int64(maxChunkSize),
				)

				s.mockClient.EXPECT().
					CreateMultipartUpload(context.Background(), createMultipartUploadInput).
					Return(createMultipartUploadOutput, nil)
				s.mockClient.EXPECT().
					UploadPart(context.Background(), newUploadPartInputMatcher(uploadPartInput)).
					Return(nil, expectedErr)

				uploader := store.NewAWSClient(chunksize, s.mockClient)

				err := uploader.Upload(context.Background(), upload)

				s.ErrorIs(err, expectedErr)
			})
			s.Run("from complete multipart upload", func() {
				expectedErr := errors.New("oh no")
				upload := &models.FileUpload{
					Key:  key,
					Size: int64(len(body)),
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

				uploader := store.NewAWSClient(chunksize, s.mockClient)

				err := uploader.Upload(context.Background(), upload)

				s.ErrorIs(err, expectedErr)
			})
		})
	})
}
