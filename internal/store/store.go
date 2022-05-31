package store

import (
	"context"
	"fmt"
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

// StoreFileUpload loads a file, generates an encryption key from that file and
// uploads it to the file bucket using the relevant Uploader instance.
func (s *Store) StoreFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	file, err := s.fs.Open(fileUpload.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	size, err := fileSize(file)
	if err != nil {
		return nil, fmt.Errorf("unable to determine file size: %w", err)
	}

	encKey, err := s.ekg.GenerateKey(fileUpload)
	if err != nil {
		return nil, err
	}
	upload := fsmodels.NewFileUploadFromBusiness(
		fileUpload.EncryptionAlgorithm, encKey, s.csAlg, s.sc, size, fileUpload, file,
	)

	if err := s.client.Upload(ctx, upload); err != nil {
		return nil, fmt.Errorf("unable to upload file: %w", err)
	}

	return fileUpload, nil
}

func fileSize(file fs.File) (int64, error) {
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
