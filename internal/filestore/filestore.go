package filestore

import (
	"context"
	"fmt"
	"io/fs"

	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/filestore.go -package=mocks -source=$GOFILE

type BucketSelector interface {
	SelectBucket(fileUpload *models.FileUpload) string
}

// EncryptionKeyGenerator defines the interface required to generate an
// encryption key for the provided file upload.
type EncryptionKeyGenerator interface {
	GenerateKey(fileUpload *models.FileUpload) (models.EncryptionKey, error)
}

// Uploader defines the interface required to upload a file upload.
type Uploader interface {
	Upload(ctx context.Context, upload *fsmodels.FileUpload) error
}

// UploaderConstructor defines how an Uploader instance should be constructed
// from a file object.
type UploaderConstructor func(file fs.File) (Uploader, error)

// FileStore encapsulates the logic required to store a file in a storage
// bucket.
type FileStore struct {
	fs                  fs.FS
	ekg                 EncryptionKeyGenerator
	csAlg               models.ChecksumAlgorithm
	uploaderConstructor UploaderConstructor
}

// New instantiates a new file store with provided filesystem, uploader
// selector, checksum algorithm and encryption key generator.
func New(
	fs fs.FS,
	uploaderConstructor UploaderConstructor,
	csAlg models.ChecksumAlgorithm,
	ekg EncryptionKeyGenerator,
) *FileStore {

	return &FileStore{fs, ekg, csAlg, uploaderConstructor}
}

// StoreFileUpload loads a file, generates an encryption key from that file and
// uploads it to the file bucket using the relevant Uploader instance.
func (s *FileStore) StoreFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	file, err := s.fs.Open(fileUpload.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	encKey, err := s.ekg.GenerateKey(fileUpload)
	if err != nil {
		return nil, err
	}
	upload := fsmodels.NewFileUploadFromBusiness(
		fileUpload.EncryptionAlgorithm, encKey, s.csAlg, fileUpload, file,
	)

	uploader, err := s.uploaderConstructor(file)
	if err != nil {
		return nil, fmt.Errorf("unable to select file uploader: %w", err)
	}

	if err := uploader.Upload(ctx, upload); err != nil {
		return nil, fmt.Errorf("unable to upload file: %w", err)
	}

	return fileUpload, nil
}
