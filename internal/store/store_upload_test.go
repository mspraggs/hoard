package store_test

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"

	"github.com/mspraggs/hoard/internal/processor"
	"github.com/mspraggs/hoard/internal/store"
)

func (s *StoreTestSuite) TestUpload() {
	key := "some-key"
	path := "some/path"
	body := []byte{0, 1, 2, 3}
	salt := []byte{4, 5, 6}

	bucket := "some-bucket"
	checksumAlgorithm := types.ChecksumAlgorithmCrc32
	encryptionAlgorithm := processor.EncryptionAlgorithmAES256

	encKey := []byte{8, 9, 0}
	keyParams := "key-params"
	eTag := "some-etag"
	version := "some-version"

	sseKey := base64.StdEncoding.EncodeToString(encKey)
	hashedEncKey := md5.Sum(encKey)
	sseKeyMD5 := base64.StdEncoding.EncodeToString(hashedEncKey[:])
	sseAlg := string(types.ServerSideEncryptionAes256)

	fs, err := newMemFS(map[string][]byte{path: body})
	s.Require().NoError(err)

	inputFile := &processor.File{
		Key:       key,
		LocalPath: path,
	}
	expectedOutputFile := &processor.File{
		Key:                 key,
		LocalPath:           path,
		Bucket:              bucket,
		Salt:                salt,
		EncryptionAlgorithm: encryptionAlgorithm,
		KeyParams:           keyParams,
		ETag:                eTag,
		Version:             version,
	}

	s.Run("given file smaller than chunk size", func() {
		chunksize := int64(5)

		s.Run("reads and uploads versioned file", func() {
			ctx := context.WithValue(context.Background(), contextKey("key"), "value")

			file, err := fs.Open(path)
			s.Require().NoError(err)
			defer file.Close()

			putObjectInput := newTestPutObjectInput(
				expectedOutputFile,
				bucket,
				encKey,
				checksumAlgorithm,
				file,
			)
			putObjectOutput := &s3.PutObjectOutput{
				ETag:      &eTag,
				VersionId: &version,
			}

			s.mockSalter.EXPECT().
				Salt().Return(salt)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(salt).Return(encKey)
			s.mockEncKeyGen.EXPECT().
				String().Return(keyParams)
			s.mockClient.EXPECT().
				PutObject(ctx, newPutObjectInputMatcher(putObjectInput)).
				Return(putObjectOutput, nil)

			store := store.New(
				s.mockClient,
				fs,
				s.mockEncKeyGen,
				s.mockSalter,
				bucket,
				store.WithChunkSize(chunksize),
			)

			outputFile, err := store.Upload(ctx, inputFile)

			s.Require().NoError(err)
			s.Equal(expectedOutputFile, outputFile)
		})
		s.Run("reads and uploads unversioned file", func() {
			ctx := context.WithValue(context.Background(), contextKey("key"), "value")

			expectedOutputFile := &processor.File{
				Key:                 key,
				LocalPath:           path,
				Bucket:              bucket,
				Salt:                salt,
				EncryptionAlgorithm: encryptionAlgorithm,
				KeyParams:           keyParams,
				ETag:                eTag,
			}

			file, err := fs.Open(path)
			s.Require().NoError(err)
			defer file.Close()

			putObjectInput := newTestPutObjectInput(
				expectedOutputFile,
				bucket,
				encKey,
				checksumAlgorithm,
				file,
			)
			putObjectOutput := &s3.PutObjectOutput{
				ETag:      &eTag,
				VersionId: nil,
			}

			s.mockSalter.EXPECT().
				Salt().Return(salt)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(salt).Return(encKey)
			s.mockEncKeyGen.EXPECT().
				String().Return(keyParams)
			s.mockClient.EXPECT().
				PutObject(ctx, newPutObjectInputMatcher(putObjectInput)).
				Return(putObjectOutput, nil)

			store := store.New(
				s.mockClient,
				fs,
				s.mockEncKeyGen,
				s.mockSalter,
				bucket,
				store.WithChunkSize(chunksize),
			)

			outputFile, err := store.Upload(ctx, inputFile)

			s.Require().NoError(err)
			s.Equal(expectedOutputFile, outputFile)
		})
		s.Run("handles error from client", func() {
			expectedErr := errors.New("oh no")
			ctx := context.WithValue(context.Background(), contextKey("key"), "value")

			file, err := fs.Open(path)
			s.Require().NoError(err)
			defer file.Close()

			putObjectInput := newTestPutObjectInput(
				expectedOutputFile,
				bucket,
				encKey,
				checksumAlgorithm,
				file,
			)

			s.mockSalter.EXPECT().
				Salt().Return(salt)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(salt).Return(encKey)
			s.mockEncKeyGen.EXPECT().
				String().Return(keyParams)
			s.mockClient.EXPECT().
				PutObject(ctx, newPutObjectInputMatcher(putObjectInput)).
				Return(nil, expectedErr)

			store := store.New(
				s.mockClient,
				fs,
				s.mockEncKeyGen,
				s.mockSalter,
				bucket,
				store.WithChunkSize(chunksize),
			)

			outputFile, err := store.Upload(ctx, inputFile)

			s.Nil(outputFile)
			s.ErrorIs(err, expectedErr)
		})
	})

	s.Run("given file larger than chunk size", func() {
		uploadID := "some-upload-id"
		eTags := []string{"one", "two"}
		chunksize := int64(3)

		createMultipartUploadInput := &s3.CreateMultipartUploadInput{
			Key:                  &key,
			Bucket:               &bucket,
			SSECustomerKey:       &sseKey,
			SSECustomerKeyMD5:    &sseKeyMD5,
			SSECustomerAlgorithm: &sseAlg,
			ChecksumAlgorithm:    types.ChecksumAlgorithmCrc32,
			StorageClass:         types.StorageClassStandard,
		}
		createMultipartUploadOutput := &s3.CreateMultipartUploadOutput{
			UploadId: &uploadID,
		}
		completeUploadInput := &s3.CompleteMultipartUploadInput{
			Key:    &key,
			Bucket: &bucket,
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
			SSECustomerKey:       &sseKey,
			SSECustomerKeyMD5:    &sseKeyMD5,
			SSECustomerAlgorithm: &sseAlg,
			UploadId:             &uploadID,
		}
		completeUploadOutput := &s3.CompleteMultipartUploadOutput{
			ETag:      &eTag,
			VersionId: &version,
		}

		s.Run("reads and uploads versioned file", func() {
			ctx := context.WithValue(context.Background(), contextKey("key"), "value")

			file, err := fs.Open(path)
			s.Require().NoError(err)
			defer file.Close()

			uploadPartInputs := []*s3.UploadPartInput{
				newTestUploadPartInput(
					inputFile, bucket, encKey, uploadID, checksumAlgorithm, 1,
					chunksize, bytes.NewReader(body[:chunksize]),
				),
				newTestUploadPartInput(
					inputFile, bucket, encKey, uploadID, checksumAlgorithm, 2,
					int64(1), bytes.NewReader(body[chunksize:]),
				),
			}

			s.mockSalter.EXPECT().
				Salt().Return(salt)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(salt).Return(encKey)
			s.mockEncKeyGen.EXPECT().
				String().Return(keyParams)
			s.mockClient.EXPECT().
				CreateMultipartUpload(ctx, createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)
			gomock.InOrder(
				s.mockClient.EXPECT().
					UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[0])).
					DoAndReturn(s.makeDoUploadPart(body[:chunksize], chunksize, eTags[0])),
				s.mockClient.EXPECT().
					UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[1])).
					DoAndReturn(s.makeDoUploadPart(body[chunksize:], chunksize, eTags[1])),
			)
			s.mockClient.EXPECT().
				CompleteMultipartUpload(ctx, completeUploadInput).
				Return(completeUploadOutput, nil)

			store := store.New(
				s.mockClient,
				fs,
				s.mockEncKeyGen,
				s.mockSalter,
				bucket,
				store.WithChunkSize(chunksize),
			)

			outputFile, err := store.Upload(ctx, inputFile)

			s.Require().NoError(err)
			s.Equal(expectedOutputFile, outputFile)
		})
		s.Run("reads and uploads unversioned file", func() {
			ctx := context.WithValue(context.Background(), contextKey("key"), "value")

			expectedOutputFile := &processor.File{
				Key:                 key,
				LocalPath:           path,
				Bucket:              bucket,
				Salt:                salt,
				EncryptionAlgorithm: encryptionAlgorithm,
				KeyParams:           keyParams,
				ETag:                eTag,
			}

			file, err := fs.Open(path)
			s.Require().NoError(err)
			defer file.Close()

			uploadPartInputs := []*s3.UploadPartInput{
				newTestUploadPartInput(
					inputFile, bucket, encKey, uploadID, checksumAlgorithm, 1,
					chunksize, bytes.NewReader(body[:chunksize]),
				),
				newTestUploadPartInput(
					inputFile, bucket, encKey, uploadID, checksumAlgorithm, 2,
					int64(1), bytes.NewReader(body[chunksize:]),
				),
			}

			completeUploadOutput := &s3.CompleteMultipartUploadOutput{
				ETag: &eTag,
			}

			s.mockSalter.EXPECT().
				Salt().Return(salt)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(salt).Return(encKey)
			s.mockEncKeyGen.EXPECT().
				String().Return(keyParams)
			s.mockClient.EXPECT().
				CreateMultipartUpload(ctx, createMultipartUploadInput).
				Return(createMultipartUploadOutput, nil)
			gomock.InOrder(
				s.mockClient.EXPECT().
					UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[0])).
					DoAndReturn(s.makeDoUploadPart(body[:chunksize], chunksize, eTags[0])),
				s.mockClient.EXPECT().
					UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[1])).
					DoAndReturn(s.makeDoUploadPart(body[chunksize:], chunksize, eTags[1])),
			)
			s.mockClient.EXPECT().
				CompleteMultipartUpload(ctx, completeUploadInput).
				Return(completeUploadOutput, nil)

			store := store.New(
				s.mockClient,
				fs,
				s.mockEncKeyGen,
				s.mockSalter,
				bucket,
				store.WithChunkSize(chunksize),
			)

			outputFile, err := store.Upload(ctx, inputFile)

			s.Require().NoError(err)
			s.Equal(expectedOutputFile, outputFile)
		})
		s.Run("handles error", func() {
			expectedErr := errors.New("oh no")

			s.Run("from create upload", func() {
				ctx := context.WithValue(context.Background(), contextKey("key"), "value")

				file, err := fs.Open(path)
				s.Require().NoError(err)
				defer file.Close()

				s.mockSalter.EXPECT().
					Salt().Return(salt)
				s.mockEncKeyGen.EXPECT().
					GenerateKey(salt).Return(encKey)
				s.mockEncKeyGen.EXPECT().
					String().Return(keyParams)
				s.mockClient.EXPECT().
					CreateMultipartUpload(ctx, createMultipartUploadInput).
					Return(nil, expectedErr)

				store := store.New(
					s.mockClient,
					fs,
					s.mockEncKeyGen,
					s.mockSalter,
					bucket,
					store.WithChunkSize(chunksize),
				)

				outputFile, err := store.Upload(ctx, inputFile)

				s.Nil(outputFile)
				s.ErrorIs(err, expectedErr)
			})
			s.Run("from upload part", func() {
				ctx := context.WithValue(context.Background(), contextKey("key"), "value")

				file, err := fs.Open(path)
				s.Require().NoError(err)
				defer file.Close()

				uploadPartInputs := []*s3.UploadPartInput{
					newTestUploadPartInput(
						inputFile, bucket, encKey, uploadID, checksumAlgorithm, 1,
						chunksize, bytes.NewReader(body[:chunksize]),
					),
					newTestUploadPartInput(
						inputFile, bucket, encKey, uploadID, checksumAlgorithm, 2,
						int64(1), bytes.NewReader(body[chunksize:]),
					),
				}

				s.mockSalter.EXPECT().
					Salt().Return(salt)
				s.mockEncKeyGen.EXPECT().
					GenerateKey(salt).Return(encKey)
				s.mockEncKeyGen.EXPECT().
					String().Return(keyParams)
				s.mockClient.EXPECT().
					CreateMultipartUpload(ctx, createMultipartUploadInput).
					Return(createMultipartUploadOutput, nil)
				gomock.InOrder(
					s.mockClient.EXPECT().
						UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[0])).
						DoAndReturn(s.makeDoUploadPart(body[:chunksize], chunksize, eTags[0])),
					s.mockClient.EXPECT().
						UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[1])).
						Return(nil, expectedErr),
				)

				store := store.New(
					s.mockClient,
					fs,
					s.mockEncKeyGen,
					s.mockSalter,
					bucket,
					store.WithChunkSize(chunksize),
				)

				outputFile, err := store.Upload(ctx, inputFile)

				s.Nil(outputFile)
				s.ErrorIs(err, expectedErr)
			})
			s.Run("from complete upload", func() {
				ctx := context.WithValue(context.Background(), contextKey("key"), "value")

				file, err := fs.Open(path)
				s.Require().NoError(err)
				defer file.Close()

				uploadPartInputs := []*s3.UploadPartInput{
					newTestUploadPartInput(
						inputFile, bucket, encKey, uploadID, checksumAlgorithm, 1,
						chunksize, bytes.NewReader(body[:chunksize]),
					),
					newTestUploadPartInput(
						inputFile, bucket, encKey, uploadID, checksumAlgorithm, 2,
						int64(1), bytes.NewReader(body[chunksize:]),
					),
				}

				s.mockSalter.EXPECT().
					Salt().Return(salt)
				s.mockEncKeyGen.EXPECT().
					GenerateKey(salt).Return(encKey)
				s.mockEncKeyGen.EXPECT().
					String().Return(keyParams)
				s.mockClient.EXPECT().
					CreateMultipartUpload(ctx, createMultipartUploadInput).
					Return(createMultipartUploadOutput, nil)
				gomock.InOrder(
					s.mockClient.EXPECT().
						UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[0])).
						DoAndReturn(s.makeDoUploadPart(body[:chunksize], chunksize, eTags[0])),
					s.mockClient.EXPECT().
						UploadPart(ctx, newUploadPartInputMatcher(uploadPartInputs[1])).
						DoAndReturn(s.makeDoUploadPart(body[chunksize:], chunksize, eTags[1])),
				)
				s.mockClient.EXPECT().
					CompleteMultipartUpload(ctx, completeUploadInput).
					Return(nil, expectedErr)

				store := store.New(
					s.mockClient,
					fs,
					s.mockEncKeyGen,
					s.mockSalter,
					bucket,
					store.WithChunkSize(chunksize),
				)

				outputFile, err := store.Upload(ctx, inputFile)

				s.Nil(outputFile)
				s.ErrorIs(err, expectedErr)
			})
		})
	})

	s.Run("handles error", func() {
		s.Run("when opening file", func() {
			ctx := context.WithValue(context.Background(), contextKey("key"), "value")

			inputFile := &processor.File{
				LocalPath: "doesnt/exist",
			}

			store := store.New(nil, fs, nil, nil, "")

			outputFile, err := store.Upload(ctx, inputFile)

			s.Nil(outputFile)
			s.ErrorIs(err, os.ErrNotExist)
		})
		s.Run("when getting file info", func() {
			expectedErr := errors.New("oh no")

			ctx := context.WithValue(context.Background(), contextKey("key"), "value")

			fs := &fakeFS{expectedErr}
			s.mockSalter.EXPECT().
				Salt().Return(salt)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(salt).Return(encKey)
			s.mockEncKeyGen.EXPECT().
				String().Return(keyParams)

			store := store.New(nil, fs, s.mockEncKeyGen, s.mockSalter, "")

			outputFile, err := store.Upload(ctx, inputFile)

			s.Nil(outputFile)
			s.ErrorIs(err, expectedErr)
		})
	})

	s.Run("returns error for invalid chunk size", func() {
		store := store.New(nil, nil, nil, nil, "", store.WithChunkSize(0))

		outputFile, err := store.Upload(context.Background(), nil)

		s.Nil(outputFile)
		s.ErrorContains(err, "invalid chunk size")
	})
}
