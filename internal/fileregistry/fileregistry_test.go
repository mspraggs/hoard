package fileregistry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	pkgerrors "github.com/mspraggs/hoard/internal/errors"
	"github.com/mspraggs/hoard/internal/fileregistry"
	"github.com/mspraggs/hoard/internal/fileregistry/mocks"
	"github.com/mspraggs/hoard/internal/models"
)

type FileRegistryTestSuite struct {
	suite.Suite
	controller          *gomock.Controller
	mockStore           *mocks.MockStore
	mockInTransactioner *mocks.MockInTransactioner
}

type MockClock struct {
	now time.Time
}

func TestFileRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(FileRegistryTestSuite))
}

func (s *FileRegistryTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockStore = mocks.NewMockStore(s.controller)
	s.mockInTransactioner = mocks.NewMockInTransactioner(s.controller)
}

func (s *FileRegistryTestSuite) TestRegisterFileUpload() {
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

		s.mockInTransactioner.EXPECT().
			InTransaction(context.Background(), gomock.Any()).
			DoAndReturn(s.mockInTransaction)
		s.mockStore.EXPECT().
			GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
			Return(nil, models.ChangeTypeCreate, pkgerrors.ErrNotFound)
		s.mockStore.EXPECT().
			InsertFileUpload(context.Background(), expectedRequestID, inputFileUpload).
			Return(storedFileUpload, nil)

		registry := fileregistry.New(clock, s.mockInTransactioner)

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

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(storedFileUpload, models.ChangeTypeCreate, nil)

			registry := fileregistry.New(clock, s.mockInTransactioner)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().NoError(err)
			s.Equal(storedFileUpload, registeredFileUpload)
		})
		s.Run("and returns error for conflicting change type", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(storedFileUpload, models.ChangeTypeUpdate, nil)

			registry := fileregistry.New(clock, s.mockInTransactioner)

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

		s.Run("from InTransactioner", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				Return(nil, expectedErr)

			registry := fileregistry.New(clock, s.mockInTransactioner)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
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

			registry := fileregistry.New(clock, s.mockInTransactioner)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, expectedErr)
		})
		s.Run("from insert file upload ", func() {
			clock := &MockClock{}

			s.mockInTransactioner.EXPECT().
				InTransaction(context.Background(), gomock.Any()).
				DoAndReturn(s.mockInTransaction)
			s.mockStore.EXPECT().
				GetFileUploadByChangeRequestID(context.Background(), expectedRequestID).
				Return(nil, models.ChangeTypeCreate, pkgerrors.ErrNotFound)
			s.mockStore.EXPECT().
				InsertFileUpload(context.Background(), expectedRequestID, inputFileUpload).
				Return(nil, expectedErr)

			registry := fileregistry.New(clock, s.mockInTransactioner)

			registeredFileUpload, err := registry.RegisterFileUpload(
				context.Background(),
				inputFileUpload,
			)

			s.Require().Nil(registeredFileUpload)
			s.ErrorIs(err, expectedErr)
		})
	})
}

func (s *FileRegistryTestSuite) TestMarkFileUploadUploaded() {
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

		registry := fileregistry.New(clock, s.mockInTransactioner)

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

			registry := fileregistry.New(clock, s.mockInTransactioner)

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

			registry := fileregistry.New(clock, s.mockInTransactioner)

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

			registry := fileregistry.New(clock, s.mockInTransactioner)

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

			registry := fileregistry.New(clock, s.mockInTransactioner)

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

			registry := fileregistry.New(clock, s.mockInTransactioner)

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

func (s *FileRegistryTestSuite) mockInTransaction(
	c context.Context,
	fn fileregistry.TxnFunc,
) (interface{}, error) {

	return fn(c, s.mockStore)
}
