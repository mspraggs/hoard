package store

import (
	"context"
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mspraggs/hoard/internal/util"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=./mocks/store.go -package=mocks -source=$GOFILE

const defaultChunkSize = 10 * 1024 * 1024

// Salter defines the interface required to generate a salt for use when
// generating and encryption key.
type Salter interface {
	Salt() []byte
}

// EncryptionKeyGenerator defines the interface required to generate an
// encryption key for the provided file upload.
type EncryptionKeyGenerator interface {
	GenerateKey(salt []byte) []byte
	String() string
}

// Client is the interface required to interact with a storage backend.
type Client interface {
	PutObject(
		ctx context.Context,
		input *s3.PutObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
	CreateMultipartUpload(
		ctx context.Context,
		input *s3.CreateMultipartUploadInput,
		optFns ...func(*s3.Options),
	) (*s3.CreateMultipartUploadOutput, error)
	UploadPart(
		ctx context.Context,
		input *s3.UploadPartInput,
		optFns ...func(*s3.Options),
	) (*s3.UploadPartOutput, error)
	CompleteMultipartUpload(
		ctx context.Context,
		input *s3.CompleteMultipartUploadInput,
		optFns ...func(*s3.Options),
	) (*s3.CompleteMultipartUploadOutput, error)
}

// Option defines the interface for configuring options on a Store instance.
type Option func(*Store)

// Store encapsulates the logic required to store a file in a storage
// bucket.
type Store struct {
	chunksize int64
	bucket    string
	log       *zap.SugaredLogger
	client    Client
	fs        fs.FS
	ekg       EncryptionKeyGenerator
	salter    Salter
	encAlg    EncryptionAlgorithm
	csAlg     ChecksumAlgorithm
	sc        StorageClass
}

// New instantiates a new file store with provided filesystem, uploader
// selector, checksum algorithm and encryption key generator.
func New(
	client Client,
	fs fs.FS,
	ekg EncryptionKeyGenerator,
	salter Salter,
	bucket string,
	opts ...Option,
) *Store {

	store := &Store{
		client:    client,
		fs:        fs,
		ekg:       ekg,
		salter:    salter,
		log:       util.MustNewLogger(),
		bucket:    bucket,
		chunksize: defaultChunkSize,
		encAlg:    EncryptionAlgorithm(types.ServerSideEncryptionAes256),
		csAlg:     types.ChecksumAlgorithmCrc32,
		sc:        types.StorageClassStandard,
	}
	for _, opt := range opts {
		opt(store)
	}

	return store
}

// WithChunkSize returns a Option that sets the chunk size on the provided
// store.
func WithChunkSize(chunksize int64) Option {
	return func(s *Store) {
		s.chunksize = chunksize
	}
}

// WithChecksumAlgorithm returns a Option that sets the checksum algorithm
// on the provided store.
func WithChecksumAlgorithm(algorithm ChecksumAlgorithm) Option {
	return func(s *Store) {
		s.csAlg = algorithm
	}
}

// WithEncryptionAlgorithm returns a Option that sets the checksum algorithm
// on the provided store.
func WithEncryptionAlgorithm(algorithm EncryptionAlgorithm) Option {
	return func(s *Store) {
		s.encAlg = algorithm
	}
}

// WithStorageClass returns a Option that sets the checksum algorithm
// on the provided store.
func WithStorageClass(storageClass StorageClass) Option {
	return func(s *Store) {
		s.sc = storageClass
	}
}
