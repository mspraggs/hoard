package processor_test

import (
	"context"
	"errors"
	"time"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/processor"
)

func (s *ProcessorTestSuite) TestUploadFileUpload() {
	insertionTime := time.Now()
	updateTime := insertionTime.Add(time.Second)
	inputFileUpload := &models.FileUpload{
		ID: "foo",
	}
	registeredFileUpload := &models.FileUpload{
		ID:                 "foo",
		CreatedAtTimestamp: insertionTime,
	}
	uploadedFileUpload := &models.FileUpload{
		ID:                  "foo",
		CreatedAtTimestamp:  insertionTime,
		UploadedAtTimestamp: updateTime,
	}

	s.Run("registers and uploads file then marks as uploaded", func() {
		s.mockRegistry.EXPECT().
			RegisterFileUpload(context.Background(), inputFileUpload).
			Return(registeredFileUpload, nil)
		s.mockRegistry.EXPECT().
			GetUploadedFileUpload(context.Background(), registeredFileUpload.ID).
			Return(nil, nil)
		s.mockStore.EXPECT().
			StoreFileUpload(context.Background(), registeredFileUpload).
			Return(registeredFileUpload, nil)
		s.mockRegistry.EXPECT().
			MarkFileUploadUploaded(context.Background(), registeredFileUpload).
			Return(uploadedFileUpload, nil)

		processor := processor.New(s.mockStore, s.mockRegistry)

		handledFileUpload, err := processor.UploadFileUpload(context.Background(), inputFileUpload)

		s.Require().NoError(err)
		s.Equal(uploadedFileUpload, handledFileUpload)
	})

	s.Run("skips upload of registered file", func() {
		s.mockRegistry.EXPECT().
			RegisterFileUpload(context.Background(), inputFileUpload).
			Return(uploadedFileUpload, nil)
		s.mockRegistry.EXPECT().
			GetUploadedFileUpload(context.Background(), registeredFileUpload.ID).
			Return(uploadedFileUpload, nil)

		processor := processor.New(s.mockStore, s.mockRegistry)

		handledFileUpload, err := processor.UploadFileUpload(context.Background(), inputFileUpload)

		s.Require().NoError(err)
		s.Equal(uploadedFileUpload, handledFileUpload)
	})

	s.Run("returns error", func() {
		forwardedErr := errors.New("oh no")

		s.Run("forwarded from RegisterFileUpload", func() {
			s.mockRegistry.EXPECT().
				RegisterFileUpload(context.Background(), inputFileUpload).
				Return(nil, forwardedErr)

			processor := processor.New(s.mockStore, s.mockRegistry)

			handledFileUpload, err := processor.UploadFileUpload(context.Background(), inputFileUpload)

			s.Require().Nil(handledFileUpload)
			s.ErrorIs(err, forwardedErr)
		})

		s.Run("forwarded from GetUploadedFileUpload", func() {
			s.mockRegistry.EXPECT().
				RegisterFileUpload(context.Background(), inputFileUpload).
				Return(registeredFileUpload, nil)
			s.mockRegistry.EXPECT().
				GetUploadedFileUpload(context.Background(), registeredFileUpload.ID).
				Return(nil, forwardedErr)

			processor := processor.New(s.mockStore, s.mockRegistry)

			handledFileUpload, err := processor.UploadFileUpload(context.Background(), inputFileUpload)

			s.Require().Nil(handledFileUpload)
			s.ErrorIs(err, forwardedErr)
		})

		s.Run("forwarded from StoreFileUpload", func() {
			s.mockRegistry.EXPECT().
				RegisterFileUpload(context.Background(), inputFileUpload).
				Return(registeredFileUpload, nil)
			s.mockRegistry.EXPECT().
				GetUploadedFileUpload(context.Background(), registeredFileUpload.ID).
				Return(nil, nil)
			s.mockStore.EXPECT().
				StoreFileUpload(context.Background(), registeredFileUpload).
				Return(nil, forwardedErr)

			processor := processor.New(s.mockStore, s.mockRegistry)

			handledFileUpload, err := processor.UploadFileUpload(context.Background(), inputFileUpload)

			s.Require().Nil(handledFileUpload)
			s.ErrorIs(err, forwardedErr)
		})

		s.Run("forwarded from MarkFileUploadUploaded", func() {
			s.mockRegistry.EXPECT().
				RegisterFileUpload(context.Background(), inputFileUpload).
				Return(registeredFileUpload, nil)
			s.mockRegistry.EXPECT().
				GetUploadedFileUpload(context.Background(), registeredFileUpload.ID).
				Return(nil, nil)
			s.mockStore.EXPECT().
				StoreFileUpload(context.Background(), registeredFileUpload).
				Return(registeredFileUpload, nil)
			s.mockRegistry.EXPECT().
				MarkFileUploadUploaded(context.Background(), registeredFileUpload).
				Return(nil, forwardedErr)

			processor := processor.New(s.mockStore, s.mockRegistry)

			handledFileUpload, err := processor.UploadFileUpload(context.Background(), inputFileUpload)

			s.Require().Nil(handledFileUpload)
			s.ErrorIs(err, forwardedErr)
		})
	})
}
