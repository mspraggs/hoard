package uploader_test

import (
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/filestore/uploader"
	"github.com/mspraggs/hoard/internal/filestore/uploader/mocks"
)

type UploaderSelectorTestSuite struct {
	suite.Suite
	controller        *gomock.Controller
	mockSmallUploader *mocks.MockUploader
	mockLargeUploader *mocks.MockUploader
}

func TestUploaderSelectorTestSuite(t *testing.T) {
	suite.Run(t, new(UploaderSelectorTestSuite))
}

func (s *UploaderSelectorTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockSmallUploader = mocks.NewMockUploader(s.controller)
	s.mockLargeUploader = mocks.NewMockUploader(s.controller)
}

type fakeFile struct {
	info *fakeFileInfo
	err  error
}

func (f *fakeFile) Read(bs []byte) (int, error) {
	return 0, nil
}

func (f *fakeFile) Close() error {
	return nil
}

func (f *fakeFile) Stat() (fs.FileInfo, error) {
	return f.info, f.err
}

type fakeFileInfo struct {
	size int64
}

func (fi *fakeFileInfo) Size() int64        { return fi.size }
func (fi *fakeFileInfo) Name() string       { return "" }
func (fi *fakeFileInfo) Mode() fs.FileMode  { return fs.FileMode(0) }
func (fi *fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *fakeFileInfo) IsDir() bool        { return false }
func (fi *fakeFileInfo) Sys() any           { return nil }

func (s *UploaderSelectorTestSuite) TestSelectUploader() {
	s.Run("selects and returns file uploader", func() {
		threshold := int64(5)
		type testCase struct {
			name             string
			size             int64
			expectedUploader uploader.Uploader
		}

		cases := []testCase{
			{
				name:             "when size is below threshold",
				size:             threshold - 1,
				expectedUploader: s.mockSmallUploader,
			},
			{
				name:             "when size is at threshold",
				size:             threshold,
				expectedUploader: s.mockSmallUploader,
			},
			{
				name:             "when size is above threshold",
				size:             threshold + 1,
				expectedUploader: s.mockLargeUploader,
			},
		}

		for _, c := range cases {
			s.Run(c.name, func() {
				fakeFile := &fakeFile{&fakeFileInfo{c.size}, nil}

				selector := uploader.NewUploaderSelector(
					s.mockSmallUploader, s.mockLargeUploader, threshold,
				)

				uploader, err := selector.SelectUploader(fakeFile)

				s.Require().NoError(err)
				s.Equal(c.expectedUploader, uploader)
			})
		}
	})

	s.Run("forwards error from file info lookup", func() {
		expectedErr := errors.New("oh no")
		fakeFile := &fakeFile{nil, expectedErr}

		selector := uploader.NewUploaderSelector(
			s.mockSmallUploader, s.mockLargeUploader, int64(5),
		)

		uploader, err := selector.SelectUploader(fakeFile)

		s.Require().Nil(uploader)
		s.ErrorIs(err, expectedErr)
	})
}
