package processor_test

import (
	"context"
	"errors"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/processor"
)

func (s *ProcessorTestSuite) TestDeleteFileUpload() {
	id := "foo"
	fileUpload := &models.FileUpload{
		ID: id,
	}

	s.Run("deletes file upload and returns nil", func() {
		s.mockStore.EXPECT().
			EraseFileUpload(context.Background(), fileUpload).Return(nil)
		s.mockRegistry.EXPECT().
			MarkFileUploadDeleted(context.Background(), fileUpload).Return(nil)

		processor := processor.New(s.mockStore, s.mockRegistry)

		err := processor.DeleteFileUpload(context.Background(), fileUpload)

		s.Require().NoError(err)
	})

	s.Run("handles and forwards error", func() {
		expectedErr := errors.New("oh no")

		s.Run("from store", func() {
			s.mockStore.EXPECT().
				EraseFileUpload(context.Background(), fileUpload).Return(expectedErr)

			processor := processor.New(s.mockStore, nil)

			err := processor.DeleteFileUpload(context.Background(), fileUpload)

			s.ErrorIs(err, expectedErr)
		})
		s.Run("from registry", func() {
			s.mockStore.EXPECT().
				EraseFileUpload(context.Background(), fileUpload).Return(nil)
			s.mockRegistry.EXPECT().
				MarkFileUploadDeleted(context.Background(), fileUpload).Return(expectedErr)

			processor := processor.New(s.mockStore, s.mockRegistry)

			err := processor.DeleteFileUpload(context.Background(), fileUpload)

			s.ErrorIs(err, expectedErr)
		})
	})
}
