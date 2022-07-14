package store_test

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/psanford/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/mspraggs/hoard/internal/processor"
	"github.com/mspraggs/hoard/internal/store"
	"github.com/mspraggs/hoard/internal/store/mocks"
)

type contextKey string

type StoreTestSuite struct {
	suite.Suite
	controller    *gomock.Controller
	mockEncKeyGen *mocks.MockEncryptionKeyGenerator
	mockSalter    *mocks.MockSalter
	mockClient    *mocks.MockClient
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.mockEncKeyGen = mocks.NewMockEncryptionKeyGenerator(s.controller)
	s.mockSalter = mocks.NewMockSalter(s.controller)
	s.mockClient = mocks.NewMockClient(s.controller)
}

func (s *StoreTestSuite) newStore(fs fs.FS, bucket string) *store.Store {
	return store.New(s.mockClient, fs, s.mockEncKeyGen, s.mockSalter, bucket)
}

func (s *StoreTestSuite) makeDoUploadPart(
	expectedBytes []byte,
	maxChunkSize int64,
	eTag string,
) func(context.Context, *s3.UploadPartInput, ...func(*s3.Options)) (*s3.UploadPartOutput, error) {

	return func(
		ctx context.Context,
		upi *s3.UploadPartInput,
		optFns ...func(*s3.Options),
	) (*s3.UploadPartOutput, error) {

		bs := make([]byte, maxChunkSize)
		n, err := upi.Body.Read(bs)
		s.Require().Equal(expectedBytes, bs[:n])
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			return nil, err
		}
		return &s3.UploadPartOutput{ETag: &eTag}, nil
	}
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

func newTestPutObjectInput(
	file *processor.File,
	bucket string,
	encKey []byte,
	csAlg types.ChecksumAlgorithm,
	body io.Reader,
) *s3.PutObjectInput {

	sseKey := base64.StdEncoding.EncodeToString(encKey)
	hashedEncKey := md5.Sum(encKey)
	sseKeyMD5 := base64.StdEncoding.EncodeToString(hashedEncKey[:])
	sseAlg := string(types.ServerSideEncryptionAes256)
	return &s3.PutObjectInput{
		Key:                  &file.Key,
		Bucket:               &bucket,
		SSECustomerKey:       &sseKey,
		SSECustomerKeyMD5:    &sseKeyMD5,
		SSECustomerAlgorithm: &sseAlg,
		ChecksumAlgorithm:    csAlg,
		StorageClass:         types.StorageClassStandard,
		Body:                 body,
	}
}

func newTestUploadPartInput(
	file *processor.File,
	bucket string,
	encKey []byte,
	uploadID string,
	csAlg types.ChecksumAlgorithm,
	chunkNum int32,
	chunkSize int64,
	body io.Reader,
) *s3.UploadPartInput {

	sseKey := base64.StdEncoding.EncodeToString(encKey)
	hashedEncKey := md5.Sum(encKey)
	sseKeyMD5 := base64.StdEncoding.EncodeToString(hashedEncKey[:])
	sseAlg := string(types.ServerSideEncryptionAes256)
	return &s3.UploadPartInput{
		Key:                  &file.Key,
		UploadId:             &uploadID,
		Bucket:               &bucket,
		PartNumber:           chunkNum,
		ContentLength:        chunkSize,
		SSECustomerKey:       &sseKey,
		SSECustomerKeyMD5:    &sseKeyMD5,
		SSECustomerAlgorithm: &sseAlg,
		ChecksumAlgorithm:    csAlg,
		Body:                 body,
	}
}

type putObjectInputMatcher struct {
	expected *s3.PutObjectInput
}

func newPutObjectInputMatcher(expected *s3.PutObjectInput) *putObjectInputMatcher {
	return &putObjectInputMatcher{expected}
}

func (m *putObjectInputMatcher) Matches(actual interface{}) bool {
	actualInput, ok := actual.(*s3.PutObjectInput)
	if !ok {
		return false
	}

	actualBody, err := io.ReadAll(actualInput.Body)
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
		m.expected, actualInput,
		cmpopts.IgnoreUnexported(s3.PutObjectInput{}),
		cmpopts.IgnoreFields(s3.PutObjectInput{}, "Body"),
	)
}

func (m *putObjectInputMatcher) String() string {
	return fmt.Sprintf("is equal to %v", m.expected)
}

type uploadPartInputMatcher struct {
	expected *s3.UploadPartInput
}

func newUploadPartInputMatcher(expected *s3.UploadPartInput) *uploadPartInputMatcher {
	return &uploadPartInputMatcher{expected}
}

func (m *uploadPartInputMatcher) Matches(actual interface{}) bool {
	fmt.Println(cmp.Diff(m.expected, actual,
		cmpopts.IgnoreUnexported(s3.UploadPartInput{}),
		cmpopts.IgnoreFields(s3.UploadPartInput{}, "Body"),
	))
	actualInput, ok := actual.(*s3.UploadPartInput)
	if !ok {
		return false
	}

	return cmp.Equal(
		m.expected, actualInput,
		cmpopts.IgnoreUnexported(s3.UploadPartInput{}),
		cmpopts.IgnoreFields(s3.UploadPartInput{}, "Body"),
	)
}

func (m *uploadPartInputMatcher) String() string {
	return fmt.Sprintf("is equal to %v", m.expected)
}

type fakeFS struct {
	err error
}

func (fs *fakeFS) Open(path string) (fs.File, error) {
	return &fakeFile{fs.err}, nil
}

type fakeFile struct {
	err error
}

func (f *fakeFile) Read(bs []byte) (int, error) { return 0, nil }
func (f *fakeFile) Close() error                { return nil }
func (f *fakeFile) Stat() (fs.FileInfo, error) {
	return nil, f.err
}
