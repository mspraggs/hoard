package dirscanner_test

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/dirscanner"
	"github.com/mspraggs/hoard/internal/dirscanner/mocks"
	"github.com/mspraggs/hoard/internal/processor"
)

var errPathNotFound = errors.New("path not found")

type DirScannerTestSuite struct {
	suite.Suite
	controller    *gomock.Controller
	mockProcessor *mocks.MockProcessor
}

func TestDirScannerTestSuite(t *testing.T) {
	suite.Run(t, new(DirScannerTestSuite))
}

func (s *DirScannerTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockProcessor = mocks.NewMockProcessor(s.controller)
}

func (s *DirScannerTestSuite) TestScan() {
	numThreads := 2

	paths := []string{
		"foo/bar",
		"foo/baz",
		"top-level",
		"some/deeply/nested/path",
	}

	s.Run("scans filesystem and handles uploads", func() {
		ctx := context.Background()

		fs := s.newMemFS(paths)

		s.newHandlerCallsFromPaths(ctx, paths)

		dirScanner := dirscanner.New(fs, []dirscanner.Processor{s.mockProcessor}, numThreads)

		dirScanner.Scan(ctx)
	})

	s.Run("stops handlers upon context canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		fs := s.newMemFS(paths)

		dirScanner := dirscanner.New(fs, []dirscanner.Processor{s.mockProcessor}, numThreads)

		err := dirScanner.Scan(ctx)

		s.Require().ErrorIs(context.Canceled, err)
	})

	s.Run("handles error", func() {
		expectedErr := errors.New("oh no")
		paths := []string{"foo"}

		fs := s.newMemFS(paths)

		s.Run("from filesystem", func() {
			ctx := context.Background()

			dirScanner := dirscanner.New(
				&fakeBadFS{expectedErr},
				[]dirscanner.Processor{s.mockProcessor},
				numThreads,
			)

			dirScanner.Scan(ctx)
		})
		s.Run("from processor", func() {
			ctx := context.Background()

			s.mockProcessor.EXPECT().
				Process(ctx, paths[0]).Return(nil, expectedErr)

			dirScanner := dirscanner.New(
				fs,
				[]dirscanner.Processor{s.mockProcessor},
				numThreads,
			)

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

func (s *DirScannerTestSuite) newHandlerCallsFromPaths(
	ctx context.Context,
	paths []string,
) []*gomock.Call {

	calls := make([]*gomock.Call, len(paths))

	for i, path := range paths {
		file := &processor.File{
			LocalPath: path,
		}
		calls[i] = s.mockProcessor.EXPECT().
			Process(ctx, path).Return(file, nil)
	}

	return calls
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
