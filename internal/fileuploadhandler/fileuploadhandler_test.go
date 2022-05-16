package fileuploadhandler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/fileuploadhandler"
	"github.com/mspraggs/hoard/internal/fileuploadhandler/mocks"
	"github.com/mspraggs/hoard/internal/models"
)

type FileUploadHandlerTestSuite struct {
	suite.Suite
	controller       *gomock.Controller
	now              time.Time
	mockFileRegistry *mocks.MockFileRegistry
	mockFileStore    *mocks.MockFileStore
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(FileUploadHandlerTestSuite))
}

func (s *FileUploadHandlerTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.now = time.Time{}
	s.mockFileRegistry = mocks.NewMockFileRegistry(s.controller)
	s.mockFileStore = mocks.NewMockFileStore(s.controller)
}

func (s *FileUploadHandlerTestSuite) TestHandleFileUpload() {
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
		s.mockFileRegistry.EXPECT().
			RegisterFileUpload(context.Background(), inputFileUpload).
			Return(registeredFileUpload, nil)
		s.mockFileStore.EXPECT().
			StoreFileUpload(context.Background(), registeredFileUpload).
			Return(registeredFileUpload, nil)
		s.mockFileRegistry.EXPECT().
			MarkFileUploadUploaded(context.Background(), registeredFileUpload).
			Return(uploadedFileUpload, nil)

		handler := fileuploadhandler.New(s.mockFileStore, s.mockFileRegistry)

		handledFileUpload, err := handler.HandleFileUpload(context.Background(), inputFileUpload)

		s.Require().NoError(err)
		s.Equal(uploadedFileUpload, handledFileUpload)
	})

	s.Run("skips upload of registered file", func() {
		s.mockFileRegistry.EXPECT().
			RegisterFileUpload(context.Background(), inputFileUpload).
			Return(uploadedFileUpload, nil)

		handler := fileuploadhandler.New(s.mockFileStore, s.mockFileRegistry)

		handledFileUpload, err := handler.HandleFileUpload(context.Background(), inputFileUpload)

		s.Require().NoError(err)
		s.Equal(uploadedFileUpload, handledFileUpload)
	})

	s.Run("returns error", func() {
		forwardedErr := errors.New("oh no")

		s.Run("forwarded from RegisterFileUpload", func() {
			s.mockFileRegistry.EXPECT().
				RegisterFileUpload(context.Background(), inputFileUpload).
				Return(nil, forwardedErr)

			handler := fileuploadhandler.New(s.mockFileStore, s.mockFileRegistry)

			handledFileUpload, err := handler.HandleFileUpload(context.Background(), inputFileUpload)

			s.Require().Nil(handledFileUpload)
			s.ErrorIs(err, forwardedErr)
		})

		s.Run("forwarded from StoreFileUpload", func() {
			s.mockFileRegistry.EXPECT().
				RegisterFileUpload(context.Background(), inputFileUpload).
				Return(registeredFileUpload, nil)
			s.mockFileStore.EXPECT().
				StoreFileUpload(context.Background(), registeredFileUpload).
				Return(nil, forwardedErr)

			handler := fileuploadhandler.New(s.mockFileStore, s.mockFileRegistry)

			handledFileUpload, err := handler.HandleFileUpload(context.Background(), inputFileUpload)

			s.Require().Nil(handledFileUpload)
			s.ErrorIs(err, forwardedErr)
		})

		s.Run("forwarded from MarkFileUploadUploaded", func() {
			s.mockFileRegistry.EXPECT().
				RegisterFileUpload(context.Background(), inputFileUpload).
				Return(registeredFileUpload, nil)
			s.mockFileStore.EXPECT().
				StoreFileUpload(context.Background(), registeredFileUpload).
				Return(registeredFileUpload, nil)
			s.mockFileRegistry.EXPECT().
				MarkFileUploadUploaded(context.Background(), registeredFileUpload).
				Return(nil, forwardedErr)

			handler := fileuploadhandler.New(s.mockFileStore, s.mockFileRegistry)

			handledFileUpload, err := handler.HandleFileUpload(context.Background(), inputFileUpload)

			s.Require().Nil(handledFileUpload)
			s.ErrorIs(err, forwardedErr)
		})
	})
}
