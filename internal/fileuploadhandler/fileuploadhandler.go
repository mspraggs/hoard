package fileuploadhandler

import (
	"context"
	"fmt"

	"github.com/mspraggs/hoard/internal/models"
)

//go:generate mockgen -destination=./mocks/fileuploadhandler.go -package=mocks -source=$GOFILE

// FileRegistry specifies the interface required to register and update the
// registry of uploaded files.
type FileRegistry interface {
	RegisterFileUpload(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
	MarkFileUploadUploaded(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
}

// FileStore specifies the interface required to upload files.
type FileStore interface {
	StoreFileUpload(ctx context.Context, FileUpload *models.FileUpload) (*models.FileUpload, error)
}

// FileUploadHandler encapsulates the logic required to register a file upload and store
// it in the file store.
type FileUploadHandler struct {
	fs   FileStore
	freg FileRegistry
}

// New instantiates a new FileUploadHandler instance with provided file store and
// registry.
func New(fs FileStore, freg FileRegistry) *FileUploadHandler {
	return &FileUploadHandler{fs, freg}
}

// HandleFileUpload registers the provided file upload in the file registry and
// uploads the file to the file store.
func (h *FileUploadHandler) HandleFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	createdFileUpload, err := h.freg.RegisterFileUpload(ctx, fileUpload)
	if err != nil {
		return nil, fmt.Errorf("error creating file upload: %w", err)
	}

	if !createdFileUpload.UploadedAtTimestamp.IsZero() {
		// TODO: Logging
		return createdFileUpload, nil
	}

	uploadedFileUpload, err := h.fs.StoreFileUpload(ctx, createdFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error while uploading file to file store: %w", err)
	}

	uploadedAndMarkedFileUpload, err := h.freg.MarkFileUploadUploaded(ctx, uploadedFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error marking file upload as uploaded: %w", err)
	}

	return uploadedAndMarkedFileUpload, nil
}
