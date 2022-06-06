package store_test

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/mspraggs/hoard/internal/store"
	"github.com/mspraggs/hoard/internal/store/models"
)

func (s *AWSClientTestSuite) TestDelete() {
	s.Run("deletes file and returns no error", func() {
		upload := &models.FileUpload{
			Key: "foo",
		}
		deleteObjectInput := &s3.DeleteObjectInput{
			Key:    &upload.Key,
			Bucket: &upload.Bucket,
		}

		s.mockClient.EXPECT().
			DeleteObject(context.Background(), deleteObjectInput).
			Return(nil, nil)

		client := store.NewAWSClient(0, s.mockClient)

		err := client.Delete(context.Background(), upload)

		s.NoError(err)
	})

	s.Run("wraps and returns error", func() {
		expectedErr := errors.New("oh no")

		s.Run("from backend client", func() {
			upload := &models.FileUpload{
				Key: "foo",
			}
			deleteObjectInput := &s3.DeleteObjectInput{
				Key:    &upload.Key,
				Bucket: &upload.Bucket,
			}

			s.mockClient.EXPECT().
				DeleteObject(context.Background(), deleteObjectInput).
				Return(nil, expectedErr)

			client := store.NewAWSClient(0, s.mockClient)

			err := client.Delete(context.Background(), upload)

			s.ErrorIs(err, expectedErr)
		})
	})
}
