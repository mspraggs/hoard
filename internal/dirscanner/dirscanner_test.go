package dirscanner_test

import (
	"context"
	"encoding/base64"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/dirscanner"
	"github.com/mspraggs/hoard/internal/dirscanner/mocks"
	"github.com/mspraggs/hoard/internal/models"
)

var errPathNotFound = errors.New("path not found")

type DirScannerTestSuite struct {
	suite.Suite
	controller            *gomock.Controller
	mockFileUploadHandler *mocks.MockFileUploadHandler
	mockVersionCalculator *mocks.MockVersionCalculator
	mockSalter            *mocks.MockSalter
}

func TestDirScannerTestSuite(t *testing.T) {
	suite.Run(t, new(DirScannerTestSuite))
}

func (s *DirScannerTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockFileUploadHandler = mocks.NewMockFileUploadHandler(s.controller)
	s.mockVersionCalculator = mocks.NewMockVersionCalculator(s.controller)
	s.mockSalter = mocks.NewMockSalter(s.controller)
}

func (s *DirScannerTestSuite) TestScan() {
	bucket := "some-bucket"
	numThreads := 2
	version := "some-version"
	salt := []byte{1, 2, 3, 4, 5}
	encryptionAlgorithm := models.EncryptionAlgorithmAES256

	paths := []string{
		"foo/bar",
		"foo/baz",
		"top-level",
		"some/deeply/nested/path",
	}
	uploads := newTestFileUploadsFromPaths(paths, bucket, encryptionAlgorithm, version, salt)

	versionCalculatorNoError := &fakeVersionCalculator{version, nil}
	salterNoError := &fakeSalter{salt, nil}

	s.Run("scans filesystem and handles uploads", func() {
		ctx := context.Background()

		fs := s.newMemFS(paths)

		s.newHandlerCallsFromUploads(ctx, uploads)

		dirScanner := dirscanner.NewBuilder().
			WithFS(fs).
			WithVersionCalculator(versionCalculatorNoError).
			WithSalter(salterNoError).
			WithBucket(bucket).
			WithNumHandlerThreads(numThreads).
			WithEncryptionAlgorithm(encryptionAlgorithm).
			AddFileUploadHandler(s.mockFileUploadHandler).
			Build()

		dirScanner.Scan(ctx)
	})

	s.Run("stops handlers upon context canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		fs := s.newMemFS(paths)

		dirScanner := dirscanner.NewBuilder().
			WithFS(fs).
			WithVersionCalculator(nil).
			WithSalter(nil).
			WithBucket(bucket).
			WithNumHandlerThreads(numThreads).
			WithEncryptionAlgorithm(encryptionAlgorithm).
			AddFileUploadHandler(s.mockFileUploadHandler).
			Build()

		err := dirScanner.Scan(ctx)

		s.Require().ErrorIs(context.Canceled, err)
	})

	s.Run("handles error", func() {
		expectedErr := errors.New("oh no")
		paths := []string{"foo"}
		uploads := newTestFileUploadsFromPaths(paths, bucket, encryptionAlgorithm, version, salt)

		fs := s.newMemFS(paths)
		versionCalculatorError := &fakeVersionCalculator{"", expectedErr}
		salterError := &fakeSalter{nil, expectedErr}

		s.Run("from filesystem", func() {
			ctx := context.Background()

			dirScanner := dirscanner.NewBuilder().
				WithFS(&fakeBadFS{expectedErr}).
				WithVersionCalculator(nil).
				WithSalter(nil).
				AddFileUploadHandler(nil).
				Build()

			dirScanner.Scan(ctx)
		})
		s.Run("from version calculator", func() {
			ctx := context.Background()

			dirScanner := dirscanner.NewBuilder().
				WithFS(fs).
				WithVersionCalculator(versionCalculatorError).
				WithSalter(nil).
				AddFileUploadHandler(nil).
				Build()

			dirScanner.Scan(ctx)
		})
		s.Run("from salter", func() {
			ctx := context.Background()

			dirScanner := dirscanner.NewBuilder().
				WithFS(fs).
				WithVersionCalculator(versionCalculatorNoError).
				WithSalter(salterError).
				AddFileUploadHandler(nil).
				Build()

			dirScanner.Scan(ctx)
		})
		s.Run("from handler", func() {
			ctx := context.Background()

			s.mockFileUploadHandler.EXPECT().
				HandleFileUpload(ctx, uploads[0]).Return(nil, expectedErr)

			dirScanner := dirscanner.NewBuilder().
				WithFS(fs).
				WithVersionCalculator(versionCalculatorNoError).
				WithSalter(salterNoError).
				WithBucket(bucket).
				WithEncryptionAlgorithm(encryptionAlgorithm).
				AddFileUploadHandler(s.mockFileUploadHandler).
				Build()

			dirScanner.Scan(ctx)
		})
	})
}

func (s *DirScannerTestSuite) newMemFS(paths []string) *memfs.FS {
	memFS := memfs.New()

	for _, path := range paths {
		err := memFS.MkdirAll(filepath.Dir(path), fs.FileMode(0))
		s.Require().NoError(err)
		err = memFS.WriteFile(path, []byte{}, fs.FileMode(0))
		s.Require().NoError(err)
	}

	return memFS
}

func (s *DirScannerTestSuite) newHandlerCallsFromUploads(
	ctx context.Context,
	uploads []*models.FileUpload,
) []*gomock.Call {

	calls := make([]*gomock.Call, len(uploads))

	for i, upload := range uploads {
		calls[i] = s.mockFileUploadHandler.EXPECT().
			HandleFileUpload(ctx, upload).Return(upload, nil)
	}

	return calls
}

func newTestFileUploadsFromPaths(
	paths []string,
	bucket string,
	encAlg models.EncryptionAlgorithm,
	version string,
	salt []byte,
) []*models.FileUpload {

	uploads := make([]*models.FileUpload, len(paths))

	for i, path := range paths {
		upload := &models.FileUpload{
			Bucket:              bucket,
			LocalPath:           path,
			Version:             version,
			Salt:                base64.RawStdEncoding.EncodeToString(salt),
			EncryptionAlgorithm: encAlg,
		}
		uploads[i] = upload
	}

	return uploads
}

type fakeBadFS struct {
	err error
}

func (fs *fakeBadFS) Open(path string) (fs.File, error) {
	return nil, nil
}

func (fs *fakeBadFS) Stat(path string) (fs.FileInfo, error) {
	return nil, fs.err
}

type fakeSalter struct {
	salt []byte
	err  error
}

func (s *fakeSalter) Salt(path string) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.salt, nil
}

type fakeVersionCalculator struct {
	version string
	err     error
}

func (c fakeVersionCalculator) CalculateVersion(path string) (string, error) {
	if c.err != nil {
		return "", c.err
	}
	return c.version, nil
}
