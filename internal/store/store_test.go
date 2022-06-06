package store_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

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

	if actualUpload.Body != nil && m.expected.Body != nil {
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
	} else if actualUpload.Body != m.expected.Body {
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
