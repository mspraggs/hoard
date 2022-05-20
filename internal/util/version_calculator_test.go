package util_test

import (
	"errors"
	"io/fs"
	"syscall"
	"testing"
	"time"

	"github.com/mspraggs/hoard/internal/util"
	"github.com/stretchr/testify/suite"
)

var (
	errNotFound            = errors.New("not found")
	errUnableToGetFileInfo = errors.New("unable to get file info")
)

type VersionCalculatorTestSuite struct {
	suite.Suite
}

func TestVersionCalculatorTestSuite(t *testing.T) {
	suite.Run(t, new(VersionCalculatorTestSuite))
}

func (s *VersionCalculatorTestSuite) TestCalculateVersion() {
	goodpath := "/path/to/some/file"
	badpath := "/path/to/some/other-file"
	emptyInfoPath := "/path/to/some/empty-info-file"
	invalidInfoTypePath := "/path/to/some/invalid-info-file"

	fs := fakeFS{
		goodpath: fakeFile(func() (fs.FileInfo, error) {
			return &fakeFileInfo{&syscall.Stat_t{
				Ctim: syscall.Timespec{Sec: 2, Nsec: 1},
			}}, nil
		}),
		badpath: fakeFile(func() (fs.FileInfo, error) {
			return nil, errUnableToGetFileInfo
		}),
		emptyInfoPath: fakeFile(func() (fs.FileInfo, error) {
			return &fakeFileInfo{nil}, nil
		}),
		invalidInfoTypePath: fakeFile(func() (fs.FileInfo, error) {
			return &fakeFileInfo{0}, nil
		}),
	}

	s.Run("calcluates request ID for valid file", func() {
		expectedVersion := "7Kl0BfCTCyAw1t1lMaOpbw=="
		vc := util.NewVersionCalculator(fs)

		version, err := vc.CalculateVersion(goodpath)

		s.Require().NoError(err)
		s.Equal(expectedVersion, version)
	})

	s.Run("handles and propagates error", func() {
		s.Run("from file open", func() {
			vc := util.NewVersionCalculator(fs)

			version, err := vc.CalculateVersion("non-existent-path")

			s.Empty(version)
			s.ErrorIs(err, errNotFound)
		})
		s.Run("from reading file info", func() {
			vc := util.NewVersionCalculator(fs)

			version, err := vc.CalculateVersion(badpath)

			s.Empty(version)
			s.ErrorIs(err, errUnableToGetFileInfo)
		})
		s.Run("from empty info", func() {
			vc := util.NewVersionCalculator(fs)

			version, err := vc.CalculateVersion(emptyInfoPath)

			s.Empty(version)
			s.ErrorContains(err, "unable to get")
		})
		s.Run("from unsupported info type", func() {
			vc := util.NewVersionCalculator(fs)

			version, err := vc.CalculateVersion(invalidInfoTypePath)

			s.Empty(version)
			s.ErrorContains(err, "type not supported")
		})
	})
}

type fakeFS map[string]fs.File

func (fs fakeFS) Open(path string) (fs.File, error) {
	if f, ok := fs[path]; ok {
		return f, nil
	}
	return nil, errNotFound
}

type fakeFileInfo struct {
	sys any
}

func (fi *fakeFileInfo) Name() string       { return "" }
func (fi *fakeFileInfo) Size() int64        { return 0 }
func (fi *fakeFileInfo) Mode() fs.FileMode  { return fs.FileMode(0) }
func (fi *fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *fakeFileInfo) IsDir() bool        { return false }
func (fi *fakeFileInfo) Sys() any           { return fi.sys }

type fakeFile func() (fs.FileInfo, error)

func (f fakeFile) Read(bs []byte) (int, error) { return 0, nil }
func (f fakeFile) Close() error                { return nil }
func (f fakeFile) Stat() (fs.FileInfo, error) {
	return f()
}
