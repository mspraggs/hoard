package filestore

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	fsmodels "github.com/mspraggs/hoard/internal/filestore/models"
	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/filestore.go -package=mocks -source=$GOFILE

type BucketSelector interface {
	SelectBucket(fileUpload *models.FileUpload) string
}

type Checksummer interface {
	Algorithm() models.ChecksumAlgorithm
	Checksum(reader io.Reader) (models.Checksum, error)
}

type EncryptionKeyGenerator interface {
	GenerateKey(fileUpload *models.FileUpload) (models.EncryptionKey, error)
}

type Uploader interface {
	Upload(
		ctx context.Context,
		file fs.File,
		cs Checksummer,
		upload *fsmodels.FileUpload,
	) error
}

type UploaderSelector interface {
	SelectUploader(file fs.File) (Uploader, error)
}

type FileStore struct {
	fs               fs.FS
	ekg              EncryptionKeyGenerator
	cs               Checksummer
	uploaderSelector UploaderSelector
}

func New(
	fs fs.FS,
	uploaderSelector UploaderSelector,
	ekg EncryptionKeyGenerator,
	cs Checksummer,
) *FileStore {

	return &FileStore{fs, ekg, cs, uploaderSelector}
}

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
	csAlg := s.cs.Algorithm()
	upload := fsmodels.NewFileUploadFromBusiness(
		fileUpload.EncryptionAlgorithm, encKey, csAlg, fileUpload,
	)

	uploader, err := s.uploaderSelector.SelectUploader(file)
	if err != nil {
		return nil, fmt.Errorf("unable to select file uploader: %w", err)
	}

	if err := uploader.Upload(ctx, file, s.cs, upload); err != nil {
		return nil, fmt.Errorf("unable to upload file: %w", err)
	}

	return fileUpload, nil
}
