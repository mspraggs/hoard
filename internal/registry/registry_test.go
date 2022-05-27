package registry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	pkgerrors "github.com/mspraggs/hoard/internal/errors"
	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/registry"
	"github.com/mspraggs/hoard/internal/registry/mocks"
)

type RegistryTestSuite struct {
	suite.Suite
	controller          *gomock.Controller
	mockStore           *mocks.MockStore
	mockInTransactioner *mocks.MockInTransactioner
	mockRequestIDMaker  *mocks.MockRequestIDMaker
}

type MockClock struct {
	now time.Time
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

func (s *RegistryTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockStore = mocks.NewMockStore(s.controller)
	s.mockInTransactioner = mocks.NewMockInTransactioner(s.controller)
	s.mockRequestIDMaker = mocks.NewMockRequestIDMaker(s.controller)
}

func (s *RegistryTestSuite) TestRegisterFileUpload() {
	expectedRequestID := "yPi3JHKKbWaEEG5eZOlM6BHJll0Z3UTdBzz4bPQ7wjg="
	inputFileUpload := &models.FileUpload{
		ID:        "foo",
		LocalPath: "bar",
		Version:   "baz",
	}
	storedFileUpload := &models.FileUpload{
		ID:                 "foo",
		LocalPath:          "bar",
		Version:            "baz",
		CreatedAtTimestamp: time.Unix(1, 0),
	}

	s.Run("inserts file upload into store", func() {
		clock := &MockClock{}

		s.mockRequestIDMaker.EXPECT().
			MakeRequestID(inputFileUpload).Return(expectedRequestID, nil)
		s.mockInTransactioner.EXPECT().
			InTransaction(context.Background(), gomock.Any()).
			DoAndReturn(s.mockInTransaction)
		s.mockStore.EXPECT().
			GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
			Return(nil, models.ChangeTypeCreate, pkgerrors.ErrNotFound)
		s.mockStore.EXPECT().
			InsertFileUpload(context.Background(), expectedRequestID, inputFileUpload).
			Return(storedFileUpload, nil)

		registry := registry.New(clock, s.mockInTransactioner, s.mockRequestIDMaker)

		registeredFileUpload, err := registry.RegisterFileUpload(
			context.Background(),
			inputFileUpload,
		)

		s.Require().NoError(err)
		s.Equal(storedFileUpload, registeredFileUpload)
	})

	s.Run("handles existing file upload for matching change request", func() {
		s.Run("and returns existing file upload for matching change type", func() {
			clock := &MockClock{}

			s.mockRequestIDMaker.EXPECT().
				MakeRequestID(inputFileUpload).Return(expectedRequestID, nil)
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(storedFileUpload, models.ChangeTypeCreate, nil)

			registry := registry.New(clock, s.mockInTransactioner, s.mockRequestIDMaker)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().NoError(err)
			s.Equal(storedFileUpload, registeredFileUpload)
		})
		s.Run("and returns error for conflicting change type", func() {
			clock := &MockClock{}

			s.mockRequestIDMaker.EXPECT().
				MakeRequestID(inputFileUpload).Return(expectedRequestID, nil)
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(storedFileUpload, models.ChangeTypeUpdate, nil)

			registry := registry.New(clock, s.mockInTransactioner, s.mockRequestIDMaker)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, pkgerrors.ErrInvalidRequestID)
		})
	})

	s.Run("forwards error", func() {
		expectedErr := errors.New("oh no")

		s.Run("from RequestIDMaker", func() {
			clock := &MockClock{}

			s.mockRequestIDMaker.EXPECT().
				MakeRequestID(inputFileUpload).Return("", expectedErr)

			registry := registry.New(clock, s.mockInTransactioner, s.mockRequestIDMaker)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from InTransactioner", func() {
			clock := &MockClock{}

			s.mockRequestIDMaker.EXPECT().
				MakeRequestID(inputFileUpload).Return(expectedRequestID, nil)
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				Return(nil, expectedErr)

			registry := registry.New(clock, s.mockInTransactioner, s.mockRequestIDMaker)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from get file upload for change request ID", func() {
			clock := &MockClock{}

			s.mockRequestIDMaker.EXPECT().
				MakeRequestID(inputFileUpload).Return(expectedRequestID, nil)
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(nil, models.ChangeType(0), expectedErr)

			registry := registry.New(clock, s.mockInTransactioner, s.mockRequestIDMaker)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from insert file upload ", func() {
			clock := &MockClock{}

			s.mockRequestIDMaker.EXPECT().
				MakeRequestID(inputFileUpload).Return(expectedRequestID, nil)
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(nil, models.ChangeTypeCreate, pkgerrors.ErrNotFound)
			s.mockStore.EXPECT().
				InsertFileUpload(context.Background(), expectedRequestID, inputFileUpload).
				Return(nil, expectedErr)

			registry := registry.New(clock, s.mockInTransactioner, s.mockRequestIDMaker)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, expectedErr)
		})
	})
}

func (s *RegistryTestSuite) TestGetUploadedFileUpload() {
	fileUpload := &models.FileUpload{
		ID:                  "foo",
		UploadedAtTimestamp: time.Unix(1, 0),
	}

	s.Run("gets uploaded file upload", func() {
		s.mockInTransactioner.EXPECT().
			InTransaction(context.Background(), gomock.Any()).
			DoAndReturn(s.mockInTransaction)
		s.mockStore.EXPECT().
			GetFileUploadByChangeRequestID(context.Background(), fileUpload.ID).
			Return(fileUpload, models.ChangeTypeUpdate, nil)

		registry := registry.New(nil, s.mockInTransactioner, nil)

		uploadedFileUpload, err := registry.GetUploadedFileUpload(
			context.Background(), fileUpload.ID,
		)

		s.Require().NoError(err)
		s.Equal(fileUpload, uploadedFileUpload)
	})

	s.Run("returns nil", func() {
		s.Run("when file upload history row not found", func() {
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), fileUpload.ID).
				Return(nil, models.ChangeType(0), pkgerrors.ErrNotFound)

			registry := registry.New(nil, s.mockInTransactioner, nil)

			uploadedFileUpload, err := registry.GetUploadedFileUpload(
				context.Background(), fileUpload.ID,
			)

			s.Require().NoError(err)
			s.Nil(uploadedFileUpload)
		})
		s.Run("for file upload with invalid change type", func() {
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), fileUpload.ID).
				Return(fileUpload, models.ChangeTypeCreate, nil)

			registry := registry.New(nil, s.mockInTransactioner, nil)

			uploadedFileUpload, err := registry.GetUploadedFileUpload(
				context.Background(), fileUpload.ID,
			)

			s.Require().NoError(err)
			s.Nil(uploadedFileUpload)
		})
		s.Run("for file upload with zero upload timestamp", func() {
			nonUploadedFileUpload := &models.FileUpload{
				ID: fileUpload.ID,
			}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), fileUpload.ID).
				Return(nonUploadedFileUpload, models.ChangeTypeUpdate, nil)

			registry := registry.New(nil, s.mockInTransactioner, nil)

			uploadedFileUpload, err := registry.GetUploadedFileUpload(
				context.Background(), fileUpload.ID,
			)

			s.Require().NoError(err)
			s.Nil(uploadedFileUpload)
		})
	})

	s.Run("forwards error", func() {
		expectedErr := errors.New("oh no")

		s.Run("from InTransactioner", func() {
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				Return(nil, expectedErr)

			registry := registry.New(nil, s.mockInTransactioner, nil)

			markedUploadedFileUpload, err := registry.MarkFileUploadUploaded(
				context.Background(),
				fileUpload,
			)

			s.Require().Nil(markedUploadedFileUpload)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from get file upload for change request ID", func() {
			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), fileUpload.ID).
				Return(nil, models.ChangeType(0), expectedErr)

			registry := registry.New(nil, s.mockInTransactioner, nil)

			uploadedFileUpload, err := registry.GetUploadedFileUpload(
				context.Background(), fileUpload.ID,
			)

			s.Nil(uploadedFileUpload)
			s.ErrorIs(err, expectedErr)
		})
	})
}

func (s *RegistryTestSuite) TestMarkFileUploadUploaded() {
	inputFileUpload := &models.FileUpload{
		ID: "foo",
	}
	updatedFileUpload := &models.FileUpload{
		ID:                  "foo",
		UploadedAtTimestamp: time.Unix(1, 0),
	}
	expectedRequestID := inputFileUpload.ID

	s.Run("updates file upload in store", func() {
		clock := &MockClock{}

		s.mockInTransactioner.EXPECT().
			InTransaction(context.Background(), gomock.Any()).
			DoAndReturn(s.mockInTransaction)
		s.mockStore.EXPECT().
			GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
			Return(nil, models.ChangeTypeUpdate, pkgerrors.ErrNotFound)
		s.mockStore.EXPECT().
			UpdateFileUpload(
				context.Background(), inputFileUpload.ID, inputFileUpload,
			).
			Return(updatedFileUpload, nil)

		registry := registry.New(clock, s.mockInTransactioner, nil)

		markedUploadedFileUpload, err := registry.MarkFileUploadUploaded(
			context.Background(),
			inputFileUpload,
		)

		s.Require().NoError(err)
		s.Equal(updatedFileUpload, markedUploadedFileUpload)
	})

	s.Run("handles existing file upload for matching change request", func() {
		s.Run("and returns existing file upload for matching change type", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(updatedFileUpload, models.ChangeTypeUpdate, nil)

			registry := registry.New(clock, s.mockInTransactioner, nil)

			markedUploadedFileUpload, err := registry.MarkFileUploadUploaded(
				context.Background(),
				inputFileUpload,
			)

			s.Require().NoError(err)
			s.Equal(updatedFileUpload, markedUploadedFileUpload)
		})
		s.Run("and returns error for conflicting change type", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(updatedFileUpload, models.ChangeTypeCreate, nil)

			registry := registry.New(clock, s.mockInTransactioner, nil)

			markedUploadedFileUpload, err := registry.MarkFileUploadUploaded(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(markedUploadedFileUpload)
			s.ErrorIs(err, pkgerrors.ErrInvalidRequestID)
		})
	})

	s.Run("forwards error", func() {
		expectedErr := errors.New("oh no")

		s.Run("from InTransactioner", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				Return(nil, expectedErr)

			registry := registry.New(clock, s.mockInTransactioner, nil)

			markedUploadedFileUpload, err := registry.MarkFileUploadUploaded(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(markedUploadedFileUpload)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from get file upload for change request ID", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(nil, models.ChangeType(0), expectedErr)

			registry := registry.New(clock, s.mockInTransactioner, nil)

			markedUploadedFileUpload, err := registry.MarkFileUploadUploaded(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(markedUploadedFileUpload)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from update file upload ", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(nil, models.ChangeTypeUpdate, pkgerrors.ErrNotFound)
			s.mockStore.EXPECT().
				UpdateFileUpload(context.Background(), expectedRequestID, inputFileUpload).
				Return(nil, expectedErr)

			registry := registry.New(clock, s.mockInTransactioner, nil)

			registeredFileUpload, err := registry.MarkFileUploadUploaded(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, expectedErr)
		})
	})
}

func (c *MockClock) Now() time.Time {
	c.now = c.now.Add(time.Second)
	return c.now
}

func (s *RegistryTestSuite) mockInTransaction(
	c context.Context,
	fn registry.TxnFunc,
) (interface{}, error) {

	return fn(c, s.mockStore)
}
