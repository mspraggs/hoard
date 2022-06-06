package store_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mspraggs/hoard/internal/models"
	fsmodels "github.com/mspraggs/hoard/internal/store/models"
)

func (s *FilestoreTestSuite) TestStoreFileUpload() {
	s.Run("uploads file with key and returns upload", func() {
		fileID := "some-file"
		path := "path/to/file"
		body := []byte{1, 2, 3}
		encKey := models.EncryptionKey(body)

		fs, err := newMemFS(map[string][]byte{path: body})
		s.Require().NoError(err)

		businessFileUpload := &models.FileUpload{
			ID:        fileID,
			LocalPath: path,
		}
		fsFileUpload := &fsmodels.FileUpload{
			Key:                 fileID,
			EncryptionKey:       fsmodels.EncryptionKey(body),
			EncryptionAlgorithm: types.ServerSideEncryptionAes256,
			ChecksumAlgorithm:   types.ChecksumAlgorithmSha256,
			StorageClass:        types.StorageClassStandard,
			Size:                3,
			Body:                bytes.NewReader(body),
		}

		s.mockEncKeyGen.EXPECT().
			GenerateKey(businessFileUpload).Return(encKey, nil)

		s.mockClient.EXPECT().
			Upload(context.Background(), newFileUploadMatcher(fsFileUpload)).
			Return(nil)

		store := s.newStore(fs)

		uploadedFileUpload, err := store.StoreFileUpload(context.Background(), businessFileUpload)

		s.Require().NoError(err)
		s.Require().Equal(businessFileUpload, uploadedFileUpload)
	})

	s.Run("wraps and returns error", func() {
		s.Run("from filesystem Open", func() {
			expectedErr := fs.ErrNotExist
			fileID := "some-file"
			path := "path/to/file"

			fs, err := newMemFS(map[string][]byte{})
			s.Require().NoError(err)

			businessFileUpload := &models.FileUpload{
				ID:        fileID,
				LocalPath: path,
			}

			store := s.newStore(fs)

			uploadedFileUpload, err := store.StoreFileUpload(
				context.Background(),
				businessFileUpload,
			)

			s.Require().Nil(uploadedFileUpload)
			s.Require().ErrorIs(err, expectedErr)
		})
		s.Run("from encryption key generator", func() {
			expectedErr := errors.New("oh no")
			fileID := "some-file"
			path := "path/to/file"
			encKey := models.EncryptionKey([]byte{})

			businessFileUpload := &models.FileUpload{
				ID:        fileID,
				LocalPath: path,
			}

			fs, err := newMemFS(map[string][]byte{path: {}})
			s.Require().NoError(err)

			s.mockEncKeyGen.EXPECT().GenerateKey(businessFileUpload).Return(encKey, expectedErr)

			store := s.newStore(fs)

			uploadedFileUpload, err := store.StoreFileUpload(
				context.Background(),
				businessFileUpload,
			)

			s.Require().Nil(uploadedFileUpload)
			s.Require().ErrorIs(err, expectedErr)
		})
		s.Run("from client", func() {
			expectedErr := errors.New("oh no")
			fileID := "some-file"
			path := "path/to/file"
			body := []byte{1, 2, 3}
			encKey := models.EncryptionKey(body)

			fs, err := newMemFS(map[string][]byte{path: body})
			s.Require().NoError(err)

			businessFileUpload := &models.FileUpload{
				ID:        fileID,
				LocalPath: path,
			}
			fsFileUpload := &fsmodels.FileUpload{
				Key:                 fileID,
				EncryptionKey:       fsmodels.EncryptionKey(body),
				EncryptionAlgorithm: types.ServerSideEncryptionAes256,
				ChecksumAlgorithm:   types.ChecksumAlgorithmSha256,
				StorageClass:        types.StorageClassStandard,
				Size:                int64(len(body)),
				Body:                bytes.NewReader(body),
			}

			s.mockEncKeyGen.EXPECT().
				GenerateKey(businessFileUpload).Return(encKey, nil)

			s.mockClient.EXPECT().
				Upload(context.Background(), newFileUploadMatcher(fsFileUpload)).
				Return(expectedErr)

			store := s.newStore(fs)

			uploadedFileUpload, err := store.StoreFileUpload(
				context.Background(),
				businessFileUpload,
			)

			s.Require().Nil(uploadedFileUpload)
			s.Require().ErrorIs(err, expectedErr)
		})
	})
}
