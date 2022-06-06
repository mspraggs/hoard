package store_test

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mspraggs/hoard/internal/models"
	fsmodels "github.com/mspraggs/hoard/internal/store/models"
)

func (s *FilestoreTestSuite) TestEraseFileUpload() {
	s.Run("erases file and returns nil", func() {
		fileID := "some-file"

		businessFileUpload := &models.FileUpload{
			ID: fileID,
		}
		fsFileUpload := &fsmodels.FileUpload{
			Key:                 fileID,
			EncryptionKey:       fsmodels.EncryptionKey{},
			EncryptionAlgorithm: types.ServerSideEncryptionAes256,
			ChecksumAlgorithm:   types.ChecksumAlgorithmSha256,
			StorageClass:        types.StorageClassStandard,
			Size:                0,
			Body:                nil,
		}

		s.mockClient.EXPECT().
			Delete(context.Background(), newFileUploadMatcher(fsFileUpload)).
			Return(nil)

		store := s.newStore(nil)

		err := store.EraseFileUpload(context.Background(), businessFileUpload)

		s.Require().NoError(err)
	})

	s.Run("wraps and returns error", func() {
		s.Run("from client", func() {
			expectedErr := errors.New("oh no")
			fileID := "some-file"

			businessFileUpload := &models.FileUpload{
				ID: fileID,
			}
			fsFileUpload := &fsmodels.FileUpload{
				Key:                 fileID,
				EncryptionKey:       fsmodels.EncryptionKey{},
				EncryptionAlgorithm: types.ServerSideEncryptionAes256,
				ChecksumAlgorithm:   types.ChecksumAlgorithmSha256,
				StorageClass:        types.StorageClassStandard,
				Size:                0,
				Body:                nil,
			}

			s.mockClient.EXPECT().
				Delete(context.Background(), newFileUploadMatcher(fsFileUpload)).
				Return(expectedErr)

			store := s.newStore(nil)

			err := store.EraseFileUpload(
				context.Background(),
				businessFileUpload,
			)

			s.Require().ErrorIs(err, expectedErr)
		})
	})
}
