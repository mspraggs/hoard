package processor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/mspraggs/hoard/internal/models"
	"github.com/mspraggs/hoard/internal/util"
)

//go:generate mockgen -destination=./mocks/processor.go -package=mocks -source=$GOFILE

// Registry specifies the interface required to register and update the
// registry of uploaded files.
type Registry interface {
	RegisterFileUpload(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
	GetUploadedFileUpload(ctx context.Context, ID string) (*models.FileUpload, error)
	MarkFileUploadUploaded(
		ctx context.Context,
		fileUpload *models.FileUpload,
	) (*models.FileUpload, error)
}

// Store specifies the interface required to upload files.
type Store interface {
	StoreFileUpload(ctx context.Context, FileUpload *models.FileUpload) (*models.FileUpload, error)
}

// Processor encapsulates the logic required to register a file upload and store
// it in the file store.
type Processor struct {
	fs   Store
	freg Registry
	log  *zap.SugaredLogger
}

// New instantiates a new Processor instance with provided file store and
// registry.
func New(fs Store, freg Registry) *Processor {
	log := util.MustNewLogger()
	return &Processor{fs, freg, log}
}

// UploadFileUpload registers the provided file upload in the file registry and
// uploads the file to the file store.
func (h *Processor) UploadFileUpload(
	ctx context.Context,
	fileUpload *models.FileUpload,
) (*models.FileUpload, error) {

	h.log.Infow("Handling file upload", "path", fileUpload.LocalPath)

	createdFileUpload, err := h.freg.RegisterFileUpload(ctx, fileUpload)
	if err != nil {
		return nil, fmt.Errorf("error creating file upload: %w", err)
	}
	h.log.Infow(
		"Registered file upload",
		"id", createdFileUpload.ID,
		"path", createdFileUpload.LocalPath,
	)

	uploadedFileUpload, err := h.freg.GetUploadedFileUpload(ctx, createdFileUpload.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving uploaded file upload: %w", err)
	}
	if uploadedFileUpload != nil {
		h.log.Infow(
			"Skipping uploaded file upload",
			"id", uploadedFileUpload.ID,
			"path", uploadedFileUpload.LocalPath,
		)
		return uploadedFileUpload, nil
	}

	uploadedFileUpload, err = h.fs.StoreFileUpload(ctx, createdFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error while uploading file to file store: %w", err)
	}
	h.log.Infow(
		"Stored file upload in storage backend",
		"id", uploadedFileUpload.ID,
		"path", uploadedFileUpload.LocalPath,
	)

	uploadedAndMarkedFileUpload, err := h.freg.MarkFileUploadUploaded(ctx, uploadedFileUpload)
	if err != nil {
		return nil, fmt.Errorf("error marking file upload as uploaded: %w", err)
	}
	h.log.Infow(
		"Marked file upload as uploaded",
		"id", uploadedAndMarkedFileUpload.ID,
		"path", uploadedAndMarkedFileUpload.LocalPath,
	)

	return uploadedAndMarkedFileUpload, nil
}
