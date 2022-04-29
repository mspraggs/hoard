package filestore_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/filestore"
	"github.com/mspraggs/hoard/internal/filestore/mocks"
	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
	"github.com/mspraggs/hoard/internal/models"
)

var errFileNotFound = errors.New("file not found")

type FilestoreTestSuite struct {
	suite.Suite
	controller           *gomock.Controller
	mockChecksummer      *mocks.MockChecksummer
	mockEncKeyGen        *mocks.MockEncryptionKeyGenerator
	mockUploaderSelector *mocks.MockUploaderSelector
}

type fakeFS map[string]*fakeFile

func (fs fakeFS) Open(path string) (fs.File, error) {
	if f, ok := fs[path]; ok {
		return f, nil
	}
	return nil, errFileNotFound
}

type fakeFile struct {
}

func (f *fakeFile) Read(bs []byte) (int, error) {
	return 0, nil
}

func (f *fakeFile) Close() error {
	return nil
}

func (f *fakeFile) Stat() (fs.FileInfo, error) {
	return nil, nil
}

func TestFileStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FilestoreTestSuite))
}

func (s *FilestoreTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockChecksummer = mocks.NewMockChecksummer(s.controller)
	s.mockEncKeyGen = mocks.NewMockEncryptionKeyGenerator(s.controller)
	s.mockUploaderSelector = mocks.NewMockUploaderSelector(s.controller)
}

func (s *FilestoreTestSuite) TestStoreFileUpload() {
	s.Run("uploads file with key and returns upload", func() {
		fileID := "some-file"
		path := "/path/to/file"
		encKey := models.EncryptionKey("meh")

		businessFileUpload := &models.FileUpload{
			ID:        fileID,
			LocalPath: path,
		}
		fsFileUpload := &fsmodels.FileUpload{
			Key:                 fileID,
			EncryptionKey:       fsmodels.EncryptionKey("meh"),
			EncryptionAlgorithm: types.ServerSideEncryptionAes256,
			ChecksumAlgorithm:   types.ChecksumAlgorithmSha256,
		}

		fakeFile := &fakeFile{}
		fakeFileSystem := fakeFS{path: fakeFile}

		s.mockEncKeyGen.EXPECT().
			Algorithm().Return(models.EncryptionAlgorithmAES256)
		s.mockEncKeyGen.EXPECT().
			GenerateKey(businessFileUpload).Return(encKey)
		s.mockChecksummer.EXPECT().
			Algorithm().Return(models.ChecksumAlgorithmSHA256)

		mockUploader := mocks.NewMockUploader(s.controller)
		mockUploader.EXPECT().
			Upload(context.Background(), fakeFile, s.mockChecksummer, fsFileUpload).
			Return(nil)

		s.mockUploaderSelector.EXPECT().SelectUploader(fakeFile).Return(mockUploader, nil)

		store := s.newStore(fakeFileSystem)

		uploadedFileUpload, err := store.StoreFileUpload(context.Background(), businessFileUpload)

		s.Require().NoError(err)
		s.Require().Equal(businessFileUpload, uploadedFileUpload)
	})

	s.Run("wraps and returns error", func() {
		s.Run("from filesystem", func() {
			expectedErr := errFileNotFound
			fileID := "some-file"
			path := "/path/to/file"

			businessFileUpload := &models.FileUpload{
				ID:        fileID,
				LocalPath: path,
			}

			store := s.newStore(&fakeFS{})

			uploadedFileUpload, err := store.StoreFileUpload(
				context.Background(),
				businessFileUpload,
			)

			s.Require().Nil(uploadedFileUpload)
			s.Require().ErrorIs(err, expectedErr)
		})
		s.Run("from upload selector", func() {
			expectedErr := errors.New("oh no")
			fileID := "some-file"
			path := "/path/to/file"
			encKey := models.EncryptionKey("meh")

			businessFileUpload := &models.FileUpload{
				ID:        fileID,
				LocalPath: path,
			}

			fakeFile := &fakeFile{}
			fakeFileSystem := fakeFS{path: fakeFile}

			s.mockEncKeyGen.EXPECT().
				Algorithm().Return(models.EncryptionAlgorithmAES256)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(businessFileUpload).Return(encKey)
			s.mockChecksummer.EXPECT().
				Algorithm().Return(models.ChecksumAlgorithmSHA256)

			s.mockUploaderSelector.EXPECT().SelectUploader(fakeFile).Return(nil, expectedErr)

			store := s.newStore(fakeFileSystem)

			uploadedFileUpload, err := store.StoreFileUpload(
				context.Background(),
				businessFileUpload,
			)

			s.Require().Nil(uploadedFileUpload)
			s.Require().ErrorIs(err, expectedErr)
		})
		s.Run("from uploader", func() {
			expectedErr := errors.New("oh no")
			fileID := "some-file"
			path := "/path/to/file"
			encKey := models.EncryptionKey("meh")

			businessFileUpload := &models.FileUpload{
				ID:        fileID,
				LocalPath: path,
			}
			fsFileUpload := &fsmodels.FileUpload{
				Key:                 fileID,
				EncryptionKey:       fsmodels.EncryptionKey("meh"),
				EncryptionAlgorithm: types.ServerSideEncryptionAes256,
				ChecksumAlgorithm:   types.ChecksumAlgorithmSha256,
			}

			fakeFile := &fakeFile{}
			fakeFileSystem := fakeFS{path: fakeFile}

			s.mockEncKeyGen.EXPECT().
				Algorithm().Return(models.EncryptionAlgorithmAES256)
			s.mockEncKeyGen.EXPECT().
				GenerateKey(businessFileUpload).Return(encKey)
			s.mockChecksummer.EXPECT().
				Algorithm().Return(models.ChecksumAlgorithmSHA256)

			mockUploader := mocks.NewMockUploader(s.controller)
			mockUploader.EXPECT().
				Upload(context.Background(), fakeFile, s.mockChecksummer, fsFileUpload).
				Return(expectedErr)

			s.mockUploaderSelector.EXPECT().SelectUploader(fakeFile).Return(mockUploader, nil)

			store := s.newStore(fakeFileSystem)

			uploadedFileUpload, err := store.StoreFileUpload(
				context.Background(),
				businessFileUpload,
			)

			s.Require().Nil(uploadedFileUpload)
			s.Require().ErrorIs(err, expectedErr)
		})
	})
}

func (s *FilestoreTestSuite) newStore(fs fs.FS) *filestore.FileStore {
	return filestore.New(
		fs,
		s.mockUploaderSelector,
		s.mockEncKeyGen,
		s.mockChecksummer,
	)
}
