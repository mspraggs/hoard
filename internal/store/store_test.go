package store_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/store"
	"github.com/mspraggs/hoard/internal/store/mocks"
	fsmodels "github.com/mspraggs/hoard/internal/store/models"
)

type FilestoreTestSuite struct {
	suite.Suite
	controller    *gomock.Controller
	mockEncKeyGen *mocks.MockEncryptionKeyGenerator
	mockClient    *mocks.MockClient
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FilestoreTestSuite))
}

func (s *FilestoreTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockEncKeyGen = mocks.NewMockEncryptionKeyGenerator(s.controller)
	s.mockClient = mocks.NewMockClient(s.controller)
}

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

func (s *FilestoreTestSuite) newStore(fs fs.FS) *store.Store {

	return store.New(
		s.mockClient, fs, models.ChecksumAlgorithmSHA256, s.mockEncKeyGen,
		models.StorageClassStandard,
	)
}

func newMemFS(files map[string][]byte) (*memfs.FS, error) {
	fs := memfs.New()

	for path, body := range files {
		err := fs.MkdirAll(filepath.Dir(path), os.FileMode(0))
		if err != nil {
			return nil, err
		}
		err = fs.WriteFile(path, body, os.FileMode(0))
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

type fileUploadMatcher struct {
	expected *fsmodels.FileUpload
}

func newFileUploadMatcher(expected *fsmodels.FileUpload) *fileUploadMatcher {
	return &fileUploadMatcher{expected}
}

func (m *fileUploadMatcher) Matches(actual interface{}) bool {
	actualUpload, ok := actual.(*fsmodels.FileUpload)
	if !ok {
		return false
	}

	actualBody, err := io.ReadAll(actualUpload.Body)
	if err != nil {
		return false
	}

	expectedBody, err := io.ReadAll(m.expected.Body)
	if err != nil {
		return false
	}

	if bytes.Compare(expectedBody, actualBody) != 0 {
		return false
	}

	return cmp.Equal(
		m.expected, actualUpload,
		cmpopts.IgnoreUnexported(fsmodels.FileUpload{}),
		cmpopts.IgnoreFields(fsmodels.FileUpload{}, "Body"),
	)
}

func (m *fileUploadMatcher) String() string {
	return fmt.Sprintf("is equal to %v", m.expected)
}
