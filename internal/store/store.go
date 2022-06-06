package store

import (
	"context"
	"io/fs"

	"github.com/mspraggs/hoard/internal/models"
	fsmodels "github.com/mspraggs/hoard/internal/store/models"
)

//go:generate mockgen -destination=./mocks/store.go -package=mocks -source=$GOFILE

// EncryptionKeyGenerator defines the interface required to generate an
// encryption key for the provided file upload.
type EncryptionKeyGenerator interface {
	GenerateKey(fileUpload *models.FileUpload) (models.EncryptionKey, error)
}

// Client defines the interface required to interact with the storage backend.
type Client interface {
	Upload(ctx context.Context, upload *fsmodels.FileUpload) error
	Delete(ctx context.Context, upload *fsmodels.FileUpload) error
}

// Store encapsulates the logic required to store a file in a storage
// bucket.
type Store struct {
	client Client
	fs     fs.FS
	ekg    EncryptionKeyGenerator
	csAlg  models.ChecksumAlgorithm
	sc     models.StorageClass
}

// New instantiates a new file store with provided filesystem, uploader
// selector, checksum algorithm and encryption key generator.
func New(
	client Client,
	fs fs.FS,
	csAlg models.ChecksumAlgorithm,
	ekg EncryptionKeyGenerator,
	sc models.StorageClass,
) *Store {

	return &Store{client, fs, ekg, csAlg, sc}
}
